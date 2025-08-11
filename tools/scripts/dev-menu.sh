#!/usr/bin/env bash

# Interactive Development Menu
# Dynamically parses Makefile targets and presents them with fzf

set -euo pipefail

# Colors and icons
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Function to get appropriate icon and color for each target
get_target_icon() {
    local target="$1"
    case "$target" in
        build)          echo "🏗️" ;;
        clean)          echo "🧹" ;;
        test)           echo "🧪" ;;
        lint)           echo "✨" ;;
        generate)       echo "⚙️" ;;
        proto)          echo "📦" ;;
        sqlc)           echo "🗄️" ;;
        mocks)          echo "🎭" ;;
        db-migrate*)    echo "🔄" ;;
        up)             echo "🐳" ;;
        down)           echo "🛑" ;;
        run)            echo "🚀" ;;
        status)         echo "📋" ;;
        help)           echo "❓" ;;
        all)            echo "🎯" ;;
        *)              echo "⚡" ;;
    esac
}

get_target_color() {
    local target="$1"
    case "$target" in
        build|all)      echo "$GREEN" ;;
        clean)          echo "$RED" ;;
        test)           echo "$YELLOW" ;;
        lint)           echo "$PURPLE" ;;
        generate|proto|sqlc|mocks) echo "$BLUE" ;;
        db-*)           echo "$CYAN" ;;
        up|down|run)    echo "$GREEN" ;;
        status|help)    echo "$WHITE" ;;
        *)              echo "$NC" ;;
    esac
}

# Function to parse Makefile and extract targets with descriptions
parse_makefile() {
    local makefile="${1:-Makefile}"
    
    if [[ ! -f "$makefile" ]]; then
        echo "Error: Makefile not found at $makefile" >&2
        exit 1
    fi
    
    # Parse Makefile to extract targets and descriptions
    awk '
    /^##/ { 
        # Store the comment
        desc = substr($0, 4)  # Remove "## "
        getline
        # Check if next line is a target (contains :)
        if ($0 ~ /^[a-zA-Z0-9_-]+:/) {
            target = $1
            gsub(/:.*/, "", target)  # Remove everything after :
            print target ":" desc
        }
    }
    ' "$makefile"
}

# Function to display the menu
show_menu() {
    local makefile="${1:-Makefile}"
    
    echo -e "${CYAN}🚀 Interactive Development Menu${NC}"
    echo -e "${WHITE}=================================${NC}"
    echo ""
    
    # Get all targets with descriptions
    local targets
    targets=$(parse_makefile "$makefile")
    
    if [[ -z "$targets" ]]; then
        echo -e "${RED}No targets found in $makefile${NC}"
        exit 1
    fi
    
    # Create formatted menu options
    local formatted_targets=""
    while IFS=':' read -r target desc; do
        local icon
        local color
        icon=$(get_target_icon "$target")
        color=$(get_target_color "$target")
        
        # Format: "icon target - description"
        formatted_targets+="${icon} ${color}${target}${NC} - ${desc}\n"
    done <<< "$targets"
    
    # Use fzf to display the menu
    local selected
    selected=$(echo -e "$formatted_targets" | fzf \
        --ansi \
        --height=80% \
        --border=rounded \
        --margin=1 \
        --padding=1 \
        --prompt="Select a command: " \
        --header="Use ↑↓ to navigate, Enter to select, Esc to quit" \
        --preview-window=right:50% \
        --preview='echo -e "Command: {2}\n\nDescription: {4..}" | fold -w 40' \
        --color="header:italic:white,prompt:bold:cyan,pointer:bold:green" \
    ) || {
        echo -e "\n${YELLOW}Menu cancelled${NC}"
        exit 0
    }
    
    # Extract the target name from the selected option
    local target_name
    target_name=$(echo "$selected" | awk '{print $2}' | sed 's/\x1b\[[0-9;]*m//g')
    
    if [[ -n "$target_name" ]]; then
        echo -e "\n${GREEN}Running: make ${target_name}${NC}"
        echo -e "${CYAN}=================================${NC}"
        
        # Execute the make command
        make "$target_name"
        
        local exit_code=$?
        echo -e "\n${CYAN}=================================${NC}"
        
        if [[ $exit_code -eq 0 ]]; then
            echo -e "${GREEN}✅ Command completed successfully${NC}"
        else
            echo -e "${RED}❌ Command failed with exit code $exit_code${NC}"
        fi
    fi
}

# Function to show help
show_help() {
    echo -e "${CYAN}Interactive Development Menu${NC}"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -f, --file FILE    Use specified Makefile (default: Makefile)"
    echo "  -h, --help         Show this help message"
    echo ""
    echo "Features:"
    echo "  • Dynamic target parsing from Makefile"
    echo "  • Fuzzy search with fzf"
    echo "  • Visual icons and colors for different target types"
    echo "  • Auto-updates when new targets are added to Makefile"
    echo ""
    echo "Adding new targets:"
    echo "  Add a comment above your target in the Makefile:"
    echo "  ## Your description here"
    echo "  your-target:"
    echo "      @echo \"Your command\""
}

# Check if fzf is installed
check_dependencies() {
    if ! command -v fzf &> /dev/null; then
        echo -e "${RED}Error: fzf is not installed${NC}" >&2
        echo -e "${YELLOW}Please install fzf:${NC}" >&2
        echo "  • macOS: brew install fzf" >&2
        echo "  • Ubuntu/Debian: apt install fzf" >&2
        echo "  • Or visit: https://github.com/junegunn/fzf" >&2
        exit 1
    fi
}

# Main function
main() {
    local makefile="Makefile"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--file)
                makefile="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}" >&2
                show_help
                exit 1
                ;;
        esac
    done
    
    # Check dependencies
    check_dependencies
    
    # Show the menu
    show_menu "$makefile"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi