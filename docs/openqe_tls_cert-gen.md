## openqe tls cert-gen

Generate TLS key/cert pair to files, signed by a given CA

### Synopsis

Generate TLS key/cert pair to files, signed by a given CA.
The CA key/cert files must be provided to sign the generated TLS certificate.
You can use the 'tls ca-gen' command to generate a CA key/cert pair for testing purpose.

```
openqe tls cert-gen [flags]
```

### Options

```
      --ca-cert-file string    The CA certificate file path to be generated to. (default "ca.crt")
      --ca-dns-name string     The SAN used to generate the TLS CA. (default "openqe.github.io")
      --ca-key-file string     The CA private key file path to be generated to. (default "ca.key")
      --ca-subject string      The CA certificate subject used to generate the TLS CA. (default "C=China, O=OpenShift, OU=Hypershift QE, CN=default-ca")
      --dns-name string        The SAN added to the TLS certificate. (default "server.openqe.github.io")
  -h, --help                   help for cert-gen
      --subject string         The TLS certificate subject. (default "C=China, O=OpenShift, OU=Hypershift QE, CN=default-server")
      --tls-cert-file string   The file path of the TLS certificate to be generated to. (default "tls.crt")
      --tls-key-file string    The file path of the TLS private key to be generated to. (default "tls.key")
```

### SEE ALSO

* [openqe tls](openqe_tls.md)	 - TLS oriented test utilities

