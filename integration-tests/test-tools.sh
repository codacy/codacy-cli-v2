#!/bin/bash

# Function to cleanup all generated test files
cleanup_all_test_files() {
    echo "🧹 Cleaning up test files..."
    
    # Remove generated SARIF and sorted files from all tool test directories
    find plugins/tools/*/test/src -name "actual.sarif" -o -name "actual.sorted.json" -o -name "expected.sorted.json" -o -name "codacy.yaml.backup" | xargs rm -f 2>/dev/null || true
    
    # Restore original codacy.yaml files if they were modified
    git checkout -- plugins/tools/*/test/src/.codacy/codacy.yaml 2>/dev/null || true
    
    # Clean up any empty .codacy directories that were created during testing
    find plugins/tools/*/test/src -name ".codacy" -type d -empty | xargs rmdir 2>/dev/null || true
    
    echo "✅ Cleanup completed"
}

# Set up trap to ensure cleanup happens even if script fails
trap cleanup_all_test_files EXIT

# Initialize failed tools file
rm -f /tmp/failed_tools.txt
touch /tmp/failed_tools.txt

# Run tests for each tool directory
for tool_dir in plugins/tools/*/; do
    tool_name=$(basename "$tool_dir")
    if [ -d "$tool_dir/test/src" ]; then
        echo "Running tests for $tool_name..."
        ./run-tool-tests.sh "$tool_name" || {
            echo "❌ Test failed for $tool_name"
            echo "$tool_name" >>/tmp/failed_tools.txt
        }
    fi
done

# Check if any tools failed
if [ -s /tmp/failed_tools.txt ] && [ "$(wc -l </tmp/failed_tools.txt)" -gt 0 ]; then
    echo -e "\n❌ The following tools failed their tests:"
    cat /tmp/failed_tools.txt
    echo "::error::Some tool tests failed. Please check the logs above for details."
    exit 1
else
    echo "✅ All tool tests passed successfully!"
fi
