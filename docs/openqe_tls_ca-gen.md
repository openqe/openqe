## openqe tls ca-gen

Generate CA key/cert pair to files

```
openqe tls ca-gen [flags]
```

### Options

```
      --ca-cert-file string   The CA certificate file path to be generated to. (default "ca.crt")
      --ca-dns-name string    The SAN used to generate the TLS CA. (default "openqe.github.io")
      --ca-key-file string    The CA private key file path to be generated to. (default "ca.key")
      --ca-subject string     The CA certificate subject used to generate the TLS CA. (default "C=China, O=OpenShift, OU=Hypershift QE, CN=default-ca")
  -h, --help                  help for ca-gen
```

### SEE ALSO

* [openqe tls](openqe_tls.md)	 - TLS oriented test utilities

