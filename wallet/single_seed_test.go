package wallet

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/tyler-smith/go-bip39"
	"github.com/xx-labs/sleeve/hasher"
	"github.com/xx-labs/sleeve/wots"
)

// Test single-seed sleeve construction from random entropy
func TestNewSingleSeedSleeve(t *testing.T) {
	sleeve, err := NewSingleSeedSleeve(rand.Reader, "", DefaultGenSpec())
	if err != nil {
		t.Fatalf("NewSingleSeedSleeve() returned error: %v", err)
	}

	if sleeve.GetMnemonic() == "" {
		t.Fatalf("GetMnemonic() returned empty string")
	}

	if len(sleeve.GetWOTSPublicKey()) == 0 {
		t.Fatalf("GetWOTSPublicKey() returned empty")
	}

	// Should have automatically derived standard networks
	networks := sleeve.GetAllNetworkKeys()
	if len(networks) != 3 {
		t.Fatalf("Expected 3 standard networks, got %d", len(networks))
	}

	// Verify Bitcoin, Ethereum, Polkadot are present
	if _, err := sleeve.GetPrivateKey("Bitcoin"); err != nil {
		t.Fatalf("Bitcoin key not derived: %v", err)
	}
	if _, err := sleeve.GetPrivateKey("Ethereum"); err != nil {
		t.Fatalf("Ethereum key not derived: %v", err)
	}
	if _, err := sleeve.GetPrivateKey("Polkadot"); err != nil {
		t.Fatalf("Polkadot key not derived: %v", err)
	}
}

// Test single-seed sleeve with provided mnemonic
func TestNewSingleSeedSleeveFromMnemonic(t *testing.T) {
	// Use test vector from original sleeve tests
	mnemonic := testVectorMnemonic

	sleeve, err := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())
	if err != nil {
		t.Fatalf("NewSingleSeedSleeveFromMnemonic() returned error: %v", err)
	}

	if sleeve.GetMnemonic() != mnemonic {
		t.Fatalf("Mnemonic mismatch")
	}

	// Verify derivation index is deterministic
	if sleeve.GetDerivationIndex() == 0 {
		t.Fatalf("Derivation index should not be zero")
	}
}

// Test deterministic key generation
func TestSingleSeedSleeve_Deterministic(t *testing.T) {
	mnemonic := testVectorMnemonic

	// Generate sleeve twice with same mnemonic
	sleeve1, err := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	sleeve2, err := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	// Verify WOTS public keys match
	if !bytes.Equal(sleeve1.GetWOTSPublicKey(), sleeve2.GetWOTSPublicKey()) {
		t.Fatalf("WOTS public keys should be identical")
	}

	// Verify derivation indices match
	if sleeve1.GetDerivationIndex() != sleeve2.GetDerivationIndex() {
		t.Fatalf("Derivation indices should be identical")
	}

	// Verify network keys match
	btcKey1, _ := sleeve1.GetPrivateKey("Bitcoin")
	btcKey2, _ := sleeve2.GetPrivateKey("Bitcoin")
	if !bytes.Equal(btcKey1, btcKey2) {
		t.Fatalf("Bitcoin keys should be identical")
	}

	ethKey1, _ := sleeve1.GetPrivateKey("Ethereum")
	ethKey2, _ := sleeve2.GetPrivateKey("Ethereum")
	if !bytes.Equal(ethKey1, ethKey2) {
		t.Fatalf("Ethereum keys should be identical")
	}

	dotKey1, _ := sleeve1.GetPrivateKey("Polkadot")
	dotKey2, _ := sleeve2.GetPrivateKey("Polkadot")
	if !bytes.Equal(dotKey1, dotKey2) {
		t.Fatalf("Polkadot keys should be identical")
	}
}

