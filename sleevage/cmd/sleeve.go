////////////////////////////////////////////////////////////////////////////////////////////
// Copyright © 2021 xx network SEZC                                                       //
//                                                                                        //
// Use of this source code is governed by a license that can be found in the LICENSE file //
////////////////////////////////////////////////////////////////////////////////////////////

package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/xx-labs/sleeve/wallet"
	"github.com/xx-labs/sleeve/wots"
)

type StandardDerivation struct {
	Path    string `json:"Path"`
	Address string `json:"Address"`
}

func (s StandardDerivation) String() string {
	return fmt.Sprintf("%s:    %s\n", s.Path, s.Address)
}

type SleeveJson struct {
	Quantum       string               `json:"QuantumPhrase"`
	Pass          string               `json:"Passphrase"`
	Path          string               `json:"DerivationPath"`
	Standard      string               `json:"StandardPhrase,omitempty"` // Empty for single-seed mode
	Address       string               `json:"Address"`
	StandardDeriv []StandardDerivation `json:"StandardDerivations,omitempty"`
	// Single-seed specific fields
	SingleSeed    bool                 `json:"SingleSeed,omitempty"`
	WOTSIndex     uint32               `json:"WOTSIndex,omitempty"`
	WOTSPublicKey string               `json:"WOTSPublicKey,omitempty"`
	NetworkKeys   []NetworkKeyInfo     `json:"NetworkKeys,omitempty"`
}

type NetworkKeyInfo struct {
	Network  string `json:"Network"`
	CoinType uint32 `json:"CoinType"`
	Path     string `json:"Path"`
	Address  string `json:"Address,omitempty"` // For display purposes
}

func (s SleeveJson) String() string {
	str := fmt.Sprintf("quantum recovery phrase: %s\n", s.Quantum)
	str += fmt.Sprintf("passphrase: %s\n", s.Pass)
	str += fmt.Sprintf("path: %s\n", s.Path)
	
	if s.SingleSeed {
		// Single-seed mode output
		str += fmt.Sprintf("generation mode: SINGLE-SEED\n")
		str += fmt.Sprintf("WOTS+ public key: %s\n", s.WOTSPublicKey)
		str += fmt.Sprintf("WOTS-derived index: %d\n", s.WOTSIndex)
		str += fmt.Sprintf("address (xx network): %s\n", s.Address)
		if len(s.NetworkKeys) > 0 {
			str += fmt.Sprintf("\nderived network keys:\n")
			for _, nk := range s.NetworkKeys {
				str += fmt.Sprintf("  %s (coin %d): %s\n", nk.Network, nk.CoinType, nk.Path)
				if nk.Address != "" {
					str += fmt.Sprintf("    address: %s\n", nk.Address)
				}
			}
		}
	} else {
		// Legacy dual-mnemonic mode output
		str += fmt.Sprintf("generation mode: DUAL-MNEMONIC (legacy)\n")
		str += fmt.Sprintf("standard recovery phrase: %s\n", s.Standard)
		str += fmt.Sprintf("address: %s", s.Address)
		if s.StandardDeriv != nil {
			str += fmt.Sprintf("\nstandard derived addresses:\n")
			for _, addr := range s.StandardDeriv {
				str += addr.String()
			}
		}
	}
	return str
}

type args struct {
	generate bool
	quantum  string
	pass     string
	spec     wallet.GenSpec
	path     string
}

func parseArgs() (args, error) {
	// If quantum phrase is not empty, then don't generate new wallet
	generate := true
	if quantumPhrase != "" {
		generate = false
	}

	// Select wots+ security level
	level := wots.DefaultParams
	switch wotsSecurityLevel {
	case "level0":
		level = wots.Level0
	case "level1":
		level = wots.Level1
	case "level2":
		level = wots.Level2
	case "level3":
		level = wots.Level3
	default:
		return args{}, errors.New(fmt.Sprintf("invalid WOTS+ security level specified: %s", wotsSecurityLevel))
	}

	spec := wallet.NewGenSpec(account, level)
	// Validate path from spec
	path, err := spec.PathFromSpec()
	if err != nil {
		return args{}, errors.New(fmt.Sprintf("error creating derivation path: %s. Account number is too large", err))
	}

	return args{
		generate: generate,
		quantum:  quantumPhrase,
		pass:     passphrase,
		spec:     spec,
		path:     path.String(),
	}, nil
}

func getSleeve(args args) (SleeveJson, error) {
	if singleSeed {
		// Use new single-seed generation
		return getSingleSeedSleeve(args)
	} else {
		// Use legacy dual-mnemonic generation
		return getDualMnemonicSleeve(args)
	}
}

