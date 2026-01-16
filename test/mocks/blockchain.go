package mocks

import (
	"context"
	"sync"

	"github.com/cchuter/telegram-trader/internal/blockchain"
)

// MockBlockchainClient is a mock implementation of blockchain.Client for testing
type MockBlockchainClient struct {
	mu sync.RWMutex

	// Configurable responses
	Balances map[string]string // address -> balance

	// Error simulation
	ConnectErr    error
	GetBalanceErr error
	CloseErr      error

	// Call tracking
	ConnectCalls    int
	GetBalanceCalls map[string]int // address -> call count
	CloseCalls      int
}

// NewMockBlockchainClient creates a new mock blockchain client
func NewMockBlockchainClient() *MockBlockchainClient {
	return &MockBlockchainClient{
		Balances:        make(map[string]string),
		GetBalanceCalls: make(map[string]int),
	}
}

// Connect initializes the blockchain client connection
func (m *MockBlockchainClient) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ConnectCalls++

	if m.ConnectErr != nil {
		return m.ConnectErr
	}

	return nil
}

// GetBalance returns the wallet balance as a string
func (m *MockBlockchainClient) GetBalance(ctx context.Context, address string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetBalanceCalls[address]++

	if m.GetBalanceErr != nil {
		return "", m.GetBalanceErr
	}

	balance, exists := m.Balances[address]
	if !exists {
		return "0.0", nil
	}

	return balance, nil
}

// Close closes the blockchain client connection
func (m *MockBlockchainClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CloseCalls++

	if m.CloseErr != nil {
		return m.CloseErr
	}

	return nil
}

// SetBalance configures the balance for a specific address
func (m *MockBlockchainClient) SetBalance(address, balance string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Balances[address] = balance
}

// SetBalanceErr configures GetBalance to return an error
func (m *MockBlockchainClient) SetBalanceErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetBalanceErr = err
}

// Reset clears all mock data and call tracking
func (m *MockBlockchainClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Balances = make(map[string]string)
	m.GetBalanceCalls = make(map[string]int)
	m.ConnectErr = nil
	m.GetBalanceErr = nil
	m.CloseErr = nil
	m.ConnectCalls = 0
	m.CloseCalls = 0
}

// GetCallCount returns the number of times GetBalance was called for an address
func (m *MockBlockchainClient) GetCallCount(address string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.GetBalanceCalls[address]
}

// Verify that MockBlockchainClient implements blockchain.Client interface
var _ blockchain.Client = (*MockBlockchainClient)(nil)
