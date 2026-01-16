package validation

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/xssnick/tonutils-go/address"
)

// ValidTokenSymbols is the whitelist of supported token symbols
var ValidTokenSymbols = map[string]bool{
	"TON":  true,
	"GALA": true,
	"GTON": true,
}

// ValidateTonAddress validates a TON address with checksum verification
// Uses tonutils-go library for proper validation
func ValidateTonAddress(addr string) error {
	// Use tonutils-go to parse and validate the address
	_, err := address.ParseAddr(addr)
	if err != nil {
		return fmt.Errorf("invalid TON address: %v", err)
	}
	return nil
}

// ValidateGalaAddress validates a GalaChain address (hex format, 64 chars)
func ValidateGalaAddress(address string) error {
	// Check length (64 hex characters = 32 bytes)
	if len(address) != 64 {
		return fmt.Errorf("invalid GalaChain address length (expected 64 characters, got %d)", len(address))
	}

	// Verify hex format
	if !isHex(address) {
		return fmt.Errorf("address must be hexadecimal")
	}

	return nil
}

// isHex checks if a string contains only hexadecimal characters
func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// ValidateAmount validates an amount string (positive, max 18 decimals)
func ValidateAmount(amount string) (*big.Float, error) {
	// Parse as big.Float for precision
	value, ok := new(big.Float).SetString(amount)
	if !ok {
		return nil, fmt.Errorf("invalid number format")
	}

	// Check positive
	if value.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Check decimal places by counting from the original string
	parts := strings.Split(amount, ".")
	if len(parts) == 2 {
		if len(parts[1]) > 18 {
			return nil, fmt.Errorf("amount has too many decimal places (max 18)")
		}
	}

	return value, nil
}

// ValidateTokenSymbol validates a token symbol against the whitelist
func ValidateTokenSymbol(symbol string) error {
	normalized := strings.ToUpper(symbol)
	if !ValidTokenSymbols[normalized] {
		return fmt.Errorf("unsupported token symbol (supported: TON, GALA, GTON)")
	}
	return nil
}
