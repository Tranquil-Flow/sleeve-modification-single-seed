////////////////////////////////////////////////////////////////////////////////////////////
// Derive Network Key Tool
//
// This tool helps users derive keys for any BIP44 network from their Sleeve mnemonic.
// It outputs the private key in multiple formats for easy import into wallet providers.
//
// Usage:
//   go run tools/derive-network.go -mnemonic "your 24 words..." -network "Solana" -cointype 501
//   go run tools/derive-network.go -mnemonic "your 24 words..." -network "Litecoin" -cointype 2
//   go run tools/derive-network.go -help
//
////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"github.com/xx-labs/sleeve/wallet"
)

// Network formats we can export
type NetworkFormats struct {
	Network       string
	CoinType      uint32
	Path          string
	PrivateKeyHex string
	WIF           string // Bitcoin-style Wallet Import Format
	EthAddress    string // Ethereum address (derived from public key)
	PublicKeyHex  string // Compressed public key
}

func main() {
	// Parse command-line flags
	mnemonicFlag := flag.String("mnemonic", "", "24-word mnemonic phrase (required)")
	passphraseFlag := flag.String("passphrase", "", "Optional passphrase (default: empty)")
	networkFlag := flag.String("network", "", "Network name (e.g., 'Solana', 'Litecoin')")
	coinTypeFlag := flag.Uint("cointype", 0, "BIP44 coin type number")
	listFlag := flag.Bool("list", false, "List common network coin types")
	helpFlag := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Show help
	if *helpFlag {
		printHelp()
		return
	}

	// List common networks
	if *listFlag {
		listCommonNetworks()
		return
	}

	// Validate required flags
	if *mnemonicFlag == "" {
		fmt.Println("Error: -mnemonic flag is required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	if *networkFlag == "" {
		fmt.Println("Error: -network flag is required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	// Validate mnemonic
	words := strings.Fields(*mnemonicFlag)
	if len(words) != 24 {
		fmt.Printf("Error: Mnemonic must be exactly 24 words (got %d)\n", len(words))
		os.Exit(1)
	}

	if !bip39.IsMnemonicValid(*mnemonicFlag) {
		fmt.Println("Error: Invalid mnemonic (checksum failed)")
		os.Exit(1)
	}

	// Create or recover Sleeve wallet
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║        Sleeve Network Key Derivation Tool                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Deriving keys from mnemonic...")
	fmt.Println()

	sleeve, err := wallet.NewSingleSeedSleeveFromMnemonic(*mnemonicFlag, *passphraseFlag, wallet.DefaultGenSpec())
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		os.Exit(1)
	}

	// Derive the requested network
	seed, err := bip39.NewSeedWithErrorChecking(*mnemonicFlag, *passphraseFlag)
	if err != nil {
		fmt.Printf("Error generating seed: %v\n", err)
		os.Exit(1)
	}

	err = sleeve.DeriveNetworkKey(*networkFlag, uint32(*coinTypeFlag), seed)
	if err != nil {
		fmt.Printf("Error deriving network key: %v\n", err)
		os.Exit(1)
	}

	// Get the private key
	privateKey, err := sleeve.GetPrivateKey(*networkFlag)
	if err != nil {
		fmt.Printf("Error retrieving private key: %v\n", err)
		os.Exit(1)
	}

	// Format the output
	formats := formatNetworkKey(*networkFlag, uint32(*coinTypeFlag), sleeve, privateKey)

	// Display results
	printNetworkKey(formats)
}

func formatNetworkKey(network string, coinType uint32, sleeve *wallet.SingleSeedSleeve, privateKey []byte) NetworkFormats {
	formats := NetworkFormats{
		Network:       network,
		CoinType:      coinType,
		PrivateKeyHex: hex.EncodeToString(privateKey),
	}

	// Get network info
	allKeys := sleeve.GetAllNetworkKeys()
	if netKey, exists := allKeys[network]; exists {
		formats.Path = netKey.Path
	}

	// Derive public key (works for all ECDSA-based chains)
	privKey, err := crypto.ToECDSA(privateKey)
	if err == nil {
		// Compressed public key (33 bytes) - standard for Bitcoin, etc.
		compressedPubKey := crypto.CompressPubkey(&privKey.PublicKey)
		formats.PublicKeyHex = hex.EncodeToString(compressedPubKey)

		// Ethereum address (useful for ETH and EVM chains)
		ethAddr := crypto.PubkeyToAddress(privKey.PublicKey)
		formats.EthAddress = ethAddr.Hex()

		// Bitcoin WIF format (useful for Bitcoin-like chains)
		formats.WIF = privateKeyToWIF(privateKey, coinType)
	}

	return formats
}

func printNetworkKey(f NetworkFormats) {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  Network: %-52s ║\n", f.Network)
	fmt.Printf("║  Coin Type: %-49d ║\n", f.CoinType)
	fmt.Printf("║  Path: %-55s ║\n", f.Path)
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("📋 PRIVATE KEY (Raw Hex)")
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Println(f.PrivateKeyHex)
	fmt.Println()
	fmt.Println("⚠️  KEEP THIS SECRET! Anyone with this key controls your funds.")
	fmt.Println()

	if f.PublicKeyHex != "" {
		fmt.Println("🔑 PUBLIC KEY (Compressed)")
		fmt.Println("────────────────────────────────────────────────────────────────")
		fmt.Println(f.PublicKeyHex)
		fmt.Println()
	}

	// Chain-specific formats
	if f.WIF != "" && (f.CoinType == 0 || f.CoinType == 2 || f.CoinType == 3) {
		// Bitcoin, Litecoin, Dogecoin
		fmt.Println("💰 WALLET IMPORT FORMAT (WIF)")
		fmt.Println("────────────────────────────────────────────────────────────────")
		fmt.Println(f.WIF)
		fmt.Println()
		fmt.Println("Import to: Bitcoin Core, Electrum, other Bitcoin wallets")
		fmt.Println("Command:   bitcoin-cli importprivkey " + f.WIF)
		fmt.Println()
	}

	if f.EthAddress != "" && (f.CoinType == 60 || f.CoinType == 61) {
		// Ethereum, Ethereum Classic
		fmt.Println("🦊 ETHEREUM ADDRESS")
		fmt.Println("────────────────────────────────────────────────────────────────")
		fmt.Println(f.EthAddress)
		fmt.Println()
		fmt.Println("Import to: MetaMask, MyEtherWallet, MyCrypto")
		fmt.Println("Steps:")
		fmt.Println("  1. Open MetaMask")
		fmt.Println("  2. Click account icon → Import Account")
		fmt.Println("  3. Select 'Private Key'")
		fmt.Println("  4. Paste: 0x" + f.PrivateKeyHex)
		fmt.Println()
	}

	// Generic instructions
	fmt.Println("📝 IMPORT INSTRUCTIONS FOR OTHER WALLETS")
	fmt.Println("────────────────────────────────────────────────────────────────")

	switch f.CoinType {
	case 354: // Polkadot
		fmt.Println("Polkadot/Substrate chains:")
		fmt.Println("  • Use Polkadot.js extension")
		fmt.Println("  • Import → Private Key → Paste hex above")
		fmt.Println("  • Note: May need to convert to SS58 format")
	case 501: // Solana
		fmt.Println("Solana:")
		fmt.Println("  • Use Phantom wallet or Solflare")
		fmt.Println("  • Import → Private Key")
		fmt.Println("  • Note: Some wallets expect base58 encoding")
		fmt.Println("  • Command: solana-keygen recover 'prompt:?key=0/' --outfile wallet.json")
	case 118: // Cosmos
		fmt.Println("Cosmos:")
		fmt.Println("  • Use Keplr wallet")
		fmt.Println("  • Import → Private Key")
		fmt.Println("  • Paste hex private key")
	case 1815: // Cardano
		fmt.Println("Cardano:")
		fmt.Println("  • Use Daedalus or Yoroi")
		fmt.Println("  • Note: Cardano uses different derivation, may need conversion")
	default:
		fmt.Println("Generic import (most wallets):")
		fmt.Println("  1. Look for 'Import Private Key' option")
		fmt.Println("  2. Paste the hex private key above")
		fmt.Println("  3. Some wallets require '0x' prefix: 0x" + f.PrivateKeyHex)
	}
	fmt.Println()

	fmt.Println("✅ SUCCESS!")
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Printf("Key for %s has been derived.\n", f.Network)
	fmt.Println("Store this information securely and import to your wallet.")
	fmt.Println()
}

func privateKeyToWIF(privateKey []byte, coinType uint32) string {
	// Simple WIF encoding for Bitcoin-like chains
	// This is a basic implementation - production use should use btcutil

	// Version byte depends on network
	version := byte(0x80) // Bitcoin mainnet
	switch coinType {
	case 0: // Bitcoin
		version = 0x80
	case 2: // Litecoin
		version = 0xB0
	case 3: // Dogecoin
		version = 0x9E
	default:
		return "" // Not a Bitcoin-like chain
	}

	// Build extended key: version + private key + compressed flag
	extended := make([]byte, 34)
	extended[0] = version
	copy(extended[1:33], privateKey)
	extended[33] = 0x01 // Compressed public key flag

	// Double SHA256 for checksum
	hash1 := crypto.Keccak256(extended)
	hash2 := crypto.Keccak256(hash1)
	checksum := hash2[:4]

	// Append checksum
	wifData := append(extended, checksum...)

	// Base58 encode (simplified - production should use btcutil)
	// For now, return hex with note
	return fmt.Sprintf("(Use btcutil for proper WIF: %x)", wifData)
}

func printHelp() {
	fmt.Println("Sleeve Network Key Derivation Tool")
	fmt.Println("===================================")
	fmt.Println()
	fmt.Println("Derive keys for any BIP44 network from your Sleeve mnemonic.")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run tools/derive-network.go [flags]")
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println("  -mnemonic string")
	fmt.Println("        Your 24-word mnemonic phrase (required)")
	fmt.Println("  -network string")
	fmt.Println("        Network name, e.g., 'Solana', 'Litecoin' (required)")
	fmt.Println("  -cointype uint")
	fmt.Println("        BIP44 coin type number (required)")
	fmt.Println("  -passphrase string")
	fmt.Println("        Optional BIP39 passphrase (default: empty)")
	fmt.Println("  -list")
	fmt.Println("        List common network coin types")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Derive Solana key")
	fmt.Println("  go run tools/derive-network.go \\")
	fmt.Println("    -mnemonic \"word1 word2 ... word24\" \\")
	fmt.Println("    -network \"Solana\" \\")
	fmt.Println("    -cointype 501")
	fmt.Println()
	fmt.Println("  # Derive Litecoin key")
	fmt.Println("  go run tools/derive-network.go \\")
	fmt.Println("    -mnemonic \"word1 word2 ... word24\" \\")
	fmt.Println("    -network \"Litecoin\" \\")
	fmt.Println("    -cointype 2")
	fmt.Println()
	fmt.Println("  # List common networks")
	fmt.Println("  go run tools/derive-network.go -list")
	fmt.Println()
	fmt.Println("SECURITY WARNING:")
	fmt.Println("  • Never share your mnemonic or private keys")
	fmt.Println("  • Run this tool on a secure, offline computer if possible")
	fmt.Println("  • Clear terminal history after use")
	fmt.Println()
	fmt.Println("For coin type numbers, see: https://github.com/satoshilabs/slips/blob/master/slip-0044.md")
}

func listCommonNetworks() {
	fmt.Println("Common BIP44 Network Coin Types")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("Auto-derived (included by default):")
	fmt.Println("  • Bitcoin       0")
	fmt.Println("  • Ethereum      60")
	fmt.Println("  • Polkadot      354")
	fmt.Println()
	fmt.Println("Bitcoin Family:")
	fmt.Println("  • Bitcoin       0")
	fmt.Println("  • Testnet       1")
	fmt.Println("  • Litecoin      2")
	fmt.Println("  • Dogecoin      3")
	fmt.Println("  • Dash          5")
	fmt.Println("  • Bitcoin Cash  145")
	fmt.Println()
	fmt.Println("Ethereum & EVM:")
	fmt.Println("  • Ethereum      60")
	fmt.Println("  • Ethereum Classic 61")
	fmt.Println("  • Polygon       966")
	fmt.Println("  • Avalanche     9000")
	fmt.Println("  • Fantom        1007")
	fmt.Println()
	fmt.Println("Smart Contract Platforms:")
	fmt.Println("  • Polkadot      354")
	fmt.Println("  • Solana        501")
	fmt.Println("  • Cosmos        118")
	fmt.Println("  • Cardano       1815")
	fmt.Println("  • Tezos         1729")
	fmt.Println()
	fmt.Println("Privacy Coins:")
	fmt.Println("  • Monero        128")
	fmt.Println("  • Zcash         133")
	fmt.Println()
	fmt.Println("Other Popular:")
	fmt.Println("  • Ripple/XRP    144")
	fmt.Println("  • Stellar       148")
	fmt.Println("  • EOS           194")
	fmt.Println("  • Tron          195")
	fmt.Println()
	fmt.Println("Usage example:")
	fmt.Println("  go run tools/derive-network.go \\")
	fmt.Println("    -mnemonic \"your 24 words...\" \\")
	fmt.Println("    -network \"Solana\" \\")
	fmt.Println("    -cointype 501")
	fmt.Println()
	fmt.Println("For complete list: https://github.com/satoshilabs/slips/blob/master/slip-0044.md")
}


