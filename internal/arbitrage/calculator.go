package arbitrage

import (
	"fmt"
)

const (
	// MinTONBalance is the minimum TON balance to keep (1 TON)
	MinTONBalance = 1.0
	// MinGALABalance is the minimum GALA balance to keep (10 GALA)
	MinGALABalance = 10.0
	// SafetyMargin is the safety multiplier above minimums (1.5x)
	SafetyMargin = 1.5
	// PositionSizePercent is the percentage of balance to use (50%)
	PositionSizePercent = 0.5
)

// PositionSize represents the calculated amounts for both legs of an arbitrage trade
type PositionSize struct {
	// TONAmount is the amount of TON to trade
	TONAmount float64
	// GALAAmount is the amount of GALA to trade
	GALAAmount float64
	// Valid indicates if the position size is valid (sufficient balance)
	Valid bool
	// Reason explains why position is invalid (if Valid is false)
	Reason string
}

// Calculator handles position sizing for arbitrage trades
type Calculator struct{}

// NewCalculator creates a new position sizing calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// CalculatePositionSize calculates the position size for an arbitrage trade
// It uses 50% of the available balance while ensuring minimum balances remain
// with a 1.5x safety margin.
//
// Parameters:
//   - tonBalance: current TON balance
//   - galaBalance: current GALA balance
//   - direction: arbitrage direction (BuyTonSellGala or BuyGalaSellTon)
//
// Returns:
//   - PositionSize with calculated amounts and validity status
func (c *Calculator) CalculatePositionSize(tonBalance, galaBalance float64, direction Direction) *PositionSize {
	// Apply safety margin to minimum balances
	safeTONMin := MinTONBalance * SafetyMargin   // 1.5 TON
	safeGALAMin := MinGALABalance * SafetyMargin // 15 GALA

	switch direction {
	case BuyTonSellGala:
		// Buy TON on ston.fi (spend GALA), sell on gswap (get GTON)
		// We need to ensure we have enough GALA to buy TON
		// and keep minimum balances after trade

		// Calculate maximum GALA we can spend (50% of balance)
		maxGALAToSpend := galaBalance * PositionSizePercent

		// Check if we have enough GALA after reserving minimums
		availableGALA := galaBalance - safeGALAMin
		if availableGALA <= 0 {
			return &PositionSize{
				Valid:  false,
				Reason: fmt.Sprintf("Insufficient GALA balance. Need at least %.1f GALA (1.5x safety margin), have %.2f", safeGALAMin, galaBalance),
			}
		}

		// Use the smaller of 50% balance or available after minimums
		galaAmount := maxGALAToSpend
		if galaAmount > availableGALA {
			galaAmount = availableGALA
		}

		// Check if TON balance meets minimum requirements
		if tonBalance < safeTONMin {
			return &PositionSize{
				Valid:  false,
				Reason: fmt.Sprintf("Insufficient TON balance. Need at least %.1f TON (1.5x safety margin), have %.2f", safeTONMin, tonBalance),
			}
		}

		return &PositionSize{
			TONAmount:  0, // Will receive TON from ston.fi swap
			GALAAmount: galaAmount,
			Valid:      true,
		}

	case BuyGalaSellTon:
		// Buy GALA on gswap (spend GTON), sell on ston.fi (get TON)
		// We need to ensure we have enough TON to buy GALA
		// and keep minimum balances after trade

		// Calculate maximum TON we can spend (50% of balance)
		maxTONToSpend := tonBalance * PositionSizePercent

		// Check if we have enough TON after reserving minimums
		availableTON := tonBalance - safeTONMin
		if availableTON <= 0 {
			return &PositionSize{
				Valid:  false,
				Reason: fmt.Sprintf("Insufficient TON balance. Need at least %.1f TON (1.5x safety margin), have %.2f", safeTONMin, tonBalance),
			}
		}

		// Use the smaller of 50% balance or available after minimums
		tonAmount := maxTONToSpend
		if tonAmount > availableTON {
			tonAmount = availableTON
		}

		// Check if GALA balance meets minimum requirements
		if galaBalance < safeGALAMin {
			return &PositionSize{
				Valid:  false,
				Reason: fmt.Sprintf("Insufficient GALA balance. Need at least %.1f GALA (1.5x safety margin), have %.2f", safeGALAMin, galaBalance),
			}
		}

		return &PositionSize{
			TONAmount:  tonAmount,
			GALAAmount: 0, // Will receive GALA from gswap swap
			Valid:      true,
		}

	default:
		return &PositionSize{
			Valid:  false,
			Reason: fmt.Sprintf("Unknown arbitrage direction: %s", direction),
		}
	}
}
