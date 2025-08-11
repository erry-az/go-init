#!/usr/bin/env bash

# Debug version to test menu parsing
set -e

# Source the functions from the main script
my_dir=$(dirname "$(realpath $0)")
source "$my_dir/dev-menu.sh"

echo "=== Testing parse_makefile function ==="
parse_makefile

echo ""
echo "=== Testing menu building logic ==="

# Build menu items from Makefile + special commands
menu_items=""

# Add Makefile targets
parsed_output=$(parse_makefile)
echo "Parsed output:"
echo "$parsed_output"
echo ""

while IFS=':' read -r target description; do
    if [[ -n "$target" && -n "$description" ]]; then
        icon=$(get_target_icon "$target")
        menu_items+="$icon $target|$description\n"
        echo "Added: $icon $target|$description"
    fi
done <<< "$parsed_output"

echo ""
echo "=== Final menu items ==="
echo -e "$menu_items"