func getDualMnemonicSleeve(args args) (SleeveJson, error) {
	var err error
	var sleeve *wallet.Sleeve
	if args.generate {
		sleeve, err = wallet.NewSleeve(rand.Reader, args.pass, args.spec)
		if err != nil {
			return SleeveJson{}, err
		}
	} else {
		sleeve, err = wallet.NewSleeveFromMnemonic(args.quantum, args.pass, args.spec)
		if err != nil {
			return SleeveJson{}, err
		}
	}
	json := getJson(args.path, sleeve)
	return json, nil
}

func getSingleSeedSleeve(args args) (SleeveJson, error) {
	var err error
	var sleeve *wallet.SingleSeedSleeve
	if args.generate {
		sleeve, err = wallet.NewSingleSeedSleeve(rand.Reader, args.pass, args.spec)
		if err != nil {
			return SleeveJson{}, err
		}
	} else {
		sleeve, err = wallet.NewSingleSeedSleeveFromMnemonic(args.quantum, args.pass, args.spec)
		if err != nil {
			return SleeveJson{}, err
		}
	}
	json := getSingleSeedJson(args.path, sleeve)
	return json, nil
}

func getAddress(sleeve *wallet.Sleeve) string {
	if testnet {
		return wallet.TestnetAddressFromMnemonic(sleeve.GetOutputMnemonic())
	}
	return wallet.XXNetworkAddressFromMnemonic(sleeve.GetOutputMnemonic())
}

func getJson(path string, sleeve *wallet.Sleeve) SleeveJson {
	var derivs []StandardDerivation = nil
	if derivations > 0 {
		derivs = make([]StandardDerivation, derivations)
		for i := uint32(0); i < derivations; i++ {
			derivPath := fmt.Sprintf("//%s//%d", prefix, i)
			if prefix == "" {
				// Fix path if no prefix
				derivPath = fmt.Sprintf("//%d", i)
			} else if derivations == 1 {
				// Fix path if only one derivation
				derivPath = fmt.Sprintf("//%s", prefix)
			}
			addr := wallet.XXNetworkAddressFromMnemonic(sleeve.GetOutputMnemonic() + derivPath)
			if testnet {
				addr = wallet.TestnetAddressFromMnemonic(sleeve.GetOutputMnemonic() + derivPath)
			}
			derivs[i] = StandardDerivation{
				Path:    derivPath,
				Address: addr,
			}
		}
	}
	return SleeveJson{
		Quantum:  sleeve.GetMnemonic(),
		Pass:     passphrase,
		Path:     path,
		Standard: sleeve.GetOutputMnemonic(),
		Address:  getAddress(sleeve),
		StandardDeriv: derivs,
	}
}

func getSingleSeedJson(path string, sleeve *wallet.SingleSeedSleeve) SleeveJson {
	// Get all network keys
	networkKeys := sleeve.GetAllNetworkKeys()
	
	// Build network key info array
	var netKeyInfos []NetworkKeyInfo
	for _, nk := range networkKeys {
		netKeyInfos = append(netKeyInfos, NetworkKeyInfo{
			Network:  nk.Network,
			CoinType: nk.CoinType,
			Path:     nk.Path,
			// Address calculation could be added here if needed
		})
	}

	// Get WOTS public key hex
	wotsPKHex := hex.EncodeToString(sleeve.GetWOTSPublicKey())

	// Create address from WOTS public key (using the same method as legacy)
	// Note: For single-seed, the "address" is based on WOTS PK, not output mnemonic
	address := fmt.Sprintf("WOTS+:%s", wotsPKHex[:16]) // Shortened for display

	return SleeveJson{
		Quantum:       sleeve.GetMnemonic(),
		Pass:          passphrase,
		Path:          path,
		Standard:      "", // No second mnemonic in single-seed mode
		Address:       address,
		StandardDeriv: nil,
		SingleSeed:    true,
		WOTSIndex:     sleeve.GetDerivationIndex(),
		WOTSPublicKey: wotsPKHex,
		NetworkKeys:   netKeyInfos,
	}
}

func sleeve() ([]SleeveJson, error) {
	// Parse args
	args, err := parseArgs()
	if err != nil {
		return nil, err
	}

	// Sleeve generation
	wallets := make([]SleeveJson, numWallets*numAccounts)
	// Keep start account
	startAccount := account
	for i := uint32(0); i < numWallets; i++ {
		for j := uint32(0); j < numAccounts; j++ {
			// Increase account number
			account = startAccount + j
			// Reparse args
			args, err = parseArgs()
			if err != nil {
				return nil, err
			}
			// Generate wallet
			wallets[i*numAccounts+j], err = getSleeve(args)
			if err != nil {
				return nil, err
			}
			// Set the quantum phrase if this is first wallet
			if j == 0 {
				quantumPhrase = wallets[i*numAccounts+j].Quantum
			}
		}
		// Reset quantum phrase to generate new wallet on next iteration
		quantumPhrase = ""
	}
	return wallets, nil
}