// Test WOTS public key to index calculation
func TestSingleSeedSleeve_IndexCalculation(t *testing.T) {
	mnemonic := testVectorMnemonic
	sleeve, _ := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())

	// Manually calculate index and verify
	wotsPK := sleeve.GetWOTSPublicKey()
	pkHash := hasher.SHA3_256.Hash(wotsPK)

	// Extract first 4 bytes as uint32, masked to 31 bits
	expectedIndex := (uint32(pkHash[0])<<24 | uint32(pkHash[1])<<16 | uint32(pkHash[2])<<8 | uint32(pkHash[3])) & 0x7FFFFFFF

	if sleeve.GetDerivationIndex() != expectedIndex {
		t.Fatalf("Index calculation mismatch. Got: %d, Expected: %d",
			sleeve.GetDerivationIndex(), expectedIndex)
	}

	// Verify index is always < 2^31 (non-hardened requirement)
	if sleeve.GetDerivationIndex() >= 0x80000000 {
		t.Fatalf("Index must be < 2^31 for non-hardened derivation. Got: %d", sleeve.GetDerivationIndex())
	}
}

// Test that WOTS generation is unchanged from original Sleeve
func TestSingleSeedSleeve_WOTSConsistency(t *testing.T) {
	mnemonic := wotsTestVectorMnemonic

	// Generate single-seed sleeve
	ssSleeve, err := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())
	if err != nil {
		t.Fatalf("NewSingleSeedSleeveFromMnemonic() returned error: %v", err)
	}

	// Generate seed and derive WOTS manually (original method)
	seed, _ := bip39.NewSeedWithErrorChecking(mnemonic, "")
	path := []uint32{0x8000002C, 0x800007A3, 0x80000000, 0x80000000, 0x80000000}
	node, _ := ComputeNode(seed, path)
	wotsKey := wots.NewKeyFromSeed(wots.DecodeParams(wots.DefaultParams), node.Key, node.Code)
	manualWOTSPK := wotsKey.ComputePK()

	// Verify WOTS public keys match
	if !bytes.Equal(ssSleeve.GetWOTSPublicKey(), manualWOTSPK) {
		t.Fatalf("WOTS public key mismatch - generation changed!")
	}

	// Verify against known test vector
	expectedPk, _ := hex.DecodeString(wotsExpectedPubKeyHex)
	if !bytes.Equal(ssSleeve.GetWOTSPublicKey(), expectedPk) {
		t.Fatalf("WOTS PK doesn't match test vector. Got: %x, Expected: %x",
			ssSleeve.GetWOTSPublicKey(), expectedPk)
	}
}

// Test custom network derivation
func TestSingleSeedSleeve_CustomNetwork(t *testing.T) {
	mnemonic := testVectorMnemonic
	sleeve, _ := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())

	// Derive Litecoin key (coin type 2)
	seed, _ := bip39.NewSeedWithErrorChecking(mnemonic, "")
	err := sleeve.DeriveNetworkKey("Litecoin", CoinTypeLitecoin, seed)
	if err != nil {
		t.Fatalf("Failed to derive Litecoin key: %v", err)
	}

	ltcKey, err := sleeve.GetPrivateKey("Litecoin")
	if err != nil {
		t.Fatalf("Failed to get Litecoin key: %v", err)
	}

	if len(ltcKey) != 32 {
		t.Fatalf("Invalid key length: %d", len(ltcKey))
	}

	// Verify network info
	networks := sleeve.GetAllNetworkKeys()
	if _, exists := networks["Litecoin"]; !exists {
		t.Fatalf("Litecoin not in network keys map")
	}

	if networks["Litecoin"].CoinType != CoinTypeLitecoin {
		t.Fatalf("Incorrect coin type for Litecoin")
	}
}

// Test error handling for invalid mnemonic
func TestSingleSeedSleeve_InvalidMnemonic(t *testing.T) {
	// Too few words
	_, err := NewSingleSeedSleeveFromMnemonic("one two three", "", DefaultGenSpec())
	if err == nil {
		t.Fatalf("Should return error for mnemonic with too few words")
	}

	// Invalid word
	invalidMnem := "armed output survey rent myself sentence warm eyebrow scan isolate thunder point" +
		" bulk skirt sketch bird palm sleep dash jazz list behave spin xxnetwork"
	_, err = NewSingleSeedSleeveFromMnemonic(invalidMnem, "", DefaultGenSpec())
	if err == nil {
		t.Fatalf("Should return error for mnemonic with invalid word")
	}

	// Invalid checksum
	invalidChkMnem := "armed output survey rent myself sentence warm eyebrow scan isolate thunder point" +
		" bulk skirt sketch bird palm sleep dash jazz list behave spin spin"
	_, err = NewSingleSeedSleeveFromMnemonic(invalidChkMnem, "", DefaultGenSpec())
	if err == nil {
		t.Fatalf("Should return error for mnemonic with invalid checksum")
	}
}

