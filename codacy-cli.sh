#!/usr/bin/env bash

set -e +o pipefail

# Use the locally built codacy-cli-v2 binary
CODACY_CLI_V2_BIN="/Users/luia.ventura/Work/Repos/gh/codacy-cli-v2/codacy-cli-v2"

if [ ! -x "$CODACY_CLI_V2_BIN" ]; then
    echo "Error: codacy-cli-v2 binary not found at $CODACY_CLI_V2_BIN. Please build it first."
    exit 1
fi

exec "$CODACY_CLI_V2_BIN" "$@"