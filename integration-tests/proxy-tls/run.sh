#!/bin/bash
# Real-life proxy/TLS test for codacy-cli-v2 (OD-30).
#
# Runs the ACTUAL cli-v2 binary through a REAL mitmproxy MITM proxy against the
# real app.codacy.com, simulating a corporate TLS-intercepting proxy. Asserts:
#
#   A. proxy + custom CA (SSL_CERT_FILE)   -> success, traffic seen by proxy
#   B. proxy, no CA                        -> TLS verification failure, traffic seen
#   C. proxy + CODACY_CLI_INSECURE         -> success, traffic seen
#   D. NO_PROXY for app.codacy.com         -> success, proxy NOT traversed
#
# Cases A and C require the OD-30 feature (custom CA + insecure toggle). Before
# that is implemented they FAIL with "certificate is not trusted" — that failure
# is the baseline that proves the feature is needed. After implementation, green.
#
# Loopable:  PROXY_TLS_LOOP=5 ./run.sh
# Requires:  mitmproxy (mitmdump).  brew install mitmproxy
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI="$REPO_ROOT/cli-v2"
PROXY_PORT="${PROXY_PORT:-8899}"
CA="$HOME/.mitmproxy/mitmproxy-ca-cert.pem"
WORK="$(mktemp -d)"
export PROXY_CONNECT_LOG="$WORK/connects.txt"
MITM_PID=""

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }

cleanup() {
  [ -n "$MITM_PID" ] && kill "$MITM_PID" 2>/dev/null
  rm -rf "$WORK"
}
trap cleanup EXIT

command -v mitmdump >/dev/null 2>&1 || { red "mitmdump not found. Install: brew install mitmproxy"; exit 2; }
[ -x "$CLI" ] || { echo "Building cli-v2..."; (cd "$REPO_ROOT" && make build) || exit 2; }

# Start proxy with the connect-logging addon.
mitmdump -p "$PROXY_PORT" -q -s "$HERE/connect_logger.py" >"$WORK/mitm.log" 2>&1 &
MITM_PID=$!

# Wait for proxy to bind and generate its CA.
for _ in $(seq 1 40); do
  [ -f "$CA" ] && nc -z localhost "$PROXY_PORT" 2>/dev/null && break
  sleep 0.3
done
[ -f "$CA" ] || { red "mitmproxy CA not generated at $CA"; cat "$WORK/mitm.log"; exit 2; }

# Fresh, network-touching, tokenless CLI command. init hits app.codacy.com/api/v3.
# Args are VAR=val pairs prepended to the cli invocation via env.
run_init() {
  local dir="$WORK/proj.$RANDOM"
  mkdir -p "$dir"
  ( cd "$dir" && env "$@" "$CLI" init >"$WORK/last.log" 2>&1 )
  local rc=$?
  rm -rf "$dir"
  return $rc
}

proxy_saw_codacy() { grep -q "codacy.com" "$PROXY_CONNECT_LOG" 2>/dev/null; }

FAILURES=0
# check NAME EXPECT_RC(0|fail) EXPECT_PROXY(yes|no) -- VAR=val ...
check() {
  local name="$1" want_rc="$2" want_proxy="$3"; shift 3; [ "$1" = "--" ] && shift
  : >"$PROXY_CONNECT_LOG"
  run_init "$@"; local rc=$?
  sleep 0.3  # let addon flush
  local saw="no"; proxy_saw_codacy && saw="yes"
  local ok=1
  [ "$want_rc" = "0" ]    && [ "$rc" -ne 0 ] && ok=0
  [ "$want_rc" = "fail" ] && [ "$rc" -eq 0 ] && ok=0
  [ "$want_proxy" != "$saw" ] && ok=0
  if [ "$ok" -eq 1 ]; then
    green "PASS  $name (rc=$rc, proxy_saw=$saw)"
  else
    red "FAIL  $name (rc=$rc want=$want_rc, proxy_saw=$saw want=$want_proxy)"
    echo "----- cli output (tail) -----"; tail -3 "$WORK/last.log" 2>/dev/null; echo "-----------------------------"
    FAILURES=$((FAILURES+1))
  fi
}

run_suite() {
  local P="http://localhost:$PROXY_PORT"
  echo "== A: proxy + custom CA (needs OD-30) =="
  check "A custom-CA"        0    yes -- HTTPS_PROXY="$P" HTTP_PROXY="$P" SSL_CERT_FILE="$CA"
  echo "== B: proxy, no CA (expect TLS failure) =="
  check "B no-CA-fails"      fail yes -- HTTPS_PROXY="$P" HTTP_PROXY="$P"
  echo "== C: proxy + insecure (needs OD-30) =="
  check "C insecure"         0    yes -- HTTPS_PROXY="$P" HTTP_PROXY="$P" CODACY_CLI_INSECURE=1
  echo "== D: NO_PROXY bypass =="
  check "D no_proxy-bypass"  0    no  -- HTTPS_PROXY="$P" NO_PROXY="app.codacy.com,api.codacy.com" SSL_CERT_FILE="$CA"
}

LOOP="${PROXY_TLS_LOOP:-1}"
for i in $(seq 1 "$LOOP"); do
  [ "$LOOP" -gt 1 ] && echo "### iteration $i/$LOOP ###"
  run_suite
done

echo
if [ "$FAILURES" -eq 0 ]; then green "ALL PROXY/TLS CHECKS PASSED"; else red "$FAILURES check(s) FAILED"; fi
exit "$FAILURES"
