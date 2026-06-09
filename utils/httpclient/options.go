package httpclient

import (
	"os"
	"strings"
	"time"
)

// Env vars controlling TLS/CA behavior.
const (
	// EnvInsecure, when truthy, disables TLS certificate verification.
	EnvInsecure = "CODACY_CLI_INSECURE"
	// EnvCABundle points to a PEM bundle appended to the system trust pool.
	// SSL_CERT_FILE is the OpenSSL-standard name corporate tooling already sets.
	EnvCABundle = "SSL_CERT_FILE"
)

// Options configure a client built by New.
type Options struct {
	// Timeout is the http.Client timeout. Zero means no timeout.
	Timeout time.Duration
}

// Option mutates Options.
type Option func(*Options)

// WithTimeout sets the client timeout. Pass 0 for no timeout (large downloads).
func WithTimeout(d time.Duration) Option {
	return func(o *Options) { o.Timeout = d }
}

// insecureEnv reports whether TLS verification is disabled via EnvInsecure.
func insecureEnv() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(EnvInsecure))) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

// caBundlePath returns the configured CA bundle path, or "" if unset.
func caBundlePath() string {
	return strings.TrimSpace(os.Getenv(EnvCABundle))
}
