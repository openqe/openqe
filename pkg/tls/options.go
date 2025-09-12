package tls

type CAOptions struct {
	Subject    string
	DNSName    string
	CaKeyFile  string
	CaCertFile string
}

type PKIOptions struct {
	CaGenOpt *CAOptions
	Subject  string
	DNSName  string
	CertFile string
	KeyFile  string
}

func DefaultCAOptions() *CAOptions {
	return &CAOptions{
		Subject:    "C=China, O=OpenShift, OU=Hypershift QE, CN=default-ca",
		DNSName:    "openqe.github.io",
		CaKeyFile:  "ca.key",
		CaCertFile: "ca.crt",
	}
}

func DefaultPKIOptions() *PKIOptions {
	pkiOpts := &PKIOptions{
		CaGenOpt: DefaultCAOptions(),
		CertFile: "tls.crt",
		KeyFile:  "tls.key",
		Subject:  "C=China, O=OpenShift, OU=Hypershift QE, CN=default-server",
		DNSName:  "server.openqe.github.io",
	}
	return pkiOpts
}
