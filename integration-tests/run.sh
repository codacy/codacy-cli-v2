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
  # Check for specific files first, then fall back to extension
  if [[ "$file" == *"languages-config.yaml" ]]; then
    normalize_languages_config "$file"
  else
    local ext="${file##*.}"
    
    case "$ext" in
      yaml|yml)
        normalize_yaml_config "$file"
        ;;
      mjs|js)
        normalize_eslint_config "$file"
        ;;
      toml)
        normalize_toml_config "$file"
        ;;
      rc|conf|ini)
        normalize_rc_config "$file"
        ;;
      xml)
        normalize_xml_config "$file"
        ;;
      *)
        # For other files, just sort
        sort "$file"
        ;;
    esac
  fi
}

# Normalize YAML configuration files
# Normalize languages-config.yaml specifically
normalize_languages_config() {
  local file=$1
  # Sort tools by name using yq if available
  if command -v yq >/dev/null 2>&1; then
    yq e '.tools |= sort_by(.name) | sort_keys(.)' "$file" 2>/dev/null || cat "$file"
  else
    # Fallback: manual sorting using awk
    awk '
      BEGIN { in_tools = 0; current_tool = ""; tools_count = 0; }
      /^tools:/ { 
        print; 
        in_tools = 1; 
        next 
      }
      in_tools && /^  - name:/ { 
        # Start of new tool - save previous if exists
        if (current_tool != "") {
          tool_names[tools_count] = tool_name
          tool_blocks[tools_count] = current_tool
          tools_count++
        }
        # Extract tool name
        tool_name = $0
        gsub(/^.*name: */, "", tool_name)
        current_tool = $0 "\n"
        next 
      }
      in_tools && /^    / { 
        # Part of current tool
        current_tool = current_tool $0 "\n"
        next 
      }
      in_tools && /^[^ ]/ { 
        # End of tools section
        if (current_tool != "") {
          tool_names[tools_count] = tool_name
          tool_blocks[tools_count] = current_tool
          tools_count++
        }
        # Sort and print tools
        for (i = 0; i < tools_count; i++) {
          for (j = i+1; j < tools_count; j++) {
            if (tool_names[i] > tool_names[j]) {
              # Swap names
              temp_name = tool_names[i]
              tool_names[i] = tool_names[j]
              tool_names[j] = temp_name
              # Swap blocks
              temp_block = tool_blocks[i]
              tool_blocks[i] = tool_blocks[j]
              tool_blocks[j] = temp_block
            }
          }
        }
        for (i = 0; i < tools_count; i++) {
          printf "%s", tool_blocks[i]
        }
        in_tools = 0
        print
        next
      }
      !in_tools { print }
      END {
        # Handle case where file ends while in tools section
        if (in_tools && current_tool != "") {
          tool_names[tools_count] = tool_name
          tool_blocks[tools_count] = current_tool
          tools_count++
          # Sort and print tools
          for (i = 0; i < tools_count; i++) {
            for (j = i+1; j < tools_count; j++) {
              if (tool_names[i] > tool_names[j]) {
                temp_name = tool_names[i]
                tool_names[i] = tool_names[j]
                tool_names[j] = temp_name
                temp_block = tool_blocks[i]
                tool_blocks[i] = tool_blocks[j]
                tool_blocks[j] = temp_block
              }
            }
          }
          for (i = 0; i < tools_count; i++) {
            printf "%s", tool_blocks[i]
          }
        }
      }
    ' "$file"
  fi
}

normalize_yaml_config() {
  local file=$1
  # For YAML files, use yq to sort while preserving structure
  if command -v yq >/dev/null 2>&1; then
    yq e 'sort_keys(.)' "$file" 2>/dev/null || cat "$file"
  else
    # Fallback: just return the file as-is to preserve YAML structure
    cat "$file"
  fi
}

# Normalize ESLint configuration files (.mjs/.js)
normalize_eslint_config() {
  local file=$1
  # Sort the rule lines within the rules object for consistent comparison
  awk '
    /rules: \{/ { 
      print; 
      inRules = 1; 
      next 
    }
    inRules && /^\s*\}/ { 
      # Print collected rules in sorted order using pipe to sort
      for (rule in rules) {
        print rules[rule] | "sort"
      }
      close("sort")
      delete rules
      inRules = 0
      print
      next
    }
    inRules { 
      # Collect rule lines for sorting
      rules[NR] = $0
      next 
    }
    { print }
  ' "$file"
}

