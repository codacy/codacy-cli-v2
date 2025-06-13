#!/bin/bash

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
