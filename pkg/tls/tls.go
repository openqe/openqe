package tls

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gaol/openqe/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

func defaultCertCfg() CertCfg {
	def := &CertCfg{
		KeySize:      DefaultKeySize,
		DNSNames:     []string{"openqe.github.io"},
		ExtKeyUsages: nil,
		IPAddresses:  nil,
		KeyUsages:    x509.KeyUsageDigitalSignature,
		Subject:      pkix.Name{Country: []string{"China"}, Organization: []string{"OpenShift"}, OrganizationalUnit: []string{"Hypershift QE"}, CommonName: "default-ca"},
		Validity:     365 * 24 * time.Hour, // 1 year
		IsCA:         false,
	}
	cfg := *def
	if def.DNSNames != nil {
		cfg.DNSNames = append([]string{}, def.DNSNames...)
	}
	if def.ExtKeyUsages != nil {
		cfg.ExtKeyUsages = append([]x509.ExtKeyUsage{}, def.ExtKeyUsages...)
	}
	if def.IPAddresses != nil {
		cfg.IPAddresses = append([]net.IP{}, def.IPAddresses...)
	}
	return cfg
}

/* Generates CA private key and certificate */
func GenerateCA() (*rsa.PrivateKey, *x509.Certificate, error) {
	cfg := defaultCertCfg()
	cfg.IsCA = true
	return GenerateSelfSignedCertificate(&cfg)
}

/* Generates CA private key and certificate with subject and dnsName specified */
func GenerateCAWith(subject, dnsName string) (*rsa.PrivateKey, *x509.Certificate, error) {
	cfg := defaultCertCfg()
	cfg.IsCA = true
	if subject != "" {
		cfg.Subject = ParseSubject(subject)
	}
	if dnsName != "" {
		cfg.DNSNames = []string{dnsName}
	}

	return GenerateSelfSignedCertificate(&cfg)
}

// Parse the openssl fashion subject into pkix.Name used in go
func ParseSubject(subject string) pkix.Name {
	var rdns []string
	if strings.HasPrefix(subject, "/") {
		// OpenSSL oneline format: /O=xx/OU=yy/CN=zz/C=cc
		subject = strings.TrimPrefix(subject, "/")
		rdns = strings.Split(subject, "/")
	} else {
		// RFC2253 format: CN=zz,O=xx,OU=yy,C=cc
		rdns = strings.Split(subject, ",")
	}
	name := pkix.Name{}
	for _, rdn := range rdns {
		parts := strings.SplitN(strings.TrimSpace(rdn), "=", 2)
		if len(parts) != 2 {
			continue
		}
		k, v := strings.ToUpper(parts[0]), parts[1]
		switch k {
		case "C":
			name.Country = append(name.Country, v)
		case "O":
			name.Organization = append(name.Organization, v)
		case "OU":
			name.OrganizationalUnit = append(name.OrganizationalUnit, v)
		case "CN":
			name.CommonName = v
		case "L":
			name.Locality = append(name.Locality, v)
		case "ST":
			name.Province = append(name.Province, v)
		default:
			// If you want to support arbitrary attributes, you can store them in ExtraNames
			name.ExtraNames = append(name.ExtraNames,
				pkix.AttributeTypeAndValue{Type: []int{2, 5, 4, 9999}, Value: v})
		}
	}
	return name
}

/** Generates a CA key/cert pair and save them into different files **/
func GenerateCAToFiles(opts *CAOptions) error {
	subject := opts.Subject
	dnsName := opts.DNSName
	caKeyFile := opts.CaKeyFile
	caCertFile := opts.CaCertFile
	if caKeyFile == "" {
		return errors.New("caKeyFile needs to be specified to save for the TLS CA private key")
	}
	if caCertFile == "" {
		return errors.New("caCertFile needs to be specified to save for the TLS CA certificate")
	}
	key, cert, err := GenerateCAWith(subject, dnsName)
	if err != nil {
		return err
	}
	keyInPem := PrivateKeyToPem(key)
	kf, err := os.Create(caKeyFile)
	if err != nil {
		return err
	}
	defer kf.Close()
	_, err = kf.Write(keyInPem)
	if err != nil {
		return err
	}
	certInPem := CertToPem(cert)
	certf, err := os.Create(caCertFile)
	if err != nil {
		return err
	}
	defer certf.Close()
	_, err = certf.Write(certInPem)
	if err != nil {
		return err
	}
	return nil
}

