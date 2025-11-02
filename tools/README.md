# Sleeve Tools

Helper tools for working with Sleeve wallets.

---

## Network Key Derivation Tool

**Purpose:** Derive keys for any BIP44 network from your Sleeve mnemonic and get the output in wallet-ready formats.

### Why This Tool is Useful

When you create a Sleeve wallet, you get:
- ✅ **Automatic support** for Bitcoin, Ethereum, and Polkadot
- ✅ **Extensibility** for any other BIP44 network

But to use other networks (Solana, Cosmos, Litecoin, etc.), you need to:
1. Look up the coin type number
2. Derive the key
3. Format it for that chain's wallet

**This tool does all three steps for you!**

---

## Quick Start

### List Available Networks

```bash
# See coin type numbers for common networks
./tools/derive-network.sh list
```

Output:
```
Common BIP44 Network Coin Types
================================

Auto-derived (included by default):
  • Bitcoin       0
  • Ethereum      60
  • Polkadot      354

Bitcoin Family:
  • Bitcoin       0
  • Litecoin      2
  • Dogecoin      3
  • Dash          5
  ...
```

### Derive a Network Key

**Example 1: Derive Solana Key**

```bash
./tools/derive-network.sh \
  -mnemonic "your twenty four word mnemonic phrase here..." \
  -network "Solana" \
  -cointype 501
```

**Example 2: Derive Litecoin Key**

```bash
./tools/derive-network.sh \
  -mnemonic "your twenty four word mnemonic phrase here..." \
  -network "Litecoin" \
  -cointype 2
```

**Example 3: Interactive Mode**

```bash
./tools/derive-network.sh
# Then follow the prompts
```

---

## Output Format

The tool provides multiple formats for easy import:

```
╔════════════════════════════════════════════════════════════════╗
║  Network: Solana                                               ║
║  Coin Type: 501                                                ║
║  Path: m/44'/501'/0'/0/1847392011                             ║
╚════════════════════════════════════════════════════════════════╝

📋 PRIVATE KEY (Raw Hex)
────────────────────────────────────────────────────────────────
a1b2c3d4e5f6789...
⚠️  KEEP THIS SECRET! Anyone with this key controls your funds.

🔑 PUBLIC KEY (Compressed)
────────────────────────────────────────────────────────────────
02a1b2c3d4e5f6789...

📝 IMPORT INSTRUCTIONS FOR OTHER WALLETS
────────────────────────────────────────────────────────────────
Solana:
  • Use Phantom wallet or Solflare
  • Import → Private Key
  • Paste the hex key above
```

---

## Supported Network Formats

| Network | Output Includes | Import Instructions |
|---------|----------------|---------------------|
| **Bitcoin** | Raw hex, WIF, Public key | Bitcoin Core, Electrum |
| **Ethereum** | Raw hex, Address, Public key | MetaMask, MyEtherWallet |
| **Litecoin** | Raw hex, WIF, Public key | Litecoin Core, Electrum-LTC |
| **Solana** | Raw hex, Public key | Phantom, Solflare |
| **Cosmos** | Raw hex, Public key | Keplr |
| **Polkadot** | Raw hex, Public key | Polkadot.js |
| **Any other** | Raw hex, Public key | Generic instructions |

---

## Usage Scenarios

### Scenario 1: Adding Solana to Existing Wallet

You have a Sleeve wallet with Bitcoin, Ethereum, and Polkadot. Now you want Solana:

```bash
# 1. Look up Solana coin type (501)
./tools/derive-network.sh list

# 2. Derive Solana key
./tools/derive-network.sh \
  -mnemonic "your mnemonic..." \
  -network "Solana" \
  -cointype 501

# 3. Copy the hex private key from output

# 4. Import to Phantom wallet:
#    - Open Phantom
#    - Add Account → Import Private Key
#    - Paste the hex key
```

### Scenario 2: Multi-Chain Portfolio

You need keys for 10 different networks:

```bash
# Derive all networks you need
for network in Solana:501 Cosmos:118 Avalanche:9000 Polygon:966; do
  IFS=':' read -r name cointype <<< "$network"
  ./tools/derive-network.sh \
    -mnemonic "your mnemonic..." \
    -network "$name" \
    -cointype "$cointype"
done
```

