package openshift

import (
	"os"
	"strings"
	"sync"

	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	occlient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/openqe/openqe/pkg/auth"
	"github.com/openqe/openqe/pkg/tls"
	"github.com/openqe/openqe/pkg/utils"
	configv1 "github.com/openshift/api/config/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	clientCache = make(map[string]occlient.Client)
	cacheMutex  sync.Mutex
)

// GetOrCreateOCClient gets or creates a go-client to talk with openshift api server
// It will cache the client based on the kubeconfig file used
func GetOrCreateOCClient(kubeconfig string) (occlient.Client, context.Context, logr.Logger, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	ctx := context.TODO()
	log := ctrl.LoggerFrom(ctx)

	if cachedClient, ok := clientCache[kubeconfig]; ok {
		return cachedClient, ctx, log, nil
	}

	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, ctx, log, err
	}

	scheme := runtime.NewScheme()
	// Register corev1 and configv1 types to the scheme
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, ctx, log, err
	}
	if err := configv1.Install(scheme); err != nil {
		return nil, ctx, log, err
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		return nil, ctx, log, err
	}
	if err := routev1.Install(scheme); err != nil {
		return nil, ctx, log, err
	}
	client, err := occlient.New(restConfig, occlient.Options{Scheme: scheme})
	if err != nil {
		return nil, ctx, log, err
	}

	clientCache[kubeconfig] = client
	return client, ctx, log, nil
}

// CreateNamespaceIfNotExists creates a namespace in the OpenShift cluster if it doesn't already exist
// It returns the *corev1.Namespace if all work good
func CreateNamespaceIfNotExists(kubeconfig, namespace string) (*corev1.Namespace, error) {
	client, ctx, log, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	ns := &corev1.Namespace{}
	err = client.Get(ctx, occlient.ObjectKey{Name: namespace}, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			// not found, create it.
			newNamespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			if err := client.Create(ctx, newNamespace); err != nil {
				return nil, err
			}
			log.Info("Namespace %s created. \n", namespace)
			return newNamespace, nil
		}
		return nil, err
	}
	log.Info("Namespace %s exists\n", namespace)
	return ns, nil
}

// ConfigureAdditionalCA configures additional CA in the openshift cluster, it will lead to rolling out and wait until all MachineConfigPools finish updating.
// If the trusted ca bundle has been set, it will update that ConfigMap and it will roll out.
// When this method returns, the additional CA has been set up in all nodes
func ConfigureAdditionalCA(kubeconfig, caCertFile string, verbose bool) error {
	client, ctx, log, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return err
	}
	proxy := &configv1.Proxy{}
	if err = client.Get(ctx, occlient.ObjectKey{Name: "cluster"}, proxy); err != nil {
		return err
	}
	var existing *configv1.ConfigMapNameReference
	if proxy.Spec.TrustedCA.Name != "" {
		existing = &proxy.Spec.TrustedCA
	}

	shouldRolling := false
	cm := &corev1.ConfigMap{}
	caCertData, err := os.ReadFile(caCertFile)
	if err != nil {
		return err
	}
	if existing != nil {
		if err = client.Get(ctx, occlient.ObjectKey{Name: existing.Name, Namespace: "openshift-config"}, cm); err != nil {
			return err
		}
		caCert := string(caCertData)
		existingBundle := cm.Data["ca-bundle.crt"]
		alreadyAdd, err := tls.CertInConfigMap(cm, caCert)
		if err != nil {
			return err
		}
		if !alreadyAdd {
			log.Info("Updating existing trusted CA config map %s\n", existing.Name)
			cm.Data["ca-bundle.crt"] = existingBundle + "\n" + caCert
			if err := client.Update(ctx, cm); err != nil {
				return err
			}
			shouldRolling = true
		} else {
			log.Info("The CA certificate is already in the existing trusted CA config map %s, no need to update\n", existing.Name)
			shouldRolling = false
		}
	} else {
		// create a new config map with a ca-bundle.crt key from file: caCertFile
		cmName := "openqe-trusted-ca"
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: "openshift-config",
			},
			Data: map[string]string{
				"ca-bundle.crt": string(caCertData),
			},
		}
		if err := client.Create(ctx, cm); err != nil {
			return err
		}
		proxy.Spec.TrustedCA = configv1.ConfigMapNameReference{
			Name: cm.Name,
		}
		if err := client.Update(ctx, proxy); err != nil {
			return err
		}
		shouldRolling = true
	}
	if shouldRolling {
		// wait until mcp status is updating
		utils.EventuallyDefault(
			func() (string, error) {
				args := []string{"mcp", "-o", `jsonpath={.items[*].status.conditions[?(@.type=="Updating")].status}`}
				return OC_Get(kubeconfig, verbose, args...)
			},
			func(v string) bool {
				return utils.AllTrue(v)
			},
		)
		// wait until mcp status finished updating
		utils.EventuallyDoubleLong(
			func() (string, error) {
				args := []string{"mcp", "-o", `jsonpath={.items[*].status.conditions[?(@.type=="Updated")].status}`}
				return OC_Get(kubeconfig, verbose, args...)
			},
			func(v string) bool {
				return utils.AllTrue(v)
			},
		)
		// wait until mcp status is not updating anymore
		utils.EventuallyDoubleLong(
			func() (string, error) {
				args := []string{"mcp", "-o", `jsonpath={.items[*].status.conditions[?(@.type=="Updating")].status}`}
				return OC_Get(kubeconfig, verbose, args...)
			},
			func(v string) bool {
				return utils.AllFalse(v)
			},
		)
	}
	return nil
}

