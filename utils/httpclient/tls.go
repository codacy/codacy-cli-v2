package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// buildTLSConfig assembles the TLS config from env:
//   - CODACY_CLI_INSECURE truthy -> InsecureSkipVerify (with a stderr warning)
//   - SSL_CERT_FILE set           -> its PEM certs appended to the system pool
func buildTLSConfig() (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}

	if insecureEnv() {
		cfg.InsecureSkipVerify = true
		fmt.Fprintln(os.Stderr,
			"WARNING: TLS certificate verification is DISABLED (CODACY_CLI_INSECURE set). "+
				"Traffic can be intercepted. Prefer setting SSL_CERT_FILE to your proxy's CA instead.")
		return cfg, nil
	}

	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}

	if path := caBundlePath(); path != "" {
		pemBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA bundle from %s (%s): %w", EnvCABundle, path, err)
		}
		if !pool.AppendCertsFromPEM(pemBytes) {
			return nil, fmt.Errorf("no valid certificates found in CA bundle %s (%s)", EnvCABundle, path)
		}
	}

	cfg.RootCAs = pool
	return cfg, nil
}

// encodeCertPEM encodes DER certificate bytes as PEM. Used by tests.
func encodeCertPEM(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
