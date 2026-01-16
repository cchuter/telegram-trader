package ton

// Wallet represents a TON wallet
// POC: Basic structure for future wallet operations
type Wallet struct {
	Address string
}

// NewWallet creates a new wallet instance
func NewWallet(address string) *Wallet {
	return &Wallet{
		Address: address,
	}
}