### Scenario 3: Recovery on New Device

Your device is lost, but you have your mnemonic:

```bash
# Recover all your networks one by one
./tools/derive-network.sh -mnemonic "your backed up mnemonic..." -network "Bitcoin" -cointype 0
./tools/derive-network.sh -mnemonic "your backed up mnemonic..." -network "Ethereum" -cointype 60
./tools/derive-network.sh -mnemonic "your backed up mnemonic..." -network "Solana" -cointype 501
# ... etc.
```

---

## Command-Line Reference

### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `-mnemonic` | string | Yes | Your 24-word mnemonic phrase |
| `-network` | string | Yes | Network name (e.g., "Solana") |
| `-cointype` | uint | Yes | BIP44 coin type number |
| `-passphrase` | string | No | Optional BIP39 passphrase |
| `-list` | flag | No | Show common network coin types |
| `-help` | flag | No | Show help message |

### Examples

**Basic usage:**
```bash
./tools/derive-network.sh \
  -mnemonic "word1 word2 ... word24" \
  -network "NetworkName" \
  -cointype 123
```

**With passphrase:**
```bash
./tools/derive-network.sh \
  -mnemonic "word1 word2 ... word24" \
  -passphrase "mySecretPassphrase" \
  -network "NetworkName" \
  -cointype 123
```

**List networks:**
```bash
./tools/derive-network.sh -list
# or
./tools/derive-network.sh list
```

**Show help:**
```bash
./tools/derive-network.sh -help
# or
./tools/derive-network.sh --help
```

---

## Understanding the Output

### 1. Private Key (Raw Hex)

```
a1b2c3d4e5f6...
```

- **Format:** 64 hexadecimal characters (32 bytes)
- **Use:** Import directly into most wallets
- **Security:** Keep this SECRET! Anyone with this can steal your funds

### 2. Public Key (Compressed)

```
02a1b2c3d4e5f6...
```

- **Format:** 66 hex characters (33 bytes), starts with 02 or 03
- **Use:** Derive addresses, share safely (public)
- **Security:** Safe to share

### 3. Ethereum Address

```
0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
```

- **Format:** Checksummed Ethereum address
- **Use:** Your receiving address for ETH and ERC-20 tokens
- **Security:** Safe to share (public)

### 4. WIF (Wallet Import Format)

```
L3k2Jm...
```

- **Format:** Base58-encoded (starts with L, K, or 5)
- **Use:** Import into Bitcoin Core, Electrum
- **Security:** Keep SECRET! (contains private key)

---

## Security Best Practices

### ⚠️ CRITICAL SECURITY WARNINGS

1. **Never share your mnemonic**
   - Your 24-word phrase = complete control of all funds
   - Store it offline, in a secure location

2. **Never share private keys**
   - Anyone with a private key can steal those funds
   - Only import into wallets YOU control

3. **Run on secure computer**
   - Ideally offline (air-gapped) computer
   - No malware, no screen recorders

4. **Clear sensitive data**
   ```bash
   # Clear terminal history after use
   history -c
   
   # Or use history -d to remove specific commands
   history | grep mnemonic  # Find line numbers
   history -d 1234         # Delete specific line
   ```

5. **Verify addresses**
   - Always send a small test transaction first
   - Verify you can receive funds before sending large amounts

### Safe Workflow

```bash
# 1. Run on offline computer (best)
./tools/derive-network.sh ...

# 2. Write down/save the private key on paper or encrypted USB
# (NOT on cloud or computer)

# 3. Clear terminal
clear
history -c

# 4. Import private key to wallet on separate device

# 5. Verify by sending small test amount first
```

---

## Troubleshooting

### Error: "Invalid mnemonic (checksum failed)"

**Problem:** Your mnemonic has a typo or wrong word.

**Solution:**
- Check for typos
- Verify each word is in BIP39 wordlist
- Ensure exactly 24 words
- Check word order (order matters!)

### Error: "Go is not installed"

**Problem:** Go programming language not found.

