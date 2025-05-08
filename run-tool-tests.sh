#!/bin/bash

# Check if tool name is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <tool_name>"
  exit 1
fi

TOOL_NAME=$1
TOOL_DIR="plugins/tools/$TOOL_NAME/test/src"
CLI_PATH="$(pwd)/cli-v2"

# Check if tool directory exists
if [ ! -d "$TOOL_DIR" ]; then
  echo "Error: Tool directory $TOOL_DIR does not exist"
  exit 1
fi

# Check if CLI binary exists
if [ ! -f "$CLI_PATH" ]; then
  echo "Error: CLI binary not found at $CLI_PATH"
  exit 1
fi

# Change to the tool's test directory
cd "$TOOL_DIR"

# Install the tool
"$CLI_PATH" install

# Run analysis
"$CLI_PATH" analyze --tool $TOOL_NAME --format sarif --output actual.sarif

# Convert absolute paths to relative paths in the output
# Handle both path formats: with and without extra codacy-cli-v2
sed -i 's|file:///home/runner/work/codacy-cli-v2/|file:///|g' actual.sarif

# Sort all fields in both files, handling null rules array
jq --sort-keys 'if .runs[0].tool.driver.rules == null then . else .runs[0].tool.driver.rules |= sort_by(.id) end' expected.sarif > expected.sorted.json
jq --sort-keys 'if .runs[0].tool.driver.rules == null then . else .runs[0].tool.driver.rules |= sort_by(.id) end' actual.sarif > actual.sorted.json

# Run diff and capture its exit code
if ! diff expected.sorted.json actual.sorted.json; then
  echo "❌ Test output does not match expected output for $TOOL_NAME"
  echo -e "\nExpected SARIF output:"
  cat expected.sorted.json
  echo -e "\nActual SARIF output:"
  cat actual.sorted.json
  # Write to a file to track failures
  echo "$TOOL_NAME" >> /tmp/failed_tools.txt
  exit 1
else
  echo "✅ Tests passed successfully for $TOOL_NAME"
fi

# Return to original directory
cd ../../../../.. 