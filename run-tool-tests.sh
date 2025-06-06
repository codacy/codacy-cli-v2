#!/bin/bash

# Function to normalize paths in a file
normalize_paths() {
  local file=$1
  local path_prefix
  
  if [[ "$OSTYPE" == "darwin"* ]]; then
    path_prefix="/Users/runner/work/codacy-cli-v2/codacy-cli-v2/"
  else
    path_prefix="/home/runner/work/codacy-cli-v2/codacy-cli-v2/"
  fi
  
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|file://${path_prefix}|file:///|g" "$file"
    sed -i '' "s|${path_prefix}|/|g" "$file"
  else
    sed -i "s|file://${path_prefix}|file:///|g" "$file"
    sed -i "s|${path_prefix}|/|g" "$file"
  fi
}

# Function to sort SARIF file
sort_sarif() {
  local input=$1
  local output=$2
  jq --sort-keys '
    if .runs[0].tool.driver.rules == null then . else .runs[0].tool.driver.rules |= sort_by(.id) end
    | .runs[0].results |= sort_by(.ruleId, .message.text, .locations[0].physicalLocation.region.startLine, .locations[0].physicalLocation.region.startColumn)
  ' "$input" > "$output"
}

# Check if tool name is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <tool_name>"
  exit 1
fi

TOOL_NAME=$1
TOOL_DIR="plugins/tools/$TOOL_NAME/test/src"
TEST_DIR="plugins/tools/$TOOL_NAME/test"
CLI_PATH="$(pwd)/cli-v2"
EXPECTED_SARIF="$(pwd)/$TEST_DIR/expected.sarif"

# Validate environment
if [ ! -d "$TOOL_DIR" ]; then
  echo "Error: Tool directory $TOOL_DIR does not exist"
  exit 1
fi

if [ ! -f "$CLI_PATH" ]; then
  echo "Error: CLI binary not found at $CLI_PATH"
  exit 1
fi

# Change to the tool's test directory
cd "$TOOL_DIR" || exit 1

# Run analysis
"$CLI_PATH" install
"$CLI_PATH" analyze --tool "$TOOL_NAME" --format sarif --output actual.sarif

# Process SARIF files
normalize_paths actual.sarif
sort_sarif "$EXPECTED_SARIF" expected.sorted.json
sort_sarif actual.sarif actual.sorted.json
normalize_paths expected.sorted.json
normalize_paths actual.sorted.json

# Compare results
if ! diff expected.sorted.json actual.sorted.json; then
  echo "❌ Test output does not match expected output for $TOOL_NAME"
  echo -e "\nExpected SARIF output:"
  cat expected.sorted.json
  echo -e "\nActual SARIF output:"
  cat actual.sorted.json
  echo "$TOOL_NAME" >> /tmp/failed_tools.txt
  exit 1
else
  echo "✅ Tests passed successfully for $TOOL_NAME"
fi

# Return to original directory
cd ../../../../.. || exit 1 