// CreateTLSSecretIfNotExists tries to create a TLS secret from tlsKeyFile and tlsCertFile.
// If the secret with name: secretName exists already, it fails.
func CreateTLSSecretIfNotExists(kubeconfig, namespace, secretName, tlsKeyFile, tlsCertFile string) (*corev1.Secret, error) {
	client, ctx, log, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	ns, err := CreateNamespaceIfNotExists(kubeconfig, namespace)
	if err != nil {
		return nil, err
	}
	secret := &corev1.Secret{}
	err = client.Get(ctx, occlient.ObjectKey{Name: secretName, Namespace: ns.Name}, secret)
	if err == nil {
		return nil, fmt.Errorf("Secret %s already exists in namespace %s\n", secretName, ns.Name)
	}
	if !errors.IsNotFound(err) {
		// other errors, just return
		return nil, err
	}
	// ok, not found, then create it
	tlsKeyData, err := os.ReadFile(tlsKeyFile)
	if err != nil {
		return nil, err
	}
	tlsCertData, err := os.ReadFile(tlsCertFile)
	if err != nil {
		return nil, err
	}
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       tlsCertData,
			corev1.TLSPrivateKeyKey: tlsKeyData,
		},
	}
	if err := client.Create(ctx, secret); err != nil {
		return nil, err
	}
	log.Info("TLS secret %s created in namespace %s\n", secretName, namespace)
	return secret, nil
}

// CreateHTPasswdSecret creates a htpasswd style user+Bcrypt(hash(password))
func CreateHTPasswdSecret(kubeconfig, namespace, secretName, user, password string) (*corev1.Secret, error) {
	client, ctx, log, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	ns, err := CreateNamespaceIfNotExists(kubeconfig, namespace)
	if err != nil {
		return nil, err
	}
	secret := &corev1.Secret{}
	err = client.Get(ctx, occlient.ObjectKey{Name: secretName, Namespace: ns.Name}, secret)
	if err == nil {
		return nil, fmt.Errorf("Secret %s already exists in namespace %s\n", secretName, ns.Name)
	}
	if !errors.IsNotFound(err) {
		// other errors, just return
		return nil, err
	}
	htpasswdAuth, err := auth.GenerateHtpasswdBcrypt(user, password)
	if err != nil {
		return nil, err
	}
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: ns.Name,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"htpasswd": []byte(htpasswdAuth),
		},
	}
	if err := client.Create(ctx, secret); err != nil {
		return nil, err
	}
	log.Info("Htpasswd secret %s created in namespace %s\n", secretName, ns.Name)
	return secret, nil
}

// BaseDomain returns the cluster name, the base domain name if succeeds
func BaseDomain(kubeconfig string) (string, string, error) {
	client, ctx, _, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return "", "", err
	}
	ingress := &configv1.Ingress{}
	if err := client.Get(ctx, occlient.ObjectKey{Name: "cluster"}, ingress); err != nil {
		return "", "", fmt.Errorf("failed to get ingress/cluster: %w", err)
	}

	appsDomain := ingress.Spec.Domain
	// e.g. "apps.cluster-xyz.example.com"

	parts := strings.SplitN(appsDomain, ".", 2)
	if len(parts) < 2 || !strings.HasPrefix(parts[0], "apps") {
		return "", "", fmt.Errorf("unexpected apps domain format: %s", appsDomain)
	}

	// remove "apps." prefix
	withoutApps := strings.TrimPrefix(appsDomain, "apps.")
	parts = strings.SplitN(withoutApps, ".", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("unexpected apps domain format after trim: %s", withoutApps)
	}

	clusterName := parts[0] // "cluster-xyz"
	baseDomain := parts[1]  // "example.com"

	return clusterName, baseDomain, nil

}
