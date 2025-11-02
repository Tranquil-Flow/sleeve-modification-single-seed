## Sleeve

Sleeve is a novel way of embedding a quantum secure key in the
generation of curve based, non quantum secure keys.

A complete diagram of the Sleeve wallet generation can be found
in [docs](wallet/docs). This implementation of Sleeve uses a WOTS+
key as the underlying quantum secure key, and the diagram for
the generation of this key can also be found in [docs](wallet/docs).

## Generation Modes

Sleeve now supports **two generation modes**:

### 1. Dual-Mnemonic Mode (Legacy)

The original Sleeve implementation that generates two mnemonic phrases:
- **Quantum recovery phrase**: Used to generate WOTS+ quantum-secure keys
- **Standard recovery phrase**: Used for regular cryptocurrency wallets

The input for wallet generation is random entropy,
which is encoded into a mnemonic phrase using [BIP39](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki). Then a [BIP44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)
custom path of `m/44'/1955'/0'/0'/0'` is used in a [BIP32](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki) derivation
to generate a 256 bit child private key and 256 bit chain code.

The private key and chain code are used, respectively, as the
secret and public seeds in WOTS+ generation. After the WOTS+ key
is generated, we save the quantum secure public key (PK).
This public key will be used as the Wallet address when full
quantum secure capabilities are implemented in the future.

The Sleeve output is generated using SHA3_256 to hash the
sleeve secret key and WOTS+ public key together. The resulting
hash value is then encoded using BIP39, providing the output
mnemonic, which can be used to generate non quantum secure keys
on any blockchain platform.

### 2. Single-Seed Mode

An improved generation method that uses **only one mnemonic phrase** while maintaining quantum security.

#### Quick Start: Create a Quantum-Proof Ethereum Wallet

**Want to use MetaMask with quantum security? Here's how:**

1. **Generate wallet:**
   ```bash
   cd sleevage
   ./bin/sleevage-mac-m1 --single-seed
   ```

2. **Back up your 24-word mnemonic** (write it down!)

3. **Get your Ethereum key:**
   ```bash
   cd ..
   go run tools/derive-network.go \
     -mnemonic "your 24 words here..." \
     -network "Ethereum" \
     -cointype 60
   ```

4. **Import to MetaMask:**
   - Open MetaMask → Import Account → Private Key
   - Paste the private key from step 3

✅ **Done!** One mnemonic backs up your quantum-secure WOTS+ key AND your Ethereum wallet.

See **[SINGLE_SEED.md](SINGLE_SEED.md)** for complete step-by-step guide with screenshots.

#### Features

- ✅ **One mnemonic** backs up everything (quantum + classical keys)
- ✅ **Universal:** Works with Bitcoin, Ethereum, Polkadot, and all BIP44 networks
- ✅ **Quantum-secure:** WOTS+ key cryptographically bound to network keys
- ✅ **Recoverable:** Deterministic generation from mnemonic
- ✅ **MetaMask compatible:** Standard Ethereum private keys

#### Path Structure

- Quantum path: `m/44'/1955'/0'/0'/0'` (unchanged)
- Network paths: `m/44'/{coin}'/0'/0/{wots_index}`
  - Where `{wots_index} = first_4_bytes(SHA3_256(WOTS_PK)) & 0x7FFFFFFF`

#### Other Commands

```bash
# Recover from existing mnemonic
./sleevage --single-seed --quantum "your 24 word mnemonic phrase"

# Generate with passphrase (25th word)
./sleevage --single-seed --pass "your passphrase"

# Derive Bitcoin key
go run tools/derive-network.go -mnemonic "..." -network "Bitcoin" -cointype 0

# Derive Polkadot key
go run tools/derive-network.go -mnemonic "..." -network "Polkadot" -cointype 354

# See all supported networks
go run tools/derive-network.go -list
```

#### Documentation

- **[SINGLE_SEED.md](SINGLE_SEED.md)** - Complete user guide with step-by-step instructions
- **[tools/README.md](tools/README.md)** - Network derivation tool documentation

## References

Academic papers for Sleeve can be found [here](https://eprint.iacr.org/2021/872.pdf) and [here](https://eprint.iacr.org/2022/888.pdf).

This implementation of Sleeve has been audited, and the report can be found in [audit](audit).