func GenerateTLSKeyCertPairToFiles(opts *PKIOptions) error {
	subject := opts.Subject
	dnsName := opts.DNSName
	caKeyFile := opts.CaGenOpt.CaKeyFile
	caCertFile := opts.CaGenOpt.CaCertFile
	tlsKeyFile := opts.KeyFile
	tlsCertFile := opts.CertFile
	if tlsKeyFile == "" {
		return errors.New("tlsKeyFile needs to be specified to save for the TLS private key")
	}
	if tlsCertFile == "" {
		return errors.New("tlsCertFile needs to be specified to save for the TLS certificate")
	}
	key, cert, err := GenerateTLSKeyCertPair(subject, dnsName, caKeyFile, caCertFile)
	if err != nil {
		return fmt.Errorf("Failed to generate TLS certificate: %w", err)
	}
	keyInPem := PrivateKeyToPem(key)
	kf, err := os.Create(tlsKeyFile)
	if err != nil {
		return err
	}
	defer kf.Close()
	_, err = kf.Write(keyInPem)
	if err != nil {
		return err
	}
	certInPem := CertToPem(cert)
	certf, err := os.Create(tlsCertFile)
	if err != nil {
		return err
	}
	defer certf.Close()
	_, err = certf.Write(certInPem)
	if err != nil {
		return err
	}
	return nil
}

func GenerateTLSKeyCertPair(subject, dnsName, caKeyFile, caCertFile string) (*rsa.PrivateKey, *x509.Certificate, error) {
	if caKeyFile == "" || !utils.FileExists(caKeyFile) {
		return nil, nil, errors.New("A valid caKeyFile needs to be specified to read the TLS CA private key")
	}
	if caCertFile == "" || !utils.FileExists(caCertFile) {
		return nil, nil, errors.New("A valid caCertFile needs to be specified to read the TLS CA certificate")
	}
	caKeyBytes, err := os.ReadFile(caKeyFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA private key file: %w", err)
	}
	caKey, err := PemToPrivateKey(caKeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load CA private key from file: %w", err)
	}
	caCertBytes, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA certificate file: %w", err)
	}
	caCert, err := PemToCertificate(caCertBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load CA certificate from file: %w", err)
	}
	cfg := defaultCertCfg()
	cfg.IsCA = false
	if subject != "" {
		cfg.Subject = ParseSubject(subject)
	}
	if dnsName != "" {
		cfg.DNSNames = []string{dnsName}
	}
	return GenerateSignedCertificate(caKey, caCert, &cfg)
}

// CertInCAFile checks if a certificate represented by certs in PEM format is already inside the ca-bundle ConfigMap.
// Normally it can be used to test the system ca bundle file at: /etc/pki/tls/certs/ca-bundle.crt
func CertInCAFile(cert, caFile string) (bool, error) {
	if !utils.FileExists(caFile) {
		return false, nil
	}
	caContent, err := os.ReadFile(caFile)
	if err != nil {
		return false, err
	}
	return CheckCert(cert, string(caContent))
}

// CertInConfigMap checks if a certificate represented by certs in PEM format is already inside the ca-bundle ConfigMap.
func CertInConfigMap(cm *corev1.ConfigMap, cert string) (bool, error) {
	bundlePEM, ok := cm.Data["ca-bundle.crt"]
	if !ok {
		return false, nil // no bundle present yet
	}
	return CheckCert(cert, bundlePEM)
}

// CheckCert checks if a certificate represented by cert exists in the certificates represented by certs.
func CheckCert(cert string, certs string) (bool, error) {
	certBlock, _ := pem.Decode([]byte(cert))
	wantCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return false, err
	}

	// Walk through all certs
	rest := []byte(certs)
	for {
		var b *pem.Block
		b, rest = pem.Decode(rest)
		if b == nil {
			break
		}
		if b.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(b.Bytes)
		if err != nil {
			continue
		}

		// Compare Raw DER (byte-for-byte identity)
		if cert.Equal(wantCert) {
			return true, nil
		}
	}
	return false, nil
}

// CheckCACertInBundle checks if a CA certificate file is included in a CA bundle file
func CheckCACertInBundle(caCertFile, caBundleFile string) (bool, error) {
	// Check if CA cert file exists
	if !utils.FileExists(caCertFile) {
		return false, fmt.Errorf("CA certificate file does not exist: %s", caCertFile)
	}

	// Check if CA bundle file exists
	if !utils.FileExists(caBundleFile) {
		return false, fmt.Errorf("CA bundle file does not exist: %s", caBundleFile)
	}

	// Read CA cert file
	caCertBytes, err := os.ReadFile(caCertFile)
	if err != nil {
		return false, fmt.Errorf("failed to read CA certificate file: %w", err)
	}

	// Read CA bundle file
	caBundleBytes, err := os.ReadFile(caBundleFile)
	if err != nil {
		return false, fmt.Errorf("failed to read CA bundle file: %w", err)
	}

	// Check if CA cert is in bundle
	return CheckCert(string(caCertBytes), string(caBundleBytes))
}
