package stonfi

import (
	"strconv"
)

// SwapSimulateRequest represents the request body for POST /v1/swap/simulate
type SwapSimulateRequest struct {
	OfferAddress      string `json:"offer_address"`      // Source token address or "TON"
	AskAddress        string `json:"ask_address"`        // Destination token address
	Units             string `json:"units"`              // Input amount in smallest units
	SlippageTolerance string `json:"slippage_tolerance"` // Slippage tolerance (e.g., "0.01" for 1%)
}

// SwapSimulateResponse represents the response from POST /v1/swap/simulate
type SwapSimulateResponse struct {
	OfferAddress  string `json:"offer_address"`            // Source token address
	AskAddress    string `json:"ask_address"`              // Destination token address
	OfferUnits    string `json:"offer_units"`              // Input amount
	AskUnits      string `json:"ask_units"`                // Expected output amount
	FeeAddress    string `json:"fee_address"`              // Fee token address
	FeeUnits      string `json:"fee_units"`                // Fee amount
	FeePercent    string `json:"fee_percent"`              // Fee percentage (e.g., "0.3" for 0.3%)
	MinAskUnits   string `json:"min_ask_units"`            // Minimum output after slippage
	PriceImpact   string `json:"price_impact"`             // Price impact (e.g., "0.01" for 1%)
	SwapRate      string `json:"swap_rate"`                // Exchange rate
	RouterAddress string `json:"router_address,omitempty"` // DEX router address
}

// SlippagePercentage returns the slippage as a percentage
// Calculated as: (AskUnits - MinAskUnits) / AskUnits * 100
func (r *SwapSimulateResponse) SlippagePercentage() float64 {
	askUnits, err := strconv.ParseFloat(r.AskUnits, 64)
	if err != nil || askUnits == 0 {
		return 0
	}

	minAskUnits, err := strconv.ParseFloat(r.MinAskUnits, 64)
	if err != nil {
		return 0
	}

	slippage := (askUnits - minAskUnits) / askUnits * 100
	return slippage
}

// PriceImpactPercentage returns the price impact as a percentage
func (r *SwapSimulateResponse) PriceImpactPercentage() float64 {
	impact, err := strconv.ParseFloat(r.PriceImpact, 64)
	if err != nil {
		return 0
	}
	return impact * 100 // Convert to percentage
}

// FeePercentageFloat returns the fee percentage as a float
func (r *SwapSimulateResponse) FeePercentageFloat() float64 {
	feePercent, err := strconv.ParseFloat(r.FeePercent, 64)
	if err != nil {
		return 0
	}
	return feePercent
}
