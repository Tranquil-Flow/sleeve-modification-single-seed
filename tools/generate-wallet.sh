#!/bin/bash
#
# Sleeve Wallet Generator - Easy Wrapper Script
#
# Usage:
#   ./generate-wallet.sh                    # Generate single-seed wallet
#   ./generate-wallet.sh --dual             # Generate dual-seed wallet
#   ./generate-wallet.sh --recover "words"  # Recover from mnemonic
#

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Banner
echo -e "${BLUE}"
cat << "EOF"
╔═══════════════════════════════════════════════════════════════════╗
║                  Sleeve Wallet Generator                          ║
║               Quantum-Secure Wallet Creation                      ║
╚═══════════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed!${NC}"
    echo "Please install Go from: https://go.dev/dl/"
    exit 1
fi

# Parse arguments
MODE="single"
MNEMONIC=""
PASSPHRASE=""
EXPORT="true"

while [[ $# -gt 0 ]]; do
    case $1 in
        --dual)
            MODE="dual"
            shift
            ;;
        --single)
            MODE="single"
            shift
            ;;
        --recover)
            MNEMONIC="$2"
            shift 2
            ;;
        --passphrase)
            PASSPHRASE="$2"
            shift 2
            ;;
        --no-export)
            EXPORT="false"
            shift
            ;;
        --help|-h)
            echo "Sleeve Wallet Generator"
            echo ""
            echo "Usage:"
            echo "  ./generate-wallet.sh [options]"
            echo ""
            echo "Options:"
            echo "  --single         Generate single-seed wallet (default)"
            echo "  --dual           Generate dual-seed wallet (legacy)"
            echo "  --recover \"...\" Recover from existing mnemonic"
            echo "  --passphrase \"...\" Add BIP39 passphrase"
            echo "  --no-export      Don't export private keys"
            echo "  --help           Show this help"
            echo ""
            echo "Examples:"
            echo "  # Generate new single-seed wallet:"
            echo "  ./generate-wallet.sh"
            echo ""
            echo "  # Generate dual-seed wallet:"
            echo "  ./generate-wallet.sh --dual"
            echo ""
            echo "  # Recover from mnemonic:"
            echo "  ./generate-wallet.sh --recover \"word1 word2 ... word24\""
            echo ""
            echo "  # With passphrase:"
            echo "  ./generate-wallet.sh --passphrase \"my secret\""
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Build args for Go program
ARGS="-mode $MODE -export=$EXPORT"

if [ -n "$MNEMONIC" ]; then
    ARGS="$ARGS -mnemonic \"$MNEMONIC\""
fi

if [ -n "$PASSPHRASE" ]; then
    ARGS="$ARGS -passphrase \"$PASSPHRASE\""
fi

# Run the Go program
echo -e "${YELLOW}⏳ Generating wallet...${NC}"
echo ""

eval "go run tools/generate-wallet.go $ARGS"

# Success message
echo -e "${GREEN}✅ Wallet generated successfully!${NC}"
echo ""
echo -e "${YELLOW}⚠️  Remember to:${NC}"
echo "   1. Backup your mnemonic phrase securely"
echo "   2. Never share it with anyone"
echo "   3. Store it in multiple safe locations"
if [ "$EXPORT" == "true" ]; then
    echo "   4. Clear your terminal history to remove private keys"
    echo "      (Run: history -c && history -w)"
fi
echo ""