# Normalize TOML configuration files
normalize_toml_config() {
  local file=$1
  # Handle TOML arrays and key-value pairs
  awk -F'=' '
    /^[^#].*=.*\[.*\]/ {
      # Handle TOML arrays like: rules = ["a", "b", "c"]
      # Extract array content using substr and index instead of match with array
      start = index($2, "[")
      end = index($2, "]")
      if (start > 0 && end > start) {
        array_content = substr($2, start+1, end-start-1)
        if (array_content) {
          # Split and sort values
          n = split(array_content, values, /,\s*/)
          # Sort using a simple bubble sort since asort is not available
          for (i = 1; i <= n; i++) {
            for (j = i+1; j <= n; j++) {
              if (values[i] > values[j]) {
                temp = values[i]
                values[i] = values[j]
                values[j] = temp
              }
            }
          }
          printf "%s=[", $1
          for (i=1; i<=n; i++) {
            if (i>1) printf ", "
            printf "%s", values[i]
          }
          print "]"
          next
        }
      }
    }
    /^[^#].*=.*,/ {
      # Handle simple comma-separated values
      n = split($2, values, ",")
      # Sort using bubble sort
      for (i = 1; i <= n; i++) {
        for (j = i+1; j <= n; j++) {
          if (values[i] > values[j]) {
            temp = values[i]
            values[i] = values[j]
            values[j] = temp
          }
        }
      }
      printf "%s=", $1
      for (i=1; i<=n; i++) {
        if (i>1) printf ","
        printf "%s", values[i]
      }
      print ""
      next
    }
    { print }
  ' "$file" | sort
}

# Normalize RC/INI configuration files
normalize_rc_config() {
  local file=$1
  # Handle key-value pairs with comma-separated values
  awk -F'=' '
    /^[^#].*=.*,/ {
      n = split($2, values, ",")
      # Sort using bubble sort
      for (i = 1; i <= n; i++) {
        for (j = i+1; j <= n; j++) {
          if (values[i] > values[j]) {
            temp = values[i]
            values[i] = values[j]
            values[j] = temp
          }
        }
      }
      printf "%s=", $1
      for (i=1; i<=n; i++) {
        if (i>1) printf ","
        printf "%s", values[i]
      }
      print ""
      next
    }
    { print }
  ' "$file" | sort
}

# Normalize XML configuration files
normalize_xml_config() {
  local file=$1
  # Sort rule blocks and properties within properties blocks
  awk '
    BEGIN { 
      rule_blocks_count = 0;
      single_rules_count = 0; 
      end = ""; 
      in_rule_block = 0;
      in_props = 0; 
      props_count = 0;
      current_rule_block = "";
      current_rule_ref = "";
    }
    /^ *<rule ref=.*>$/ { 
      # Start of a rule block with properties
      in_rule_block = 1;
      current_rule_ref = $0;
      gsub(/^.*ref="/, "", current_rule_ref);
      gsub(/".*$/, "", current_rule_ref);
      current_rule_block = $0 "\n";
      next 
    }
              /^ *<rule ref=.*\"\/>$/ { 
       # Self-closing rule (no properties)
       single_rules[++single_rules_count] = $0;
       next 
     }
    in_rule_block && /^ *<properties>/ { 
      in_props = 1; 
      props_start = $0;
      props_count = 0;
      next 
    }
    in_rule_block && in_props && /^ *<\/properties>/ { 
      in_props = 0; 
      # Sort and add collected properties to rule block
      current_rule_block = current_rule_block props_start "\n";
      for (i = 1; i <= props_count; i++) {
        for (j = i+1; j <= props_count; j++) {
          if (props[i] > props[j]) {
            temp = props[i]
            props[i] = props[j]
            props[j] = temp
          }
        }
      }
      for (i = 1; i <= props_count; i++) {
        current_rule_block = current_rule_block props[i] "\n";
      }
      current_rule_block = current_rule_block $0 "\n";
      props_count = 0;
      next 
    }
    in_rule_block && in_props && /^ *<property/ { 
      props[++props_count] = $0; 
      next 
    }
         in_rule_block && /^ *<\/rule>/ { 
       # End of rule block
       current_rule_block = current_rule_block $0;
       rule_block_refs[rule_blocks_count] = current_rule_ref;
       rule_blocks[rule_blocks_count] = current_rule_block;
       rule_blocks_count++;
       in_rule_block = 0;
       current_rule_block = "";
       current_rule_ref = "";
       next 
     }
    in_rule_block { 
      # Part of current rule block
      current_rule_block = current_rule_block $0 "\n";
      next 
    }
    /^ *<\/ruleset>/ { 
      end = $0; 
      next 
    }
    { 
      gsub(/^ +/, "", $0); 
      print 
    }
    END {
      # Sort rule blocks by reference
      for (i = 0; i < rule_blocks_count; i++) {
        for (j = i+1; j < rule_blocks_count; j++) {
          if (rule_block_refs[i] > rule_block_refs[j]) {
            temp_ref = rule_block_refs[i]
            rule_block_refs[i] = rule_block_refs[j]
            rule_block_refs[j] = temp_ref
            temp_block = rule_blocks[i]
            rule_blocks[i] = rule_blocks[j]
            rule_blocks[j] = temp_block
          }
        }
      }
             # Print sorted rule blocks
       for (i = 0; i < rule_blocks_count; i++) {
         printf "%s\n", rule_blocks[i]
       }
      # Sort and print single rules
      for (i = 1; i <= single_rules_count; i++) {
        for (j = i+1; j <= single_rules_count; j++) {
          if (single_rules[i] > single_rules[j]) {
            temp = single_rules[i]
            single_rules[i] = single_rules[j]
            single_rules[j] = temp
          }
        }
      }
      for (i = 1; i <= single_rules_count; i++) print single_rules[i];
      if (end) print end
    }
  ' "$file"
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