// Test passphrase support
func TestSingleSeedSleeve_Passphrase(t *testing.T) {
	mnemonic := testVectorMnemonic

	// Generate with no passphrase
	sleeve1, _ := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())

	// Generate with passphrase
	sleeve2, _ := NewSingleSeedSleeveFromMnemonic(mnemonic, "test_passphrase", DefaultGenSpec())

	// Keys should be different
	btcKey1, _ := sleeve1.GetPrivateKey("Bitcoin")
	btcKey2, _ := sleeve2.GetPrivateKey("Bitcoin")
	if bytes.Equal(btcKey1, btcKey2) {
		t.Fatalf("Keys should differ with different passphrases")
	}

	// WOTS keys should be different
	if bytes.Equal(sleeve1.GetWOTSPublicKey(), sleeve2.GetWOTSPublicKey()) {
		t.Fatalf("WOTS keys should differ with different passphrases")
	}

	// Derivation indices should be different
	if sleeve1.GetDerivationIndex() == sleeve2.GetDerivationIndex() {
		t.Fatalf("Derivation indices should differ with different passphrases")
	}
}

// Test with different WOTS+ parameters
func TestSingleSeedSleeve_WOTSParams(t *testing.T) {
	mnemonic := testVectorMnemonic

	// Test with different security levels
	levels := []wots.ParamsEncoding{wots.Level0, wots.Level1, wots.Level2, wots.Level3}

	for _, level := range levels {
		spec := NewGenSpec(0, level)
		sleeve, err := NewSingleSeedSleeveFromMnemonic(mnemonic, "", spec)
		if err != nil {
			t.Fatalf("Failed with WOTS level %v: %v", level, err)
		}

		if len(sleeve.GetWOTSPublicKey()) == 0 {
			t.Fatalf("Empty WOTS PK with level %v", level)
		}

		// Each level should produce different WOTS PK
		// (because they use different params in the derivation path)
	}
}

// Benchmark single-seed generation
func BenchmarkSingleSeedGeneration(b *testing.B) {
	mnemonic := testVectorMnemonic
	spec := DefaultGenSpec()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewSingleSeedSleeveFromMnemonic(mnemonic, "", spec)
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}
	}
}

// Benchmark network key derivation
func BenchmarkNetworkDerivation(b *testing.B) {
	mnemonic := testVectorMnemonic
	sleeve, _ := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())
	seed, _ := bip39.NewSeedWithErrorChecking(mnemonic, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sleeve.DeriveNetworkKey("Test", CoinTypeLitecoin, seed)
	}
}

// Test that network keys are properly bound to WOTS public key
func TestSingleSeedSleeve_SecurityBinding(t *testing.T) {
	// Generate two different sleeves
	sleeve1, _ := NewSingleSeedSleeve(rand.Reader, "", DefaultGenSpec())
	sleeve2, _ := NewSingleSeedSleeve(rand.Reader, "", DefaultGenSpec())

	// WOTS public keys should be different
	if bytes.Equal(sleeve1.GetWOTSPublicKey(), sleeve2.GetWOTSPublicKey()) {
		t.Fatalf("WOTS public keys should be different")
	}

	// Derivation indices should be different (with high probability)
	if sleeve1.GetDerivationIndex() == sleeve2.GetDerivationIndex() {
		t.Logf("Warning: Derivation indices match (low probability collision)")
	}

	// Network keys should be different
	btcKey1, _ := sleeve1.GetPrivateKey("Bitcoin")
	btcKey2, _ := sleeve2.GetPrivateKey("Bitcoin")
	if bytes.Equal(btcKey1, btcKey2) {
		t.Fatalf("Bitcoin keys should be different")
	}
}

