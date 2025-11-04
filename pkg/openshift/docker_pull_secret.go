package openshift

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openqe/openqe/pkg/common"
	"github.com/openqe/openqe/pkg/exec"
	"github.com/openqe/openqe/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	occlient "sigs.k8s.io/controller-runtime/pkg/client"
)

// UpsertDockerPullSecret creates or update a Docker pull secret in a namespace with Auths specified
// If the secret does not exist, it creates one with the auths specified
// If the secret exists already, it updates with the auths specified
func UpsertDockerPullSecret(opts *DockerPullSecretOptions) (*corev1.Secret, error) {

	kubeconfig := opts.OcpOpts.KUBECONFIG
	client, ctx, log, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Create namespace if it doesn't exist
	ns, err := CreateNamespaceIfNotExists(kubeconfig, opts.Namespace)
	if err != nil {
		return nil, err
	}
	// Check if secret already exists
	secret := &corev1.Secret{}
	err = client.Get(ctx, occlient.ObjectKey{Name: opts.SecretName, Namespace: ns.Name}, secret)
	if err == nil {
		// it exists, let's update it.
		dockerCfgBytes, ok := secret.Data[corev1.DockerConfigJsonKey]
		if !ok {
			return nil, fmt.Errorf("Secret %s in namespace %s does not has type: %s", opts.SecretName, ns.Name, corev1.DockerConfigJsonKey)
		}
		dockerCfg := &DockerConfig{}
		if err := json.Unmarshal(dockerCfgBytes, dockerCfg); err != nil {
			return nil, fmt.Errorf("Failed to construct the DockerConfig: %w", err)
		}
		finalCfg := MergeDockerConfig(dockerCfg, opts.DockerCfg)
		finalCfgJson, err := json.Marshal(finalCfg)
		if err != nil {
			return nil, fmt.Errorf("Failed to Marshal the final DockerConfig: %w", err)
		}
		secret.Data[corev1.DockerConfigJsonKey] = finalCfgJson
		if err := client.Update(ctx, secret); err != nil {
			return nil, err
		}
		return secret, nil
	}
	if !errors.IsNotFound(err) {
		// Other errors, just return
		return nil, err
	}
	// Create Docker config JSON
	dockerConfig := opts.DockerCfg
	dockerConfigJSON, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, err
	}
	// Create the secret
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.SecretName,
			Namespace: ns.Name,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerConfigJSON,
		},
	}
	if err := client.Create(ctx, secret); err != nil {
		return nil, err
	}
	log.Info("Docker pull secret %s created in namespace %s\n", opts.SecretName, ns.Name)
	return secret, nil
}

// CheckDockerPullSecretExists checks if a Docker pull secret exists in a namespace
func CheckDockerPullSecretExists(kubeconfig, namespace, secretName string) (bool, error) {
	client, ctx, _, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return false, err
	}

	secret := &corev1.Secret{}
	err = client.Get(ctx, occlient.ObjectKey{Name: secretName, Namespace: namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// DeleteDockerPullSecret deletes a Docker pull secret
func DeleteDockerPullSecret(kubeconfig, namespace, secretName string) error {
	client, ctx, log, err := GetOrCreateOCClient(kubeconfig)
	if err != nil {
		return err
	}

	// Check if secret exists
	secret := &corev1.Secret{}
	err = client.Get(ctx, occlient.ObjectKey{Name: secretName, Namespace: namespace}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Docker pull secret %s does not exist in the namespace %s\n", secretName, namespace)
			return nil
		}
		return err
	}

	// Delete the secret
	if err := client.Delete(ctx, secret); err != nil {
		return err
	}
	log.Info("Docker pull secret %s deleted from namespace %s\n", secretName, namespace)
	return nil
}

// ValidateDockerPullSecret validates a Docker pull secret by testing Docker registry authentication
func ValidateDockerPullSecret(kubeconfig, registryURL, pullSecretFile string, globalOpts *common.GlobalOptions) (bool, error) {
	// For validation, we'll use the oc CLI to test the secret
	// This is a simple test that tries to login to the registry

	if !utils.FileExists(pullSecretFile) {
		return false, fmt.Errorf("pull secret file: %s does not exist", pullSecretFile)
	}
	cli := &exec.CLI{
		ExecPath: "oc",
		Args: []string{
			"--kubeconfig", kubeconfig,
			"registry", "login",
			"--registry", registryURL,
			"--registry-config", pullSecretFile,
		},
		Verbose: globalOpts.Verbose,
	}
	output, err := cli.Execute()
	if err != nil {
		// Check if the error is related to authentication failure
		if strings.Contains(err.Error(), "unauthorized") || strings.Contains(err.Error(), "authentication") {
			return false, nil
		}
		return false, err
	}
	if strings.Contains(output, "unauthorized") || strings.Contains(output, "authentication") {
		return false, nil
	}
	return true, nil
}
