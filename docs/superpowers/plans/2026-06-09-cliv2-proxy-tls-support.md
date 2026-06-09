# CLIv2 Proxy & TLS Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the Go CLIv2 a single HTTP client factory that honors proxy env vars, supports a custom CA bundle (`SSL_CERT_FILE`), and offers an explicit insecure-TLS toggle (`CODACY_CLI_INSECURE`), then route every HTTP callsite through it.

**Architecture:** A new `utils/httpclient` package builds `*http.Client` instances whose `http.Transport` always sets `Proxy: http.ProxyFromEnvironment` and a `TLSClientConfig` assembled from the system cert pool plus an optional `SSL_CERT_FILE` bundle (or `InsecureSkipVerify` when the insecure toggle is set). All ~7 existing callsites switch to `httpclient.New(...)`.

**Tech Stack:** Go 1.24, stdlib `net/http` + `crypto/tls` + `crypto/x509`, `testify/assert`, `net/http/httptest`. Module path: `codacy/cli-v2`.

**Spec:** `docs/superpowers/specs/2026-06-09-cliv2-proxy-tls-support-design.md`

---

## File Structure

| File | Responsibility |
|------|----------------|
| `utils/httpclient/options.go` (create) | `Options`, `WithTimeout` functional option, env-var constants + readers |
| `utils/httpclient/tls.go` (create) | `buildTLSConfig()` — system pool + `SSL_CERT_FILE` append + insecure toggle + stderr warning |
| `utils/httpclient/client.go` (create) | `New(opts ...Option) (*http.Client, error)` — wires proxy + TLS into a transport |
| `utils/httpclient/httpclient_test.go` (create) | Hermetic tests: TLS default-fail, custom CA, insecure, bad bundle, proxy set |
| `codacy-client/client.go` (modify) | `getRequest` uses `httpclient.New` |
| `tools/patterns.go` (modify) | `FetchDefaultEnabledPatterns` uses `httpclient.New` |
| `utils/download.go` (modify) | `DownloadFile` uses `httpclient.New(WithTimeout(0))` |
| `cmd/upload.go` (modify) | 4 callsites use `httpclient.New` |
| `cmd/upload_sbom.go` (modify) | package-var client built lazily via `httpclient.New`, keep `httpDoer` test injection |
| `README.md` (modify) | "Proxy & TLS" docs section |

**Important environment notes for the implementer:**
- `logger.Warn` writes to a rotating **log file** and only after `logger.Initialize` ran (`fileLogger != nil`). It will NOT surface the insecure warning to the user. The insecure warning MUST go to **stderr** via `fmt.Fprintln(os.Stderr, ...)`.
- `http.ProxyFromEnvironment` caches its resolved proxy config with a `sync.Once` for the life of the process. Do NOT write a unit test that sets `HTTPS_PROXY`/`NO_PROXY` with `t.Setenv` and asserts on `tr.Proxy(req)` — it will be flaky/cached. Unit tests assert only that `tr.Proxy != nil`. Real proxy traversal + `NO_PROXY` bypass are covered by the existing real-life harness `integration-tests/proxy-tls/run.sh`.
- Confirmed on macOS: Go does NOT auto-read `SSL_CERT_FILE` into the system pool. The pool must be built explicitly (`x509.SystemCertPool()` + `AppendCertsFromPEM`).

---

## Task 1: Create `utils/httpclient` package — options

**Files:**
- Create: `utils/httpclient/options.go`

