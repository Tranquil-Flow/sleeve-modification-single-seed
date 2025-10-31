///////////////////////////////////////////////////////////////////////////////
// Comprehensive Sleeve Wallet Generator
// Supports: Single-seed and Dual-seed modes
// Exports: Private keys and addresses for Ethereum, Polkadot, Bitcoin
///////////////////////////////////////////////////////////////////////////////

package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/xx-labs/sleeve/wallet"
	"github.com/xx-labs/sleeve/wots"
)

const banner = `
╔═══════════════════════════════════════════════════════════════════╗
║                  Sleeve Wallet Generator                          ║
║               Quantum-Secure Wallet Creation                      ║
╚═══════════════════════════════════════════════════════════════════╝
`

type Config struct {
	Mode       string // "single" or "dual"
	Mnemonic   string // Existing mnemonic (for recovery)
	Passphrase string // Optional BIP39 passphrase
	Account    uint32 // Account number
	Security   string // WOTS+ security level
	Export     bool   // Export private keys
}

func main() {
	// Parse command line flags
	cfg := parseFlags()

	// Display banner
	fmt.Print(banner)

	// Generate or recover wallet
	if cfg.Mnemonic == "" {
		fmt.Println("🔐 Generating NEW wallet...")
		fmt.Println()
	} else {
		fmt.Println("🔄 Recovering wallet from mnemonic...")
		fmt.Println()
	}

	// Generate based on mode
	if cfg.Mode == "single" {
		generateSingleSeed(cfg)
	} else {
		generateDualSeed(cfg)
	}
}

func parseFlags() Config {
	mode := flag.String("mode", "single", "Wallet mode: 'single' or 'dual'")
	mnemonic := flag.String("mnemonic", "", "Existing mnemonic (for recovery)")
	passphrase := flag.String("passphrase", "", "BIP39 passphrase (optional)")
	account := flag.Uint("account", 0, "Account number")
	security := flag.String("security", "level0", "WOTS+ security: level0-3")
	export := flag.Bool("export", true, "Export private keys for other chains")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Sleeve Wallet Generator\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate new single-seed wallet:\n")
		fmt.Fprintf(os.Stderr, "  %s -mode single\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Generate dual-seed wallet:\n")
		fmt.Fprintf(os.Stderr, "  %s -mode dual\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Recover from mnemonic:\n")
		fmt.Fprintf(os.Stderr, "  %s -mode single -mnemonic \"your 24 words\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # With passphrase:\n")
		fmt.Fprintf(os.Stderr, "  %s -mode single -passphrase \"secret\"\n\n", os.Args[0])
	}

	flag.Parse()

	return Config{
		Mode:       *mode,
		Mnemonic:   *mnemonic,
		Passphrase: *passphrase,
		Account:    uint32(*account),
		Security:   *security,
		Export:     *export,
	}
}

func generateSingleSeed(cfg Config) {
	// Parse security level
	secLevel := parseSecurityLevel(cfg.Security)
	spec := wallet.NewGenSpec(cfg.Account, secLevel)

	// Generate or recover
	var sleeve *wallet.SingleSeedSleeve
	var err error

	if cfg.Mnemonic == "" {
		sleeve, err = wallet.NewSingleSeedSleeve(rand.Reader, cfg.Passphrase, spec)
	} else {
		sleeve, err = wallet.NewSingleSeedSleeveFromMnemonic(cfg.Mnemonic, cfg.Passphrase, spec)
	}

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Display wallet info
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                    SINGLE-SEED WALLET")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Recovery phrase
	fmt.Println("🔑 RECOVERY PHRASE (24 words):")
	fmt.Println("   ⚠️  BACKUP THIS SECURELY - This is your ONLY backup!")
	fmt.Println()
	fmt.Printf("   %s\n", sleeve.GetMnemonic())
	fmt.Println()

	if cfg.Passphrase != "" {
		fmt.Printf("🔐 Passphrase: %s\n", cfg.Passphrase)
		fmt.Println("   ⚠️  You need BOTH the phrase AND passphrase to recover!")
		fmt.Println()
	}

	// WOTS+ info
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println("🛡️  QUANTUM SECURITY (WOTS+)")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Printf("   Public Key: %s\n", hex.EncodeToString(sleeve.GetWOTSPublicKey()))
	fmt.Printf("   Index:      %d\n", sleeve.GetDerivationIndex())
	fmt.Println()

	// Network keys
	if cfg.Export {
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println("                    NETWORK KEYS")
		fmt.Println("═══════════════════════════════════════════════════════════════")
		fmt.Println()

		exportNetworkKeys(sleeve)
	} else {
		fmt.Println("ℹ️  Use -export=true to show private keys for Ethereum, Bitcoin, etc.")
		fmt.Println()
	}

	// Instructions
	printInstructions(true)
}

