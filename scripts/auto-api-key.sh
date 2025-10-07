#!/bin/bash

# GudCommit Auto API Key Manager
# This script automatically generates and manages Bedrock API keys

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_SCRIPT="$SCRIPT_DIR/generate-api-key.py"
CREDENTIALS_FILE="$HOME/.gudcommit-credentials"
GUDCOMMIT_BINARY="$SCRIPT_DIR/../golang/bin/gudcommit"

echo -e "${BLUE}🔑 GudCommit Auto API Key Manager${NC}"
echo "=================================="
echo ""

# Function to check if API key is valid
check_api_key() {
    if [ -z "$GUD_BEDROCK_API_KEY" ]; then
        return 1
    fi
    
    # Check if credentials file exists and is recent (less than 11 hours old)
    if [ -f "$CREDENTIALS_FILE" ]; then
        # Get file modification time
        file_time=$(stat -f "%m" "$CREDENTIALS_FILE" 2>/dev/null || stat -c "%Y" "$CREDENTIALS_FILE" 2>/dev/null)
        current_time=$(date +%s)
        age_hours=$(( (current_time - file_time) / 3600 ))
        
        if [ $age_hours -lt 11 ]; then
            return 0
        fi
    fi
    
    return 1
}

# Function to generate new API key
generate_new_key() {
    echo -e "${YELLOW}🔄 Generating new API key...${NC}"
    
    # Check if Python is available
    if ! command -v python3 &> /dev/null; then
        echo -e "${RED}❌ Python 3 not found. Please install Python 3.${NC}"
        exit 1
    fi
    
    # Run the Python script
    if python3 "$PYTHON_SCRIPT"; then
        echo -e "${GREEN}✅ New API key generated successfully!${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to generate API key${NC}"
        return 1
    fi
}

# Function to load existing credentials
load_credentials() {
    if [ -f "$CREDENTIALS_FILE" ]; then
        echo -e "${BLUE}📁 Loading existing credentials...${NC}"
        source "$CREDENTIALS_FILE"
        return 0
    fi
    return 1
}

# Function to run gudcommit
run_gudcommit() {
    if [ ! -f "$GUDCOMMIT_BINARY" ]; then
        echo -e "${YELLOW}🔨 Building GudCommit...${NC}"
        cd "$SCRIPT_DIR/../golang" && make build
    fi
    
    echo -e "${GREEN}🚀 Running GudCommit...${NC}"
    "$GUDCOMMIT_BINARY"
}

# Function to run gudchangelog
run_gudchangelog() {
    local target_branch="$1"
    local gudchangelog_binary="$SCRIPT_DIR/../golang/bin/gudchangelog"
    
    if [ -z "$target_branch" ]; then
        echo -e "${RED}❌ Please specify target branch: ./scripts/auto-api-key.sh run gudchangelog <branch>${NC}"
        exit 1
    fi
    
    if [ ! -f "$gudchangelog_binary" ]; then
        echo -e "${YELLOW}🔨 Building GudChangelog...${NC}"
        cd "$SCRIPT_DIR/../golang" && make build
    fi
    
    echo -e "${GREEN}📋 Running GudChangelog against branch: $target_branch${NC}"
    "$gudchangelog_binary" "$target_branch"
}

# Main logic
main() {
    # Check if we have a valid API key
    if check_api_key; then
        echo -e "${GREEN}✅ Valid API key found${NC}"
        echo "Key: ${GUD_BEDROCK_API_KEY:0:10}..."
        echo ""
        
        # Ask if user wants to regenerate
        echo -e "${YELLOW}Do you want to generate a new API key? (y/n):${NC}"
        read -p "Regenerate: " regenerate
        
        if [[ "$regenerate" =~ ^[Yy]$ ]]; then
            generate_new_key
        else
            echo -e "${BLUE}Using existing API key${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  No valid API key found${NC}"
        echo ""
        
        # Try to load existing credentials first
        if load_credentials; then
            echo -e "${BLUE}📁 Loaded credentials from file${NC}"
        else
            echo -e "${YELLOW}🔄 Generating new API key...${NC}"
            generate_new_key
        fi
    fi
    
    # Load the credentials
    if [ -f "$CREDENTIALS_FILE" ]; then
        source "$CREDENTIALS_FILE"
    fi
    
    # Check if we have the API key set
    if [ -z "$GUD_BEDROCK_API_KEY" ]; then
        echo -e "${RED}❌ No API key available${NC}"
        exit 1
    fi
    
    echo ""
    echo -e "${GREEN}🎉 Ready to use GudCommit!${NC}"
    echo ""
    
    # Ask if user wants to run gudcommit
    echo -e "${YELLOW}Do you want to run GudCommit now? (y/n):${NC}"
    read -p "Run GudCommit: " run_now
    
    if [[ "$run_now" =~ ^[Yy]$ ]]; then
        run_gudcommit
    else
        echo ""
        echo -e "${BLUE}To use GudCommit manually:${NC}"
        echo "  source $CREDENTIALS_FILE"
        echo "  $GUDCOMMIT_BINARY"
    fi
}

# Handle command line arguments
case "${1:-}" in
    "generate"|"new"|"regenerate")
        generate_new_key
        ;;
    "run"|"gudcommit")
        # Check if gudchangelog is requested
        if [ "$2" = "gudchangelog" ]; then
            if check_api_key || load_credentials; then
                run_gudchangelog "$3"
            else
                echo -e "${YELLOW}No valid API key. Generating new one...${NC}"
                generate_new_key
                run_gudchangelog "$3"
            fi
        else
            if check_api_key || load_credentials; then
                run_gudcommit
            else
                echo -e "${YELLOW}No valid API key. Generating new one...${NC}"
                generate_new_key
                run_gudcommit
            fi
        fi
        ;;
    "status"|"check")
        if check_api_key; then
            echo -e "${GREEN}✅ API key is valid${NC}"
            echo "Key: ${GUD_BEDROCK_API_KEY:0:10}..."
        else
            echo -e "${RED}❌ No valid API key${NC}"
        fi
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  generate                  - Generate a new API key"
        echo "  run                       - Run GudCommit (auto-generates key if needed)"
        echo "  run gudchangelog <branch> - Run GudChangelog against target branch"
        echo "  status                    - Check API key status"
        echo "  help                      - Show this help"
        echo ""
        echo "Examples:"
        echo "  $0                        # Interactive mode"
        echo "  $0 generate               # Generate new API key"
        echo "  $0 run                    # Run GudCommit"
        echo "  $0 run gudchangelog main  # Run GudChangelog against main branch"
        echo "  $0 status                 # Check API key status"
        ;;
    "")
        main
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