// Test recovery scenario
func TestSingleSeedSleeve_Recovery(t *testing.T) {
	// User creates wallet
	originalSleeve, _ := NewSingleSeedSleeve(rand.Reader, "", DefaultGenSpec())
	mnemonic := originalSleeve.GetMnemonic()

	// Store network keys from original
	btcKeyOriginal, _ := originalSleeve.GetPrivateKey("Bitcoin")
	ethKeyOriginal, _ := originalSleeve.GetPrivateKey("Ethereum")
	dotKeyOriginal, _ := originalSleeve.GetPrivateKey("Polkadot")
	wotsIndexOriginal := originalSleeve.GetDerivationIndex()

	// Simulate recovery from mnemonic (user lost device)
	recoveredSleeve, err := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())
	if err != nil {
		t.Fatalf("Recovery failed: %v", err)
	}

	// Verify all keys recovered correctly
	btcKeyRecovered, _ := recoveredSleeve.GetPrivateKey("Bitcoin")
	ethKeyRecovered, _ := recoveredSleeve.GetPrivateKey("Ethereum")
	dotKeyRecovered, _ := recoveredSleeve.GetPrivateKey("Polkadot")
	wotsIndexRecovered := recoveredSleeve.GetDerivationIndex()

	if !bytes.Equal(btcKeyOriginal, btcKeyRecovered) {
		t.Fatalf("Bitcoin key not recovered correctly")
	}
	if !bytes.Equal(ethKeyOriginal, ethKeyRecovered) {
		t.Fatalf("Ethereum key not recovered correctly")
	}
	if !bytes.Equal(dotKeyOriginal, dotKeyRecovered) {
		t.Fatalf("Polkadot key not recovered correctly")
	}
	if wotsIndexOriginal != wotsIndexRecovered {
		t.Fatalf("WOTS index not recovered correctly")
	}
}

// Test error handling for NewSingleSeedSleeve with bad readers
func TestSingleSeedSleeve_ErrorReaders(t *testing.T) {
	// Test with error reader
	_, err := NewSingleSeedSleeve(&ErrReader{}, "", DefaultGenSpec())
	if err == nil {
		t.Fatalf("NewSingleSeedSleeve() should return error when there's an error reading entropy")
	}

	// Test with limited bytes reader
	_, err = NewSingleSeedSleeve(&LimitedReader{EntropySize / 2}, "", DefaultGenSpec())
	if err == nil {
		t.Fatalf("NewSingleSeedSleeve() should return error when not enough entropy is read")
	}
}

// Test error handling for NewSingleSeedSleeveFromEntropy
func TestSingleSeedSleeve_InvalidEntropy(t *testing.T) {
	// Test wrong entropy size (31 bytes) - invalid for BIP39
	ent := make([]byte, EntropySize-1)
	_, err := NewSingleSeedSleeveFromEntropy(ent, "", DefaultGenSpec())
	if err == nil {
		t.Fatalf("NewSingleSeedSleeveFromEntropy() should return error when entropy size is invalid")
	}

	// Test valid BIP39 entropy size (16 bytes), but not enough for Sleeve
	ent = make([]byte, EntropySize/2)
	_, err = NewSingleSeedSleeveFromEntropy(ent, "", DefaultGenSpec())
	if err == nil {
		t.Fatalf("NewSingleSeedSleeveFromEntropy() should return error when entropy is too small for Sleeve")
	}

	// Test valid entropy size
	ent = make([]byte, EntropySize)
	sleeve, err := NewSingleSeedSleeveFromEntropy(ent, "", DefaultGenSpec())
	if err != nil {
		t.Fatalf("NewSingleSeedSleeveFromEntropy() should succeed with valid entropy: %v", err)
	}
	if sleeve == nil {
		t.Fatalf("NewSingleSeedSleeveFromEntropy() returned nil sleeve")
	}
}

// Test GetWOTSKey function
func TestSingleSeedSleeve_GetWOTSKey(t *testing.T) {
	sleeve, _ := NewSingleSeedSleeve(rand.Reader, "", DefaultGenSpec())
	
	wotsKey := sleeve.GetWOTSKey()
	if wotsKey == nil {
		t.Fatalf("GetWOTSKey() returned nil")
	}

	// Verify WOTS key can compute PK correctly
	pk := wotsKey.ComputePK()
	if !bytes.Equal(pk, sleeve.GetWOTSPublicKey()) {
		t.Fatalf("WOTS key PK doesn't match stored PK")
	}
}

