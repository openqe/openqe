package openshift

import (
	"context"
	"fmt"
	"io"

	"github.com/openqe/openqe/pkg/tls"
	"github.com/openqe/openqe/pkg/utils"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	occlient "sigs.k8s.io/controller-runtime/pkg/client"
)

func log(out io.Writer, message string, args ...any) {
	fmt.Fprintf(out, message, args...)
}

// SetupImageRegistry sets up an image registry on current OpenShift cluster and returns the route of the image registry
func SetupImageRegistry(opts *ImageRegistryOptions, out io.Writer) (string, error) {
	if out == nil {
		out = io.Discard
	}
	if !utils.FileExists(opts.PkiOpts.CaGenOpt.CaCertFile) || !utils.FileExists(opts.PkiOpts.CaGenOpt.CaKeyFile) {
		if opts.Verbose {
			log(out, "CA: key: %s, cert: %s are not ready, create CA", opts.PkiOpts.CaGenOpt.CaKeyFile, opts.PkiOpts.CaGenOpt.CaCertFile)
		}
		if err := tls.GenerateCAToFiles(opts.PkiOpts.CaGenOpt); err != nil {
			return "", err
		}
	}
	log(out, "CA is ready: ca key: %s, ca certificate: %s\n", opts.PkiOpts.CaGenOpt.CaKeyFile, opts.PkiOpts.CaGenOpt.CaCertFile)

	// update cluster proxy with the additional trusted bundle
	if err := ConfigureAdditionalCA(opts.OcpOpts.KUBECONFIG, opts.PkiOpts.CaGenOpt.CaCertFile, opts.Verbose, out); err != nil {
		return "", err
	}
	log(out, "Additional CA configured")

	// check tls cert/key
	if !utils.FileExists(opts.PkiOpts.CertFile) || !utils.FileExists(opts.PkiOpts.KeyFile) {
		if opts.Verbose {
			log(out, "TLS key/cert pair: key: %s, cert: %s are not ready, create key/cert pairs\n", opts.PkiOpts.CaGenOpt.CaKeyFile, opts.PkiOpts.CaGenOpt.CaCertFile)
		}
		if opts.PkiOpts.DNSName == tls.DefaultPKIOptions().DNSName {
			// set it according to *.apps.<base-domain>
			_, baseDomain, err := BaseDomain(opts.OcpOpts.KUBECONFIG)
			if err != nil {
				return "", err
			}
			opts.PkiOpts.DNSName = "*." + "apps." + baseDomain
			if opts.Verbose {
				log(out, "Set the TLS cert DNSName to %s\n", opts.PkiOpts.DNSName)
			}
		}
		if err := tls.GenerateTLSKeyCertPairToFiles(opts.PkiOpts); err != nil {
			return "", err
		}
	}
	log(out, "TLS Key/Cert pair is ready: key: %s, certificate: %s\n", opts.PkiOpts.KeyFile, opts.PkiOpts.CertFile)

	// create namespace
	ns, err := CreateNamespaceIfNotExists(opts.OcpOpts.KUBECONFIG, opts.Namespace, out)
	if err != nil {
		return "", err
	}
	log(out, "Namespace: %s is ready\n", ns.Name)

	// create secret for tls
	tlsSecret, err := CreateTLSSecretIfNotExists(opts.OcpOpts.KUBECONFIG, opts.Namespace, "test-reg-tls-secret", opts.PkiOpts.KeyFile, opts.PkiOpts.CertFile, out)
	if err != nil {
		return "", err
	}
	log(out, "TLS secret: %s is ready\n", tlsSecret.Name)

	// create secret for htpasswd
	htpasswdSecret, err := CreateHTPasswdSecret(opts.OcpOpts.KUBECONFIG, opts.Namespace, "test-reg-htpasswd", opts.User, opts.Password, out)
	if err != nil {
		return "", err
	}
	log(out, "Htpasswd secret: %s is ready\n", htpasswdSecret.Name)

	// create deployment
	deployment, err := CreateImageRegistryDeployment(opts.OcpOpts.KUBECONFIG, opts.Namespace, opts.Name, opts.Image, tlsSecret.Name, htpasswdSecret.Name, out)
	if err != nil {
		return "", err
	}
	log(out, "Deployment: %s is ready\n", deployment.Name)

	// create service
	service, err := createImageRegistryService(opts.OcpOpts.KUBECONFIG, opts.Namespace, opts.Name, out)
	if err != nil {
		return "", err
	}
	log(out, "Service: %s is ready\n", service.Name)

	// create route
	route, err := createImageRegistryRoute(opts.OcpOpts.KUBECONFIG, opts.Namespace, opts.Name, out)
	if err != nil {
		return "", err
	}
	log(out, "Route: %s is ready\n", route.Name)
	// get route host and return
	return route.Spec.Host, nil
}