func generateDualSeed(cfg Config) {
	// Parse security level
	secLevel := parseSecurityLevel(cfg.Security)
	spec := wallet.NewGenSpec(cfg.Account, secLevel)

	// Generate or recover
	var sleeve *wallet.Sleeve
	var err error

	if cfg.Mnemonic == "" {
		sleeve, err = wallet.NewSleeve(rand.Reader, cfg.Passphrase, spec)
	} else {
		sleeve, err = wallet.NewSleeveFromMnemonic(cfg.Mnemonic, cfg.Passphrase, spec)
	}

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Display wallet info
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                 DUAL-MNEMONIC WALLET (Legacy)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Recovery phrases
	fmt.Println("🔑 QUANTUM RECOVERY PHRASE (24 words):")
	fmt.Println("   ⚠️  BACKUP THIS SECURELY!")
	fmt.Println()
	fmt.Printf("   %s\n", sleeve.GetMnemonic())
	fmt.Println()

	fmt.Println("🔑 STANDARD RECOVERY PHRASE (24 words):")
	fmt.Println("   ⚠️  BACKUP THIS TOO!")
	fmt.Println()
	fmt.Printf("   %s\n", sleeve.GetOutputMnemonic())
	fmt.Println()

	if cfg.Passphrase != "" {
		fmt.Printf("🔐 Passphrase: %s\n", cfg.Passphrase)
		fmt.Println()
	}

	// Instructions for dual mode
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println("📝 USAGE:")
	fmt.Println("   • Quantum phrase: Keep safe for quantum security")
	fmt.Println("   • Standard phrase: Use in MetaMask, Trust Wallet, etc.")
	fmt.Println("   • Import standard phrase into any BIP39 wallet")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()

	printInstructions(false)
}

func exportNetworkKeys(sleeve *wallet.SingleSeedSleeve) {
	// Ethereum
	fmt.Println("🔷 ETHEREUM")
	fmt.Println("───────────────────────────────────────────────────────────────")
	ethKey, err := sleeve.GetPrivateKey("Ethereum")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		// Get Ethereum address
		privKey, _ := crypto.ToECDSA(ethKey)
		address := crypto.PubkeyToAddress(privKey.PublicKey)

		fmt.Printf("   Address:     %s\n", address.Hex())
		fmt.Printf("   Private Key: 0x%s\n", hex.EncodeToString(ethKey))
		networkKeys := sleeve.GetAllNetworkKeys()
		if netKey, ok := networkKeys["Ethereum"]; ok {
			fmt.Printf("   Path:        %s\n", netKey.Path)
		}
		fmt.Println()
		fmt.Println("   📱 To use in MetaMask:")
		fmt.Println("      1. Click account icon → 'Import Account'")
		fmt.Println("      2. Select 'Private Key'")
		fmt.Println("      3. Paste the private key above")
		fmt.Println()
	}

	// Bitcoin
	fmt.Println("🟠 BITCOIN")
	fmt.Println("───────────────────────────────────────────────────────────────")
	btcKey, err := sleeve.GetPrivateKey("Bitcoin")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Private Key: %s\n", hex.EncodeToString(btcKey))
		networkKeys := sleeve.GetAllNetworkKeys()
		if netKey, ok := networkKeys["Bitcoin"]; ok {
			fmt.Printf("   Path:        %s\n", netKey.Path)
		}
		fmt.Println()
		fmt.Println("   📱 To use:")
		fmt.Println("      • Convert to WIF format for Electrum/other wallets")
		fmt.Println("      • Or use Sleeve library to sign transactions")
		fmt.Println()
	}

	// Polkadot
	fmt.Println("🔴 POLKADOT")
	fmt.Println("───────────────────────────────────────────────────────────────")
	dotKey, err := sleeve.GetPrivateKey("Polkadot")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Private Key: %s\n", hex.EncodeToString(dotKey))
		networkKeys := sleeve.GetAllNetworkKeys()
		if netKey, ok := networkKeys["Polkadot"]; ok {
			fmt.Printf("   Path:        %s\n", netKey.Path)
		}
		fmt.Println()
		fmt.Println("   📱 To use:")
		fmt.Println("      • Import into Polkadot.js extension")
		fmt.Println("      • Or use SubWallet")
		fmt.Println()
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
}

func printInstructions(singleSeed bool) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("                    IMPORTANT NOTES")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	if singleSeed {
		fmt.Println("✅ SINGLE-SEED MODE:")
		fmt.Println("   • You only need to backup ONE mnemonic phrase")
		fmt.Println("   • All network keys are derived from this one phrase")
		fmt.Println("   • Quantum and classical keys are cryptographically bound")
		fmt.Println()
	}

	fmt.Println("🔐 SECURITY:")
	fmt.Println("   • Never share your mnemonic with anyone")
	fmt.Println("   • Never share your private keys")
	fmt.Println("   • Store backups in multiple secure locations")
	fmt.Println("   • Consider using a hardware wallet for large amounts")
	fmt.Println()

	fmt.Println("🔄 RECOVERY:")
	if singleSeed {
		fmt.Println("   • Run: go run generate-wallet.go -mode single -mnemonic \"your words\"")
	} else {
		fmt.Println("   • Run: go run generate-wallet.go -mode dual -mnemonic \"your quantum words\"")
	}
	fmt.Println("   • You'll get back the exact same keys")
	fmt.Println()

	fmt.Println("📚 MORE INFO:")
	fmt.Println("   • Technical docs: SINGLE_SEED.md")
	fmt.Println("   • Security analysis: SECURITY_ANALYSIS.md")
	fmt.Println("   • Export guide: EXPORT_KEYS.md")
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Println("⚠️  WARNING: Delete this output after securely storing your keys!")
	fmt.Println()
}

func parseSecurityLevel(level string) wots.ParamsEncoding {
	switch level {
	case "level0":
		return wots.Level0
	case "level1":
		return wots.Level1
	case "level2":
		return wots.Level2
	case "level3":
		return wots.Level3
	default:
		return wots.DefaultParams
	}
}