- [ ] **Step 1: Write options.go**

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./utils/httpclient/`
Expected: PASS (no output). The package has no other files yet, so this only checks syntax.

- [ ] **Step 3: Commit**

```bash
git add utils/httpclient/options.go
git commit -m "feat(httpclient): add options and env-var readers"
```

---

## Task 2: TLS config builder (TDD)

**Files:**
- Create: `utils/httpclient/tls.go`
- Test: `utils/httpclient/httpclient_test.go`

- [ ] **Step 1: Write the failing tests**

Create `utils/httpclient/httpclient_test.go`:

```go
package httpclient

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeServerCAFile writes the test server's cert to a PEM file and returns its path.
func writeServerCAFile(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	certPEM := "-----BEGIN CERTIFICATE-----\n"
	// httptest sets srv.Certificate(); encode it to PEM.
	der := srv.Certificate().Raw
	_ = der
	// Use the helper from crypto/x509 via pem in the real impl test below.
	return certPEM
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

// certToPEM encodes the test server's leaf certificate as PEM.
func certToPEM(t *testing.T, srv *httptest.Server) []byte {
	t.Helper()
	return encodeCertPEM(srv.Certificate().Raw)
}

var _ = time.Second // keep time import if unused elsewhere
var _ = tls.VersionTLS12
var _ = writeServerCAFile
```

Note: the helper `writeServerCAFile` above is unused scaffolding; delete it once `certToPEM`/`encodeCertPEM` are in place. `encodeCertPEM` is defined in tls.go (Step 3).

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./utils/httpclient/ -run TestBuildTLSConfig -v`
Expected: FAIL — `undefined: buildTLSConfig` and `undefined: encodeCertPEM`.

- [ ] **Step 3: Write tls.go**

```go
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

// encodeCertPEM is a test/helper utility: encode DER cert bytes as PEM.
func encodeCertPEM(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./utils/httpclient/ -run TestBuildTLSConfig -v`
Expected: PASS — all 5 `TestBuildTLSConfig_*` tests green.

- [ ] **Step 5: Commit**

```bash
git add utils/httpclient/tls.go utils/httpclient/httpclient_test.go
git commit -m "feat(httpclient): build TLS config from SSL_CERT_FILE + insecure toggle"
```

---

## Task 3: Client factory (TDD)

**Files:**
- Create: `utils/httpclient/client.go`
- Test: `utils/httpclient/httpclient_test.go` (append)

- [ ] **Step 1: Append failing tests**

Add to `utils/httpclient/httpclient_test.go`:

```go
func TestNew_SetsProxyAndTimeout(t *testing.T) {
	os.Unsetenv(EnvInsecure)
	os.Unsetenv(EnvCABundle)
	c, err := New(WithTimeout(7 * time.Second))
	require.NoError(t, err)
	assert.Equal(t, 7*time.Second, c.Timeout)

	tr, ok := c.Transport.(*http.Transport)
	require.True(t, ok)
	// Proxy resolver must be wired (env resolution itself is covered by the
	// real-life harness; ProxyFromEnvironment caches and is unsafe to unit-test).
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./utils/httpclient/ -run TestNew -v`
Expected: FAIL — `undefined: New`.

- [ ] **Step 3: Write client.go**

```go
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
```

- [ ] **Step 4: Run the full package test suite**

Run: `go test ./utils/httpclient/ -v`
Expected: PASS — all tests green.

- [ ] **Step 5: Confirm hermetic loop is stable**

Run: `go test ./utils/httpclient/ -count=5`
Expected: `ok  codacy/cli-v2/utils/httpclient` (no FAIL across 5 iterations).

- [ ] **Step 6: Commit**

```bash
git add utils/httpclient/client.go utils/httpclient/httpclient_test.go
git commit -m "feat(httpclient): add New client factory wiring proxy + TLS"
```

---

## Task 4: Migrate `codacy-client/client.go`

**Files:**
- Modify: `codacy-client/client.go:13-32` (the `getRequest` function and `timeout` const)

- [ ] **Step 1: Replace the inline client**

In `codacy-client/client.go`, add the import `"codacy/cli-v2/utils/httpclient"` (internal-imports group). Replace the start of `getRequest`:

Current:
```go
func getRequest(url string, apiToken string) ([]byte, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
```

New:
```go
func getRequest(url string, apiToken string) ([]byte, error) {
	client, err := httpclient.New(httpclient.WithTimeout(timeout))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	req, err := http.NewRequest("GET", url, nil)
```

If `net/http` becomes otherwise unused in the file, keep it — `http.NewRequest` still uses it. (`fmt` is already imported.)

- [ ] **Step 2: Compile**

Run: `go build ./codacy-client/`
Expected: PASS.

- [ ] **Step 3: Run existing package tests**

Run: `go test ./codacy-client/`
Expected: PASS (or "no test files" — either is acceptable; no behavior change for the happy path).

- [ ] **Step 4: Commit**

```bash
git add codacy-client/client.go
git commit -m "refactor(codacy-client): route getRequest through httpclient factory"
```

---

## Task 5: Migrate `tools/patterns.go`

**Files:**
- Modify: `tools/patterns.go:12-25`

- [ ] **Step 1: Replace the inline client**

Add import `"codacy/cli-v2/utils/httpclient"`. Replace:

Current:
```go
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Fetch default patterns from Codacy API
	url := fmt.Sprintf("https://app.codacy.com/api/v3/tools/%s/patterns", toolUUID)
	req, err := http.NewRequest("GET", url, nil)
```

New:
```go
	client, err := httpclient.New(httpclient.WithTimeout(10 * time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	// Fetch default patterns from Codacy API
	url := fmt.Sprintf("https://app.codacy.com/api/v3/tools/%s/patterns", toolUUID)
	req, err := http.NewRequest("GET", url, nil)
```

If the `time` import becomes unused after this change, remove it. (It is still used here via `10 * time.Second`, so keep it.)

- [ ] **Step 2: Compile**

Run: `go build ./tools/`
Expected: PASS.

- [ ] **Step 3: Run tests**

Run: `go test ./tools/ -run TestFetchDefaultEnabledPatterns`
Expected: PASS or "no tests to run" (no such test exists today — acceptable).

- [ ] **Step 4: Commit**

```bash
git add tools/patterns.go
git commit -m "refactor(tools): route FetchDefaultEnabledPatterns through httpclient"
```

---

## Task 6: Migrate `utils/download.go`

**Files:**
- Modify: `utils/download.go:50` (and imports)

- [ ] **Step 1: Replace the inline client**

Add import `"codacy/cli-v2/utils/httpclient"`. Replace:

Current:
```go
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
```

New:
```go
	client, err := httpclient.New(httpclient.WithTimeout(0)) // no timeout: large binaries
	if err != nil {
		return "", fmt.Errorf("failed to create http client: %w", err)
	}
	req, err := http.NewRequest("GET", url, nil)
```

- [ ] **Step 2: Compile**

Run: `go build ./utils/...`
Expected: PASS.

- [ ] **Step 3: Run tests**

Run: `go test ./utils/...`
Expected: PASS (or "no test files").

- [ ] **Step 4: Commit**

```bash
git add utils/download.go
git commit -m "refactor(utils): route DownloadFile through httpclient factory"
```

---

## Task 7: Migrate `cmd/upload.go` (4 callsites)

**Files:**
- Modify: `cmd/upload.go` at lines ~254, ~272 (`&http.Client{}`) and ~344, ~375 (`http.DefaultClient.Do`)

- [ ] **Step 1: Replace `resultsFinalWithProjectToken` client (~line 254)**

Add import `"codacy/cli-v2/utils/httpclient"`. Replace:

```go
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
```

with:

```go
	client, err := httpclient.New()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
```

- [ ] **Step 2: Replace `resultsFinalWithAPIToken` client (~line 272)**

Apply the identical change as Step 1 to the second `client := &http.Client{}` in `resultsFinalWithAPIToken`.

- [ ] **Step 3: Replace `sendResultsWithProjectToken` DefaultClient (~line 344)**

Replace:

```go
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error sending results: %v\n", err)
		os.Exit(1)
	}
```

with:

```go
	client, err := httpclient.New()
	if err != nil {
		fmt.Printf("Error creating http client: %v\n", err)
		os.Exit(1)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending results: %v\n", err)
		os.Exit(1)
	}
```

- [ ] **Step 4: Replace `sendResultsWithAPIToken` DefaultClient (~line 375)**

Apply the identical change as Step 3 to the second `http.DefaultClient.Do(req)` in `sendResultsWithAPIToken`.

- [ ] **Step 5: Compile (catches unused `net/http` import)**

Run: `go build ./cmd/`
Expected: PASS. If the compiler reports `"net/http" imported and not used`, remove the `net/http` import; if it reports it IS used (e.g. `http.NewRequest` remains), keep it. `cmd/upload.go` still calls `http.NewRequest`, so the import stays.

- [ ] **Step 6: Run upload tests**

Run: `go test ./cmd/ -run TestUpload -v`
Expected: PASS (existing `cmd/upload_test.go` must stay green).

- [ ] **Step 7: Commit**

```bash
git add cmd/upload.go
git commit -m "refactor(upload): route legacy upload calls through httpclient"
```

---

## Task 8: Migrate `cmd/upload_sbom.go` (preserve test injection)

**Files:**
- Modify: `cmd/upload_sbom.go:32` (package var) and the upload function around line 242

**Context:** `sbomHTTPClient httpDoer = &http.Client{Timeout: 5 * time.Minute}` is a package-level var that `upload_sbom_test.go` overrides for injection. A package-var initializer cannot return an error, so we make the default lazy.

- [ ] **Step 1: Write/extend the failing test**

In `cmd/upload_sbom_test.go`, add a test asserting the default client is created via the factory when none is injected:

```go
func TestDefaultSBOMClient_UsesHTTPClientFactory(t *testing.T) {
	sbomHTTPClient = nil // force default path
	c, err := defaultSBOMClient()
	require.NoError(t, err)
	require.NotNil(t, c)
	hc, ok := c.(*http.Client)
	require.True(t, ok)
	assert.Equal(t, 5*time.Minute, hc.Timeout)
}
```

Add imports `"github.com/stretchr/testify/require"` and `"time"` to the test file if not already present.

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./cmd/ -run TestDefaultSBOMClient -v`
Expected: FAIL — `undefined: defaultSBOMClient`.

- [ ] **Step 3: Change the package var to lazy default**

Add import `"codacy/cli-v2/utils/httpclient"`. Replace:

```go
	sbomHTTPClient httpDoer = &http.Client{Timeout: 5 * time.Minute}
```

with:

```go
	// sbomHTTPClient is nil by default and resolved lazily via defaultSBOMClient.
	// Tests may set it to a stub implementing httpDoer.
	sbomHTTPClient httpDoer
```

Add this helper near the `httpDoer` interface definition:

```go
// defaultSBOMClient returns the injected client if set, else a factory client.
func defaultSBOMClient() (httpDoer, error) {
	if sbomHTTPClient != nil {
		return sbomHTTPClient, nil
	}
	return httpclient.New(httpclient.WithTimeout(5 * time.Minute))
}
```

- [ ] **Step 4: Use the helper at the request callsite (~line 242)**

Replace:

```go
	resp, err := sbomHTTPClient.Do(req)
```

with:

```go
	client, err := defaultSBOMClient()
	if err != nil {
		return fmt.Errorf("failed to create http client: %w", err)
	}
	resp, err := client.Do(req)
```

(If the enclosing function does not return `error`, adapt to its existing error style — check the signature of the function containing line 242 and match how it reports errors. The function is the SBOM upload handler; it returns an error.)

- [ ] **Step 5: Compile**

Run: `go build ./cmd/`
Expected: PASS. Remove the now-unused `time` import only if the compiler flags it (it remains used by `defaultSBOMClient`).

- [ ] **Step 6: Run SBOM tests**

Run: `go test ./cmd/ -run "TestUploadSBOM|TestDefaultSBOMClient" -v`
Expected: PASS — existing SBOM tests plus the new one are green.

- [ ] **Step 7: Commit**

```bash
git add cmd/upload_sbom.go cmd/upload_sbom_test.go
git commit -m "refactor(sbom): lazy default client via httpclient, keep test injection"
```

---

## Task 9: Full build + test sweep

**Files:** none (verification task)

- [ ] **Step 1: Build everything**

Run: `go build ./...`
Expected: PASS (no output).

- [ ] **Step 2: Run the whole unit suite**

Run: `go test ./...`
Expected: PASS for all packages (pre-existing network-dependent tests behave as before).

- [ ] **Step 3: Vet**

Run: `go vet ./...`
Expected: no findings in touched packages.

- [ ] **Step 4: Commit (if vet/build produced fixups)**

```bash
git add -A
git commit -m "chore: build/test/vet sweep for proxy-tls support" || echo "nothing to commit"
```

---

## Task 10: Documentation

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add a "Proxy & TLS" section**

Append to `README.md` (place after the configuration/usage section; match existing heading depth):

```markdown
## Proxy & TLS

The CLI honors standard proxy environment variables for all outbound HTTP(S):

- `HTTP_PROXY` / `HTTPS_PROXY` — proxy URL for plain/HTTPS requests
- `NO_PROXY` — comma-separated hosts that bypass the proxy

### Corporate proxies with TLS interception

If your proxy presents its own (MITM) certificate, point the CLI at the proxy's
CA bundle so TLS verification still passes:

```sh
export SSL_CERT_FILE=/path/to/corporate-ca.pem
```

`SSL_CERT_FILE` certificates are appended to the system trust store.

### Disabling TLS verification (last resort)

```sh
export CODACY_CLI_INSECURE=1
```

This disables certificate verification entirely and prints a warning. Prefer
`SSL_CERT_FILE`. Insecure mode is never enabled by default.

### Testing proxy/TLS behavior

`integration-tests/proxy-tls/run.sh` runs the CLI through a real `mitmproxy`
(`brew install mitmproxy`) against `app.codacy.com` and asserts the matrix above.
Loop with `PROXY_TLS_LOOP=5 integration-tests/proxy-tls/run.sh`.
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: document proxy and TLS configuration"
```

---

## Task 11: Real-life acceptance gate

**Files:** none (uses existing `integration-tests/proxy-tls/run.sh`)

- [ ] **Step 1: Build the binary**

Run: `make build`
Expected: produces `./cli-v2`.

- [ ] **Step 2: Run the real-life harness**

Run: `PROXY_PORT=8899 integration-tests/proxy-tls/run.sh`
Expected: ALL FOUR cases PASS now —
- A custom-CA: PASS (rc=0, proxy_saw=yes) ← was FAIL pre-implementation
- B no-CA-fails: PASS
- C insecure: PASS (rc=0, proxy_saw=yes) ← was FAIL pre-implementation
- D no_proxy-bypass: PASS
Final line: `ALL PROXY/TLS CHECKS PASSED`, exit 0.

- [ ] **Step 3: Confirm stability under loop**

Run: `PROXY_TLS_LOOP=3 PROXY_PORT=8899 integration-tests/proxy-tls/run.sh`
Expected: all cases PASS in every iteration.

> If A or C still fail with "certificate is not trusted", the TLS config is not
> being applied at that callsite — re-check that the failing path goes through
> `httpclient.New` (Tasks 4–8) and not a stray `http.Client`/`http.DefaultClient`.

---

## Self-Review Notes

- **Spec coverage:** proxy via `ProxyFromEnvironment` (Task 3 + all migrations); `NO_PROXY` (inherited from Go default, verified Task 11 case D); custom CA via `SSL_CERT_FILE` (Task 2, verified Tasks 3 & 11); insecure toggle (Task 2, verified Tasks 3 & 11); no insecure default (Task 2 default path, verified Task 11 case B); central factory + all 7 callsites (Tasks 4–8); error-on-bad-CA (Task 3, `TestNew_PropagatesCABundleError`); docs (Task 10); test loops — hermetic (Task 3 Step 5) and real-life (Task 11).
- **Insecure warning to stderr** (not `logger.Warn`) because the logger writes to a file and only post-Initialize — see env notes.
- **Proxy env unit-test caching** avoided per the `ProxyFromEnvironment` sync.Once note; real proxy/NO_PROXY behavior covered by the harness.
- **Out of scope (follow-up tickets, per spec):** extension/MCP mapping of `http.proxyStrictSSL` → `CODACY_CLI_INSECURE`; `CODACY_API_BASE_URL` override for black-box CLI runs.
```