func CreateImageRegistryDeployment(kubeconfig, namespace, name, image, tlsSecret, htpasswdSecret string, out io.Writer) (*appsv1.Deployment, error) {
	if out == nil {
		out = io.Discard
	}
	client, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()
	// ns, err := CreateNamespaceIfNotExists(kubeconfig, namespace, out)
	// if err != nil {
	// 	return nil, err
	// }
	deploy := &appsv1.Deployment{}
	err = client.Get(ctx, occlient.ObjectKey{Name: name, Namespace: namespace}, deploy)
	if err == nil {
		return nil, fmt.Errorf("Deployment %s already exists in namespace %s\n", name, namespace)
	}
	if !errors.IsNotFound(err) {
		return nil, err
	}
	// create it
	deploy = NewDeployment(namespace, name, image, 5000)
	// Volume for TLS secret
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "tls-cert",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: tlsSecret,
			},
		},
	})
	// Volume for htpasswd secret
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "htpasswd",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: htpasswdSecret,
			},
		},
	})
	// Mount the volumes to the container
	deploy.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{
			Name:      "tls-cert",
			MountPath: "/tls",
			ReadOnly:  true,
		},
		{
			Name:      "htpasswd",
			MountPath: "/auth",
			ReadOnly:  true,
		},
	}
	// Add env to the container
	deploy.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{
		{
			Name:  "REGISTRY_HTTP_TLS_CERTIFICATE",
			Value: "/tls/tls.crt",
		},
		{
			Name:  "REGISTRY_HTTP_TLS_KEY",
			Value: "/tls/tls.key",
		},
		{
			Name:  "REGISTRY_AUTH",
			Value: "htpasswd",
		},
		{
			Name:  "REGISTRY_AUTH_HTPASSWD_REALM",
			Value: "Registry Realm",
		},
		{
			Name:  "REGISTRY_AUTH_HTPASSWD_PATH",
			Value: "/auth/htpasswd",
		},
	}
	// create the deployment
	if err := client.Create(ctx, deploy); err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "Deployment %s created in namespace %s\n", name, namespace)
	return deploy, nil
}

func createImageRegistryService(kubeconfig, namespace, name string, out io.Writer) (*corev1.Service, error) {
	if out == nil {
		out = io.Discard
	}
	client, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()
	// ns, err := CreateNamespaceIfNotExists(kubeconfig, namespace, out)
	// if err != nil {
	// 	return nil, err
	// }
	svc := &corev1.Service{}
	err = client.Get(ctx, occlient.ObjectKey{Name: name, Namespace: namespace}, svc)
	if err == nil {
		return nil, fmt.Errorf("Service %s already exists in namespace %s\n", name, namespace)
	}
	if !errors.IsNotFound(err) {
		return nil, err
	}
	// create it
	svc = NewService(namespace, name, name, 5000, 5000)
	if err := client.Create(ctx, svc); err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "Service %s created in namespace %s\n", name, namespace)
	return svc, nil
}

func createImageRegistryRoute(kubeconfig, namespace, name string, out io.Writer) (*routev1.Route, error) {
	if out == nil {
		out = io.Discard
	}
	client, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()
	// ns, err := CreateNamespaceIfNotExists(kubeconfig, namespace, out)
	// if err != nil {
	// 	return nil, err
	// }
	route := &routev1.Route{}
	err = client.Get(ctx, occlient.ObjectKey{Name: name, Namespace: namespace}, route)
	if err == nil {
		return nil, fmt.Errorf("Route %s already exists in namespace %s\n", name, namespace)
	}
	if !errors.IsNotFound(err) {
		return nil, err
	}
	// create it
	route = NewRoute(namespace, name, name, 5000)
	if err := client.Create(ctx, route); err != nil {
		return nil, err
	}
	fmt.Fprintf(out, "Route %s created in namespace %s\n", name, namespace)
	return route, nil
}
