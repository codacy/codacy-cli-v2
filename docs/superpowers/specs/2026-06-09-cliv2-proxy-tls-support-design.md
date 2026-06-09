# CLIv2 Proxy & TLS Support — Design

- **Ticket:** [OD-30](https://linear.app/codacy/issue/OD-30/cliv2-verify-and-add-proxy-support) ([CF-2439](https://codacy.atlassian.net/browse/CF-2439))
- **Date:** 2026-06-09
- **Repo:** `codacy-cli-v2`
- **Status:** Approved design, pending implementation plan

## Goal

Ensure the Go analysis CLIv2 respects standard proxy and SSL configuration when run
directly, launched by the MCP server, or launched by the VS Code extension. Add custom
CA bundle support and an explicit insecure-TLS toggle that maps the VS Code
`http.proxyStrictSSL` setting.

## Current State (verified)

Roughly 7 HTTP callsites, each building its own client; none set a custom `Transport`:

| File | Callsite | Notes |
|------|----------|-------|
| `codacy-client/client.go` | `getRequest` | `&http.Client{Timeout: 10s}` (API v3 GETs) |
| `tools/patterns.go` | `FetchDefaultEnabledPatterns` | `&http.Client{Timeout: 10s}` |
| `utils/download.go` | `DownloadFile` | `&http.Client{}` (tool/runtime binary downloads) |
| `cmd/upload.go` | `resultsFinalWithProjectToken`, `resultsFinalWithAPIToken` | `&http.Client{}` (legacy api.codacy.com) |
| `cmd/upload.go` | `sendResultsWithProjectToken`, `sendResultsWithAPIToken` | `http.DefaultClient` |
| `cmd/upload_sbom.go` | `sbomHTTPClient` | `&http.Client{Timeout: 5m}` behind `httpDoer` interface |

**Key facts:**

1. Go's `http.DefaultTransport` already uses `http.ProxyFromEnvironment`. Omitting
   `Transport` (as all callsites do) means `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY`
   are **already honored today** everywhere. Proxy support is mostly *verify*, not *add*.
2. Subprocess tools/runtimes spawned by the CLI inherit the environment, so proxy env
   vars flow to them automatically.
3. No `InsecureSkipVerify` anywhere → no insecure default to remove. Good.

**Gaps:**

- No custom CA bundle support → corporate MITM proxies fail TLS.
- No shared client factory → cannot inject CA/TLS config without touching every callsite.
- No way to honor `http.proxyStrictSSL=false` from the extension.

## Approach

Central client factory (`utils/httpclient`) with an explicitly-built TLS root pool. All
callsites switch to it. Chosen over a shared-transport-only patch or a minimal
per-callsite patch because the ticket wants a single verified proxy/CA path, and the
existing `httpDoer` interface already signals a preference for injectable clients.

## Components

### `utils/httpclient` package

```
utils/httpclient/
  client.go      // New(opts...) *http.Client
  tls.go         // buildTLSConfig() (*tls.Config, error)
  options.go     // Option funcs (WithTimeout, ...) + env resolution
```

`New` always sets `Transport.Proxy = http.ProxyFromEnvironment` (locks proxy support)
and `Transport.TLSClientConfig` from `buildTLSConfig`. Returns a plain `*http.Client`,
so no callsite signatures change.

### TLS / CA construction (`tls.go`)

```go
func buildTLSConfig() (*tls.Config, error) {
    cfg := &tls.Config{MinVersion: tls.VersionTLS12}
    if insecureEnv() {
        cfg.InsecureSkipVerify = true
        logger.Warn("TLS verification disabled (CODACY_CLI_INSECURE). " +
            "Insecure — proxy MITM not validated.")
        return cfg, nil
    }
    pool, err := x509.SystemCertPool()
    if err != nil || pool == nil {
        pool = x509.NewCertPool()
    }
    if f := os.Getenv("SSL_CERT_FILE"); f != "" {
        pem, err := os.ReadFile(f)
        if err != nil {
            return nil, fmt.Errorf("read SSL_CERT_FILE %s: %w", f, err)
        }
        if !pool.AppendCertsFromPEM(pem) {
            return nil, fmt.Errorf("no valid certs in %s", f)
        }
    }
    cfg.RootCAs = pool
    return cfg, nil
}
```

The pool is built explicitly (system pool + appended `SSL_CERT_FILE`) rather than
relying on Go's implicit `SSL_CERT_FILE` handling, which varies by platform (notably
macOS). Corporate CA then works without disabling verification. Insecure mode
short-circuits and warns once.

### Configuration surface

| Env var | Effect | Default |
|---------|--------|---------|
| `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY` | Proxy routing (Go default) | unset |
| `SSL_CERT_FILE` | Extra CA bundle appended to system pool | unset |
| `CODACY_CLI_INSECURE` (`1`/`true`) | Disable TLS verification; maps `http.proxyStrictSSL=false` | off |

Env chosen as the signal channel because the extension and MCP server both control the
child process environment uniformly, with no CLI-arg plumbing.

## Callsite Migration

Replace every `&http.Client{...}` / `http.DefaultClient` with `httpclient.New(...)`:

- `codacy-client/client.go` → `New(WithTimeout(10 * time.Second))`
- `tools/patterns.go` → `New(WithTimeout(10 * time.Second))`
- `utils/download.go` → `New(WithTimeout(0))` (large binaries, no timeout)
- `cmd/upload.go` ×4 → `New()` (default)
- `cmd/upload_sbom.go` → `sbomHTTPClient = httpclient.New(WithTimeout(5 * time.Minute))`,
  preserving the `httpDoer` interface for test injection

## Error Handling

- CA read/parse failure returns an error and aborts the request — do not silently fall
  back to the system pool (surfaces misconfiguration).
- TLS handshake errors propagate as today; `codacy-client` error messages gain a hint:
  "if behind a corporate proxy, set SSL_CERT_FILE or CODACY_CLI_INSECURE".
- Insecure mode emits a single stderr warning per process.

## Testing

**Unit (`utils/httpclient`):**
- Proxy: `Transport.Proxy` non-nil after `New`.
- CA: load a test PEM via `SSL_CERT_FILE`, assert pool grows; bad PEM returns error.
- Insecure: `CODACY_CLI_INSECURE` sets `InsecureSkipVerify`.

**Integration:**
- `httptest.NewTLSServer` with self-signed cert: request fails by default, succeeds when
  `SSL_CERT_FILE` points at that cert, succeeds with `CODACY_CLI_INSECURE`.

**Manual matrix (documented in README):**
- Direct run + mitmproxy.
- `NO_PROXY` excluded host.
- Launched via MCP server.
- Launched via VS Code extension.

## Documentation

- README: new "Proxy & TLS" section covering the three env-var groups and the manual
  test matrix.

## Out of Scope (follow-up)

- `codacy-vscode-extension`: map `http.proxy` → `HTTP_PROXY`, `http.proxyStrictSSL=false`
  → `CODACY_CLI_INSECURE` when spawning the CLI. Same mapping for the MCP server.
  Tracked separately from OD-30 (this ticket is `[CLIv2]`).