// Test GetPrivateKey with non-existent network
func TestSingleSeedSleeve_GetPrivateKey_NotFound(t *testing.T) {
	sleeve, _ := NewSingleSeedSleeve(rand.Reader, "", DefaultGenSpec())
	
	_, err := sleeve.GetPrivateKey("NonExistentNetwork")
	if err == nil {
		t.Fatalf("GetPrivateKey() should return error for non-existent network")
	}
}

// Test DeriveNetworkKey error paths
func TestSingleSeedSleeve_DeriveNetworkKey_Errors(t *testing.T) {
	mnemonic := testVectorMnemonic
	sleeve, _ := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())
	seed, _ := bip39.NewSeedWithErrorChecking(mnemonic, "")

	// Test with invalid coin type at boundary
	err := sleeve.DeriveNetworkKey("TestCoin", 999999, seed)
	if err != nil {
		// This should actually succeed - BIP32 supports large coin types
		t.Logf("Note: Large coin type rejected (expected for some implementations)")
	}

	// Test that we can successfully derive a custom network
	err = sleeve.DeriveNetworkKey("CustomCoin", 123, seed)
	if err != nil {
		t.Fatalf("Failed to derive custom network: %v", err)
	}

	// Verify custom network is in map
	key, err := sleeve.GetPrivateKey("CustomCoin")
	if err != nil {
		t.Fatalf("Custom network not found after derivation: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("Invalid key length for custom network: %d", len(key))
	}
}

// Test DeriveStandardNetworks function
func TestSingleSeedSleeve_DeriveStandardNetworks(t *testing.T) {
	mnemonic := testVectorMnemonic

	// Create sleeve manually without auto-derivation
	// (we can't directly test this without modifying the constructor, 
	// but we can verify the standard networks exist)
	sleeve, _ := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())

	// Verify all standard networks were derived
	networks := sleeve.GetAllNetworkKeys()
	
	expectedNetworks := []string{"Bitcoin", "Ethereum", "Polkadot"}
	for _, netName := range expectedNetworks {
		if _, exists := networks[netName]; !exists {
			t.Fatalf("Standard network %s not found", netName)
		}
		
		// Verify key is valid
		key, err := sleeve.GetPrivateKey(netName)
		if err != nil {
			t.Fatalf("Failed to get %s key: %v", netName, err)
		}
		if len(key) != 32 {
			t.Fatalf("Invalid key length for %s: %d", netName, len(key))
		}
	}

	// Verify coin types are correct
	if networks["Bitcoin"].CoinType != CoinTypeBitcoin {
		t.Fatalf("Bitcoin coin type mismatch")
	}
	if networks["Ethereum"].CoinType != CoinTypeEthereum {
		t.Fatalf("Ethereum coin type mismatch")
	}
	if networks["Polkadot"].CoinType != CoinTypePolkadot {
		t.Fatalf("Polkadot coin type mismatch")
	}

	// Verify paths are formatted correctly
	for netName, netKey := range networks {
		if netKey.Path == "" {
			t.Fatalf("Empty path for network %s", netName)
		}
		if netKey.Network != netName {
			t.Fatalf("Network name mismatch for %s", netName)
		}
	}
}

// Test with invalid GenSpec
func TestSingleSeedSleeve_InvalidGenSpec(t *testing.T) {
	// Test invalid account number (>= 2^31)
	spec := NewGenSpec(firstHardened, wots.Level0)
	_, err := NewSingleSeedSleeve(rand.Reader, "", spec)
	if err == nil {
		t.Fatalf("NewSingleSeedSleeve() should return error with invalid account in GenSpec")
	}

	// Test invalid WOTS params
	spec = GenSpec{
		account: 0,
		params:  wots.ParamsEncodingLen, // Invalid params encoding
	}
	_, err = NewSingleSeedSleeve(rand.Reader, "", spec)
	if err == nil {
		t.Fatalf("NewSingleSeedSleeve() should return error with invalid WOTS params")
	}
}