**Solution:**
```bash
# macOS
brew install go

# Linux
sudo apt-get install golang

# Or download from https://golang.org/dl/
```

### Error: "Network X not found"

**Problem:** Network wasn't derived yet.

**Solution:** This shouldn't happen with this tool (it derives on-demand), but if it does, make sure you're using the correct network name and coin type.

### Output shows "Use btcutil for proper WIF"

**Problem:** WIF encoding requires additional library.

**Solution:** The raw hex key works in most wallets. For proper WIF format:
```bash
# Install btcutil
go get github.com/btcsuite/btcutil

# Or use the hex key directly (most wallets accept both)
```

---

## Technical Details

### What This Tool Does

1. **Loads your mnemonic** → Generates BIP39 seed
2. **Derives quantum path** → `m/44'/1955'/0'/0'/0'` (generates WOTS+ key)
3. **Calculates WOTS index** → First 31 bits of SHA3-256(WOTS_PK)
4. **Derives network path** → `m/44'/{cointype}'/0'/0/{wots_index}`
5. **Extracts private key** → 32-byte secp256k1 private key
6. **Formats for display** → Hex, WIF, addresses, etc.

### Path Structure

```
m/44'/1955'/0'/0'/0'              ← Quantum path (WOTS+)
         ↓
    WOTS Public Key
         ↓
    SHA3-256 → 31 bits → Index (e.g., 1847392011)
         ↓
m/44'/{coin}'/0'/0/1847392011     ← Network path
         ↓
    Network Private Key
```

### Compatibility

- **BIP39:** Mnemonic to seed conversion
- **BIP32:** Hierarchical deterministic derivation
- **BIP44:** Multi-account hierarchy
- **SLIP-0044:** Registered coin type numbers

All standard-compliant, works with any BIP44 wallet.

---

## Finding Coin Type Numbers

### Method 1: Use Built-In List

```bash
./tools/derive-network.sh -list
```

### Method 2: Check SLIP-0044 Registry

Official registry: https://github.com/satoshilabs/slips/blob/master/slip-0044.md

Search for your network name, find the number in the "Index" column.

### Method 3: Search Online

```
"[NetworkName] BIP44 coin type"
```

Example: "Solana BIP44 coin type" → Result: 501

---

## FAQ

**Q: Do I need to modify code to add a new network?**

A: No! Just run the tool with the new network's coin type.

**Q: Will my keys be the same as other HD wallets?**

A: The WOTS-derived index makes Sleeve keys unique. You cannot import a Sleeve-derived key into a standard HD wallet and get the same address (the paths are different).

**Q: Can I use the same mnemonic in MetaMask?**

A: MetaMask will use standard BIP44 paths (`m/44'/60'/0'/0/0`, `m/44'/60'/0'/0/1`, ...), which are different from Sleeve paths. Sleeve-derived keys won't match MetaMask addresses.

**Q: Is this secure?**

A: Yes, if you:
- Run on a secure computer
- Don't share your mnemonic/private keys
- Clear sensitive data from memory/history
- Verify wallet addresses before sending funds

**Q: Can I recover my keys if I lose this tool?**

A: Yes! As long as you have your 24-word mnemonic, you can always regenerate keys. The derivation is deterministic and follows the method described in the bounty report.

**Q: What if a network isn't in the list?**

A: Look up its BIP44 coin type in SLIP-0044 and use it with `-cointype`. The tool works with ANY coin type.

---

## Related Documentation

- **TECHNICAL_EXPLANATION.md** - Deep dive into BIP32/BIP44 implementation
- **SECURITY_SCENARIOS.md** - Security analysis and attack scenarios  
- **NETWORK_EXTENSIBILITY.md** - How network support works
- **ARCHITECTURE_CLARIFICATION.md** - Understanding the extensible design

---

## Contributing

Found a bug? Want to add better formatting for a specific network?

The tool is in `tools/derive-network.go`. Pull requests welcome!

Common improvements:
- Better WIF encoding (add btcutil)
- SS58 formatting for Substrate chains
- Base58 formatting for Solana
- Network-specific address derivation

---

## License

Same as parent project (see LICENSE file).



