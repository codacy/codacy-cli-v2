package httpclient

import "net/http"

// New returns an *http.Client whose transport honors proxy environment
// variables (HTTP_PROXY/HTTPS_PROXY/NO_PROXY) and applies CA/TLS configuration
// from SSL_CERT_FILE and CODACY_CLI_INSECURE.
//
// It returns an error if a configured CA bundle cannot be read or parsed, so
// callers fail loudly on misconfiguration rather than silently falling back to
// the system trust store.
func New(opts ...Option) (*http.Client, error) {
	o := &Options{}
	for _, fn := range opts {
		fn(o)
	}

	tlsCfg, err := buildTLSConfig()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: tlsCfg,
	}

	return &http.Client{
		Timeout:   o.Timeout,
		Transport: transport,
	}, nil
}