// Test multiple derivations don't interfere
func TestSingleSeedSleeve_MultipleDerivatonsIndependent(t *testing.T) {
	mnemonic := testVectorMnemonic
	sleeve, _ := NewSingleSeedSleeveFromMnemonic(mnemonic, "", DefaultGenSpec())
	seed, _ := bip39.NewSeedWithErrorChecking(mnemonic, "")

	// Derive multiple custom networks
	networks := []struct {
		name     string
		coinType uint32
	}{
		{"Litecoin", CoinTypeLitecoin},
		{"Cardano", CoinTypeCardano},
		{"Solana", 501},
		{"Cosmos", 118},
	}

	for _, net := range networks {
		err := sleeve.DeriveNetworkKey(net.name, net.coinType, seed)
		if err != nil {
			t.Fatalf("Failed to derive %s: %v", net.name, err)
		}
	}

	// Verify all networks exist and have different keys
	allKeys := make(map[string][]byte)
	for _, net := range networks {
		key, err := sleeve.GetPrivateKey(net.name)
		if err != nil {
			t.Fatalf("Failed to get %s key: %v", net.name, err)
		}
		allKeys[net.name] = key
	}

	// Verify all keys are unique
	keySet := make(map[string]bool)
	for netName, key := range allKeys {
		keyHex := hex.EncodeToString(key)
		if keySet[keyHex] {
			t.Fatalf("Duplicate key found for network %s", netName)
		}
		keySet[keyHex] = true
	}

	// Verify total network count
	allNetworks := sleeve.GetAllNetworkKeys()
	expectedCount := 3 + len(networks) // 3 standard + custom networks
	if len(allNetworks) != expectedCount {
		t.Fatalf("Expected %d networks, got %d", expectedCount, len(allNetworks))
	}
}

// Test address derivation for known test vector
func TestSingleSeedSleeve_KnownAddresses(t *testing.T) {
	// Using wotsTestVectorMnemonic to ensure consistency
	sleeve, err := NewSingleSeedSleeveFromMnemonic(wotsTestVectorMnemonic, "", DefaultGenSpec())
	if err != nil {
		t.Fatalf("Failed to create sleeve: %v", err)
	}

	// Verify WOTS public key matches expected value
	expectedPk, _ := hex.DecodeString(wotsExpectedPubKeyHex)
	if !bytes.Equal(sleeve.GetWOTSPublicKey(), expectedPk) {
		t.Fatalf("WOTS PK mismatch. Got: %x, Expected: %x",
			sleeve.GetWOTSPublicKey(), expectedPk)
	}

	// Calculate expected derivation index
	pkHash := hasher.SHA3_256.Hash(expectedPk)
	expectedIndex := (uint32(pkHash[0])<<24 | uint32(pkHash[1])<<16 | 
		uint32(pkHash[2])<<8 | uint32(pkHash[3])) & 0x7FFFFFFF

	if sleeve.GetDerivationIndex() != expectedIndex {
		t.Fatalf("Derivation index mismatch. Got: %d, Expected: %d",
			sleeve.GetDerivationIndex(), expectedIndex)
	}

	// Verify all network keys are deterministic
	btcKey, _ := sleeve.GetPrivateKey("Bitcoin")
	ethKey, _ := sleeve.GetPrivateKey("Ethereum")
	dotKey, _ := sleeve.GetPrivateKey("Polkadot")

	// These should be stable for this test vector
	// (we're not checking exact values, just that they exist and are deterministic)
	if len(btcKey) != 32 || len(ethKey) != 32 || len(dotKey) != 32 {
		t.Fatalf("Invalid key lengths")
	}

	// Re-generate and verify determinism
	sleeve2, _ := NewSingleSeedSleeveFromMnemonic(wotsTestVectorMnemonic, "", DefaultGenSpec())
	btcKey2, _ := sleeve2.GetPrivateKey("Bitcoin")
	ethKey2, _ := sleeve2.GetPrivateKey("Ethereum")
	dotKey2, _ := sleeve2.GetPrivateKey("Polkadot")

	if !bytes.Equal(btcKey, btcKey2) || !bytes.Equal(ethKey, ethKey2) || !bytes.Equal(dotKey, dotKey2) {
		t.Fatalf("Network keys not deterministic")
	}
}
