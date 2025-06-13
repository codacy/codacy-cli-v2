#!/bin/bash

# Function to normalize paths in a file
normalize_paths() {
  local file=$1
  # Get the repository root directory (5 levels up from current test/src directory)
  local repo_root=$(cd ../../../../.. && pwd)
  
  # Normalize absolute paths to relative ones for consistent testing
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # Replace absolute paths with relative paths in URI contexts
    sed -i '' "s#file://${repo_root}/plugins/tools/#file:///plugins/tools/#g" "$file"
    sed -i '' "s#${repo_root}/plugins/tools/#/plugins/tools/#g" "$file"
    # Handle CI runner paths for macOS
    sed -i '' "s#file:///Users/runner/work/codacy-cli-v2/codacy-cli-v2/plugins/tools/#file:///plugins/tools/#g" "$file"
    sed -i '' "s#/Users/runner/work/codacy-cli-v2/codacy-cli-v2/plugins/tools/#/plugins/tools/#g" "$file"
  else
    # Replace absolute paths with relative paths in URI contexts  
    sed -i "s#file://${repo_root}/plugins/tools/#file:///plugins/tools/#g" "$file"
    sed -i "s#${repo_root}/plugins/tools/#/plugins/tools/#g" "$file"
    # Handle CI runner paths for Linux
    sed -i "s#file:///home/runner/work/codacy-cli-v2/codacy-cli-v2/plugins/tools/#file:///plugins/tools/#g" "$file"
    sed -i "s#/home/runner/work/codacy-cli-v2/codacy-cli-v2/plugins/tools/#/plugins/tools/#g" "$file"
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