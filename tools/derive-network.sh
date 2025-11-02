#!/bin/bash
############################################################################################
# Sleeve Network Key Derivation Script
#
# Simple wrapper for deriving keys for any BIP44 network
############################################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Banner
echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║        Sleeve Network Key Derivation Tool                     ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Check for help flag
if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]] || [[ "$1" == "help" ]]; then
    cd "$PROJECT_ROOT" && go run tools/derive-network.go -help
    exit 0
fi

# Check for list flag
if [[ "$1" == "-l" ]] || [[ "$1" == "--list" ]] || [[ "$1" == "list" ]]; then
    cd "$PROJECT_ROOT" && go run tools/derive-network.go -list
    exit 0
fi

# Interactive mode if no arguments
if [ $# -eq 0 ]; then
    echo -e "${YELLOW}Interactive Mode${NC}"
    echo ""
    
    # Get mnemonic
    echo "Enter your 24-word mnemonic phrase:"
    read -r MNEMONIC
    
    # Get network name
    echo ""
    echo "Enter network name (e.g., Solana, Litecoin, Cosmos):"
    read -r NETWORK
    
    # Get coin type
    echo ""
    echo "Enter BIP44 coin type number:"
    echo "(Use 'derive-network.sh list' to see common coin types)"
    read -r COINTYPE
    
    # Optional passphrase
    echo ""
    echo "Enter passphrase (press Enter for none):"
    read -r -s PASSPHRASE
    
    # Build command
    CMD="cd $PROJECT_ROOT && go run tools/derive-network.go -mnemonic \"$MNEMONIC\" -network \"$NETWORK\" -cointype $COINTYPE"
    if [ -n "$PASSPHRASE" ]; then
        CMD="$CMD -passphrase \"$PASSPHRASE\""
    fi
    
    # Execute
    echo ""
    eval $CMD
    
else
    # Pass through all arguments
    cd "$PROJECT_ROOT" && go run tools/derive-network.go "$@"
fi

# Security reminder
echo ""
echo -e "${YELLOW}Security Reminder:${NC}"
echo "  • Clear your terminal history: history -c"
echo "  • Never share your mnemonic or private keys"
echo "  • Run on a secure, offline computer when possible"
echo ""



