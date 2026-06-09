package httpclient

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// certToPEM encodes the test server's leaf certificate as PEM.
func certToPEM(t *testing.T, srv *httptest.Server) []byte {
	t.Helper()
	return encodeCertPEM(srv.Certificate().Raw)
}

func TestBuildTLSConfig_DefaultRejectsSelfSigned(t *testing.T) {
	os.Unsetenv(EnvInsecure)
	os.Unsetenv(EnvCABundle)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	cfg, err := buildTLSConfig()
	require.NoError(t, err)
	c := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}}
	_, err = c.Get(srv.URL)
	assert.Error(t, err, "self-signed server must be rejected without a custom CA")
}

func TestBuildTLSConfig_CustomCASucceeds(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(caPath, certToPEM(t, srv), 0o600))
	t.Setenv(EnvCABundle, caPath)
	os.Unsetenv(EnvInsecure)

	cfg, err := buildTLSConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg.RootCAs)
	c := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}}
	resp, err := c.Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBuildTLSConfig_InsecureSkipsVerify(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	t.Setenv(EnvInsecure, "1")
	os.Unsetenv(EnvCABundle)

	cfg, err := buildTLSConfig()
	require.NoError(t, err)
	assert.True(t, cfg.InsecureSkipVerify)
	c := &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}}
	resp, err := c.Get(srv.URL)
	require.NoError(t, err)
	resp.Body.Close()
}

func TestBuildTLSConfig_MissingBundleErrors(t *testing.T) {
	t.Setenv(EnvCABundle, filepath.Join(t.TempDir(), "does-not-exist.pem"))
	os.Unsetenv(EnvInsecure)
	_, err := buildTLSConfig()
	assert.Error(t, err)
}

func TestBuildTLSConfig_BadBundleErrors(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "bad.pem")
	require.NoError(t, os.WriteFile(bad, []byte("not a certificate"), 0o600))
	t.Setenv(EnvCABundle, bad)
	os.Unsetenv(EnvInsecure)
	_, err := buildTLSConfig()
	assert.Error(t, err)
}

func TestNew_SetsProxyAndTimeout(t *testing.T) {
	os.Unsetenv(EnvInsecure)
	os.Unsetenv(EnvCABundle)
	c, err := New(WithTimeout(7 * time.Second))
	require.NoError(t, err)
	assert.Equal(t, 7*time.Second, c.Timeout)

	tr, ok := c.Transport.(*http.Transport)
	require.True(t, ok)
	// Proxy resolver must be wired. Env resolution itself is covered by the
	// real-life harness; ProxyFromEnvironment caches and is unsafe to unit-test.
	assert.NotNil(t, tr.Proxy)
	assert.NotNil(t, tr.TLSClientConfig)
}

func TestNew_PropagatesCABundleError(t *testing.T) {
	t.Setenv(EnvCABundle, filepath.Join(t.TempDir(), "missing.pem"))
	os.Unsetenv(EnvInsecure)
	_, err := New()
	assert.Error(t, err)
}

func TestNew_CustomCAEndToEnd(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	caPath := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(caPath, certToPEM(t, srv), 0o600))
	t.Setenv(EnvCABundle, caPath)
	os.Unsetenv(EnvInsecure)

	c, err := New()
	require.NoError(t, err)
	resp, err := c.Get(srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
