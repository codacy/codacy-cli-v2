#!/usr/bin/env bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
}

run_command() {
    echo -e "${BLUE}Running: $*${NC}"
    "$@" > test_output.log 2>&1 || true
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo -e "${RED}Command failed with exit code $exit_code${NC}"
        echo -e "${RED}Output:${NC}"
    fi
    cat test_output.log
}

assert() {
    local test_name="$1"
    local condition="$2"
    local message="$3"

    if eval "$condition"; then
        echo -e "${GREEN}✓ $test_name${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ $test_name${NC}"
        echo -e "${RED}  $message${NC}"
        echo -e "${RED}  Debug output:${NC}"
        cat test_output.log
        ((TESTS_FAILED++))
    fi
}

cleanup() {
    echo -e "${BLUE}Cleaning up test environment...${NC}"
    rm -rf ~/Library/Caches/Codacy/codacy-cli-v2 .codacy || true
    unset CODACY_CLI_V2_VERSION
}

# Start testing
print_header "Starting Version Management Tests"

# Test 1: Clean environment init
print_header "Test 1: Clean Environment Init"
cleanup
run_command ./codacy-cli.sh init
assert "Init creates version.json" "[ -f .codacy/version.json ]" "version.json was not created"
INIT_VERSION=$(cat .codacy/version.json | grep -o '"version": *"[^"]*"' | cut -d'"' -f4 || echo "")
assert "Init fetches latest version" "[ -n \"$INIT_VERSION\" ]" "No version found in version.json"

# Test 2: Run with version.json present
print_header "Test 2: Run with version.json present"
run_command ./codacy-cli.sh analyze
assert "Uses version from version.json" "grep -q \"Using version from version.json: \" test_output.log" "Did not use version from version.json"

# Test 3: Run with environment variable
print_header "Test 3: Environment Variable Tests"
export CODACY_CLI_V2_VERSION=1.0.0-main.208.054c857
run_command ./codacy-cli.sh analyze
assert "Uses version from environment variable" "grep -q \"Using version from CODACY_CLI_V2_VERSION environment variable: 1.0.0-main.208.054c857\" test_output.log" "Did not use version from environment variable"

# Test 4: Update command with environment variable
print_header "Test 4: Update with Environment Variable"
run_command ./codacy-cli.sh update
assert "Shows environment variable warning" "grep -q \"WARNING: Latest version downloaded, but using version specified in CODACY_CLI_V2_VERSION environment variable: 1.0.0-main.208.054c857\" test_output.log" "Did not show environment variable warning"
assert "Uses environment variable version after update" "grep -q \"Successfully updated to version 1.0.0-main.208.054c857\" test_output.log" "Did not use environment variable version"

# Test 5: Update command without environment variable
print_header "Test 5: Update without Environment Variable"
unset CODACY_CLI_V2_VERSION
run_command ./codacy-cli.sh update
LATEST_VERSION=$(grep "Latest version is:" test_output.log | cut -d' ' -f4 || echo "")
assert "Fetches latest version" "[ -n \"$LATEST_VERSION\" ]" "Did not fetch latest version"
assert "Uses latest version after update" "grep -q \"Successfully updated to version $LATEST_VERSION\" test_output.log" "Did not use latest version"

# Test 6: Version.json content after update
print_header "Test 6: Version.json Content"
VERSION_IN_FILE=$(cat .codacy/version.json | grep -o '"version": *"[^"]*"' | cut -d'"' -f4 || echo "")
assert "version.json contains latest version" "[ \"$VERSION_IN_FILE\" = \"$LATEST_VERSION\" ]" "version.json does not contain latest version"

# Test 7: Missing version.json fallback
print_header "Test 7: Missing version.json Fallback"
rm -f .codacy/version.json || true
run_command ./codacy-cli.sh analyze
assert "Fetches latest version when version.json missing" "grep -q \"No version found in version.json, fetching latest version...\" test_output.log" "Did not detect missing version.json"

# Cleanup
cleanup
rm -f test_output.log || true

# Print test summary
print_header "Test Summary"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo -e "Total tests: $(($TESTS_PASSED + $TESTS_FAILED))"

# Exit with failure if any tests failed
[ $TESTS_FAILED -eq 0 ] || exit 1 