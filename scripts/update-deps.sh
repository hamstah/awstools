#!/usr/bin/env bash

# Script to update Go dependencies in all modules
# Finds all directories containing go.mod files and runs 'go get -u' and 'go mod tidy' in each

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Get the script directory (where this script is located)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the project root (parent of scripts directory)
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

print_status "$BLUE" "🚀 Starting Go dependency updates..."
print_status "$BLUE" "Project root: $PROJECT_ROOT"
echo

# Find all go.mod files recursively (including nested subdirectories) and get their directories
GO_MOD_DIRS=()
while IFS= read -r gomod_file; do
    if [[ -n "$gomod_file" ]]; then
        dir=$(dirname "$gomod_file")
        GO_MOD_DIRS+=("$dir")
    fi
done < <(find "$PROJECT_ROOT" -type f -name "go.mod" | sort)

if [ ${#GO_MOD_DIRS[@]} -eq 0 ]; then
    print_status "$YELLOW" "⚠️  No go.mod files found in project"
    exit 0
fi

print_status "$GREEN" "Found ${#GO_MOD_DIRS[@]} Go modules:"
for dir in "${GO_MOD_DIRS[@]}"; do
    rel_dir="${dir#$PROJECT_ROOT/}"
    echo "  - $rel_dir"
done
echo

# Update dependencies in each module
FAILED_MODULES=()
UPDATED_MODULES=()

for dir in "${GO_MOD_DIRS[@]}"; do
    rel_dir="${dir#$PROJECT_ROOT/}"
    print_status "$BLUE" "📦 Updating dependencies in: $rel_dir"
    
    if cd "$dir" && go get -u && go mod tidy; then
        print_status "$GREEN" "✅ Successfully updated: $rel_dir"
        UPDATED_MODULES+=("$rel_dir")
    else
        print_status "$RED" "❌ Failed to update: $rel_dir"
        FAILED_MODULES+=("$rel_dir")
    fi
    echo
done

# Summary
echo "================================="
print_status "$BLUE" "📊 Update Summary"
echo "================================="

if [ ${#UPDATED_MODULES[@]} -gt 0 ]; then
    print_status "$GREEN" "✅ Successfully updated (${#UPDATED_MODULES[@]}):"
    for module in "${UPDATED_MODULES[@]}"; do
        echo "  - $module"
    done
fi

if [ ${#FAILED_MODULES[@]} -gt 0 ]; then
    echo
    print_status "$RED" "❌ Failed to update (${#FAILED_MODULES[@]}):"
    for module in "${FAILED_MODULES[@]}"; do
        echo "  - $module"
    done
    echo
    print_status "$RED" "Some modules failed to update. Please check the errors above."
    exit 1
fi

print_status "$GREEN" "🎉 All Go modules updated successfully!"
