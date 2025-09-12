## openqe openshift create-image-registry

Create an image registry on the current OpenShift cluster with tls and authentication enabled

```
openqe openshift create-image-registry [flags]
```

### Options

```
      --ca-cert-file string    The CA certificate file path to be generated to. (default "ca.crt")
      --ca-dns-name string     The SAN used to generate the TLS CA. (default "openqe.github.io")
      --ca-key-file string     The CA private key file path to be generated to. (default "ca.key")
      --ca-subject string      The CA certificate subject used to generate the TLS CA. (default "C=China, O=OpenShift, OU=Hypershift QE, CN=default-ca")
      --dns-name string        The SAN added to the TLS certificate. (default "server.openqe.github.io")
  -h, --help                   help for create-image-registry
      --image string           The image used for the image registry (default "quay.io/openshifttest/registry:2")
      --kubeconfig string      The kubeconfig file used to communicate with the OpenShift cluster (default "/home/lgao/.kube/config")
      --name string            The image registry name (default "my-registry")
      --namespace string       The namespace in which the image registry will be deployed (default "test-registry")
      --password string        The password that can be used to access the image registry (default "reg-pass")
      --subject string         The TLS certificate subject. (default "C=China, O=OpenShift, OU=Hypershift QE, CN=default-server")
      --tls-cert-file string   The file path of the TLS certificate to be generated to. (default "tls.crt")
      --tls-key-file string    The file path of the TLS private key to be generated to. (default "tls.key")
      --user string            The username that can be used to access the image registry (default "reg-user")
      --verbose                If more information should be printed during the setup.
```

### SEE ALSO

* [openqe openshift](openqe_openshift.md)	 - OpenShift oriented test utilities

