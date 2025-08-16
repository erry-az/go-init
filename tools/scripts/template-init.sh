#!/bin/bash

# Template Initialization Script
# Renames all module references and imports to make this repo a template

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
OLD_MODULE="github.com/erry-az/go-init"
NEW_MODULE=""
PROJECT_NAME=""

# Help function
show_help() {
    cat << EOF
Template Initialization Script

USAGE:
    $0 --module <new-module> [--name <project-name>]

OPTIONS:
    --module, -m    New Go module name (e.g., github.com/username/project)
    --name, -n      Project name for configs (defaults to last part of module)
    --help, -h      Show this help message

EXAMPLES:
    $0 --module github.com/myorg/backend-service
    $0 --module github.com/myorg/api --name my-api

This script will:
1. Update go.mod with new module name
2. Replace all internal imports across Go files
3. Update protobuf package names and imports
4. Update configuration files
5. Update Docker and deployment configs

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--module)
            NEW_MODULE="$2"
            shift 2
            ;;
        -n|--name)
            PROJECT_NAME="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$NEW_MODULE" ]]; then
    echo -e "${RED}Error: --module is required${NC}"
    show_help
    exit 1
fi

# Extract project name from module if not provided
if [[ -z "$PROJECT_NAME" ]]; then
    PROJECT_NAME=$(basename "$NEW_MODULE")
fi

# Validate module format
if [[ ! "$NEW_MODULE" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
    echo -e "${RED}Error: Invalid module name format${NC}"
    exit 1
fi

echo -e "${BLUE}🚀 Initializing template with:${NC}"
echo -e "  Old module: ${YELLOW}$OLD_MODULE${NC}"
echo -e "  New module: ${GREEN}$NEW_MODULE${NC}"
echo -e "  Project name: ${GREEN}$PROJECT_NAME${NC}"
echo

# Confirm with user
read -p "Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

echo -e "${BLUE}📝 Updating files...${NC}"

# Function to update files with sed (cross-platform)
update_file() {
    local file="$1"
    local old="$2"
    local new="$3"
    
    # Check if file exists
    if [[ ! -f "$file" ]]; then
        echo "Warning: File $file not found, skipping..."
        return 0
    fi
    
    # Use | as delimiter to avoid issues with forward slashes in URLs
    # Escape only the delimiter character in the strings
    local escaped_old=$(printf '%s\n' "$old" | sed 's/|/\\|/g')
    local escaped_new=$(printf '%s\n' "$new" | sed 's/|/\\|/g')
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|$escaped_old|$escaped_new|g" "$file"
    else
        # Linux
        sed -i "s|$escaped_old|$escaped_new|g" "$file"
    fi
}

# 1. Update go.mod
echo "  → Updating go.mod"
update_file "go.mod" "$OLD_MODULE" "$NEW_MODULE"

# 2. Update all Go files with imports
echo "  → Updating Go imports"
find . -name "*.go" -type f -exec grep -l "$OLD_MODULE" {} \; | while read -r file; do
    echo "    - $file"
    update_file "$file" "$OLD_MODULE" "$NEW_MODULE"
done

# 3. Update protobuf files
echo "  → Updating protobuf files"
find . -name "*.proto" -type f -exec grep -l "$OLD_MODULE" {} \; | while read -r file; do
    echo "    - $file"
    update_file "$file" "$OLD_MODULE" "$NEW_MODULE"
done

# 4. Update buf.yaml
if [[ -f "buf.yaml" ]]; then
    echo "  → Updating buf.yaml"
    update_file "buf.yaml" "$OLD_MODULE" "$NEW_MODULE"
fi

# 5. Update configuration files
echo "  → Updating configuration files"
find files/ -name "*.yaml" -o -name "*.yml" -o -name "*.json" 2>/dev/null | while read -r file; do
    if grep -q "go.init\|go_init" "$file" 2>/dev/null; then
        echo "    - $file"
        update_file "$file" "go.init" "$PROJECT_NAME"
        update_file "$file" "go_init" "${PROJECT_NAME//-/_}"
    fi
done

# 6. Update Docker files
echo "  → Updating Docker files"
if [[ -f "docker-compose.yml" ]]; then
    update_file "docker-compose.yml" "go_init" "${PROJECT_NAME//-/_}"
fi

# 7. Update Makefile if it contains project-specific references
if [[ -f "Makefile" ]] && grep -q "go.init\|go_init" "Makefile" 2>/dev/null; then
    echo "  → Updating Makefile"
    update_file "Makefile" "go_init" "${PROJECT_NAME//-/_}"
fi

# 8. Clean up generated files that might have old imports
echo -e "${BLUE}🧹 Cleaning generated files...${NC}"
make clean 2>/dev/null || echo "  (clean target not available)"

# 9. Regenerate code with new module
echo -e "${BLUE}🔄 Regenerating code...${NC}"
if command -v go &> /dev/null; then
    go mod tidy
    echo "  → go mod tidy completed"
fi

if make generate &> /dev/null; then
    echo "  → Code generation completed"
else
    echo -e "${YELLOW}  ⚠️  Could not run 'make generate' - run manually after setup${NC}"
fi

# 10. Update git remote (optional)
echo -e "${BLUE}🔗 Git remote update${NC}"
if git remote get-url origin &> /dev/null; then
    current_remote=$(git remote get-url origin)
    echo -e "  Current remote: ${YELLOW}$current_remote${NC}"
    echo -e "  ${YELLOW}Remember to update git remote to your new repository${NC}"
    echo -e "  Run: ${GREEN}git remote set-url origin <your-new-repo-url>${NC}"
fi

echo
echo -e "${GREEN}✅ Template initialization completed!${NC}"
echo
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Update git remote: git remote set-url origin <your-repo-url>"
echo "  2. Run tests: make test"
echo "  3. Start development: make dev"
echo "  4. Update README.md with your project details"
echo
echo -e "${GREEN}🎉 Happy coding!${NC}"