#!/bin/bash
set -e

# Get the absolute path of the script's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
CLI_PATH="$(pwd)/cli-v2"

echo "Script directory: $SCRIPT_DIR"
echo "Current working directory: $(pwd)"

# Check if API token is provided for token-based test
if [ -z "$CODACY_API_TOKEN" ]; then
  echo "Warning: CODACY_API_TOKEN environment variable is not set. Token-based test will be skipped."
fi

# Function to normalize and sort configuration values
normalize_config() {
  local file=$1
  local ext="${file##*.}"
  
  case "$ext" in
    yaml|yml)
      # For YAML files, use yq to sort
      yq e '.' "$file" | sort
      ;;
    rc|conf|ini)
      # For other config files, sort values after '=' and keep other lines
      awk -F'=' '
        /^[^#].*=.*,/ {
          split($2, values, ",")
          # Sort values using a simple bubble sort
          for (i=1; i<=length(values); i++) {
            for (j=i+1; j<=length(values); j++) {
              if (values[i] > values[j]) {
                temp = values[i]
                values[i] = values[j]
                values[j] = temp
              }
            }
          }
          printf "%s=", $1
          for (i=1; i<=length(values); i++) {
            if (i>1) printf ","
            printf "%s", values[i]
          }
          print ""
          next
        }
        { print }
      ' "$file" | sort
      ;;
    xml)
      # For XML files, ignore order of <rule ref=.../> lines and strip leading spaces
      awk '
        BEGIN { n = 0; end = ""; }
        /^ *<rule ref=/ { rules[++n] = $0; next }
        /^ *<\/ruleset>/ { end = $0; next }
        { gsub(/^ +/, "", $0); print }
        END {
          n = asort(rules, sorted_rules)
          for (i = 1; i <= n; i++) print sorted_rules[i]
          if (end) print end
        }
      ' "$file"
      ;;
    *)
      # For other files, just sort
      sort "$file"
      ;;
  esac
}

compare_files() {
  local expected_dir="$1"
  local actual_dir="$2"
  local label="$3"

  # Compare files in current directory
  for file in "$expected_dir"/*; do
    [ -f "$file" ] || continue
    filename=$(basename "$file")
    actual_file="$actual_dir/$filename"

    
    if [ ! -f "$actual_file" ]; then
      echo "❌ $label/$filename does not exist in actual output"
      echo "Expected: $file"
      echo "Actual should be: $actual_file"
      exit 1
    fi
    
    if diff <(normalize_config "$file") <(normalize_config "$actual_file") >/dev/null 2>&1; then
      echo "✅ $label/$filename matches expected"
    else
      echo "❌ $label/$filename does not match expected"
      echo "=== Expected (normalized) ==="
      normalize_config "$file"
      echo "=== Actual (normalized) ==="
      normalize_config "$actual_file"
      echo "=== Diff ==="
      diff <(normalize_config "$file") <(normalize_config "$actual_file") || true
      echo "==================="
      exit 1
    fi
  done

  # Compare subdirectories
  for dir in "$expected_dir"/*/; do
    [ -d "$dir" ] || continue
    dirname=$(basename "$dir")
    [ "$dirname" = "logs" ] && continue
    
    if [ ! -d "$actual_dir/$dirname" ]; then
      echo "❌ Directory $label/$dirname does not exist in actual output"
      echo "Expected: $dir"
      echo "Actual should be: $actual_dir/$dirname"
      exit 1
    fi
    compare_files "$dir" "$actual_dir/$dirname" "$label/$dirname"
  done
}

run_init_test() {
  local test_dir="$1"
  local test_name="$2"
  local use_token="$3"
  
  echo "Running test: $test_name"
  [ -d "$test_dir" ] || { echo "❌ Test directory does not exist: $test_dir"; exit 1; }
  
  cd "$test_dir" || exit 1
  rm -rf .codacy
  
  if [ "$use_token" = "true" ]; then
    [ -n "$CODACY_API_TOKEN" ] || { echo "❌ Skipping token-based test: CODACY_API_TOKEN not set"; return 0; }
    "$CLI_PATH" init --api-token "$CODACY_API_TOKEN" --organization troubleshoot-codacy-dev --provider gh --repository codacy-cli-test
  else
    "$CLI_PATH" init
  fi
  
  compare_files "expected" ".codacy" "Test $test_name"
  echo "✅ Test $test_name completed successfully"
  echo "----------------------------------------"
}

# Run both tests
echo "Starting integration tests..."
echo "----------------------------------------"

# Test 1: Init without token
run_init_test "$SCRIPT_DIR/init-without-token" "init-without-token" "false"

# Test 2: Init with token
run_init_test "$SCRIPT_DIR/init-with-token" "init-with-token" "true"

echo "All tests completed successfully! 🎉"

