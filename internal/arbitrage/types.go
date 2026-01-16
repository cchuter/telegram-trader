package arbitrage

// Direction represents the arbitrage trading direction
type Direction string

const (
	// BuyTonSellGala means buy on ston.fi (TON->GALA) and sell on gswap (GALA->GTON)
	BuyTonSellGala Direction = "BuyTonSellGala"
	// BuyGalaSellTon means buy on gswap (GTON->GALA) and sell on ston.fi (GALA->TON)
	BuyGalaSellTon Direction = "BuyGalaSellTon"
)

// Opportunity represents an arbitrage opportunity
type Opportunity struct {
	// Direction indicates which way to execute the arbitrage
	Direction Direction
	// StonfiPrice is the TON/GALA price on ston.fi (in GALA per TON)
	StonfiPrice float64
	// GswapPrice is the GTON/GALA price on gswap (in GALA per GTON)
	GswapPrice float64
	// Spread is the percentage difference between prices
	Spread float64
}
