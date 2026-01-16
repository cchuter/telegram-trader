package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/cchuter/telegram-trader/internal/dex"
)

// MockDEXClient is a mock implementation of dex.Client for testing
type MockDEXClient struct {
	mu sync.RWMutex

	// Configurable responses
	Simulations map[string]*dex.SwapSimulation // key: "fromToken:toToken:amount"

	// Error simulation
	SimulateSwapErr error
	CloseErr        error

	// Call tracking
	SimulateSwapCalls map[string]int // key: "fromToken:toToken:amount"
	CloseCalls        int
}

// NewMockDEXClient creates a new mock DEX client
func NewMockDEXClient() *MockDEXClient {
	return &MockDEXClient{
		Simulations:       make(map[string]*dex.SwapSimulation),
		SimulateSwapCalls: make(map[string]int),
	}
}

// SimulateSwap simulates a token swap and returns expected output
func (m *MockDEXClient) SimulateSwap(ctx context.Context, fromToken, toToken, amount string) (*dex.SwapSimulation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s", fromToken, toToken, amount)
	m.SimulateSwapCalls[key]++

	if m.SimulateSwapErr != nil {
		return nil, m.SimulateSwapErr
	}

	simulation, exists := m.Simulations[key]
	if !exists {
		// Return default simulation if not configured
		return &dex.SwapSimulation{
			OutputAmount: "1000.0",
			Fee:          "0.1",
			Slippage:     1.0,
			PriceImpact:  0.05,
		}, nil
	}

	// Return a copy to prevent external mutations
	return &dex.SwapSimulation{
		OutputAmount: simulation.OutputAmount,
		Fee:          simulation.Fee,
		Slippage:     simulation.Slippage,
		PriceImpact:  simulation.PriceImpact,
	}, nil
}

// Close closes any open connections
func (m *MockDEXClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CloseCalls++

	if m.CloseErr != nil {
		return m.CloseErr
	}

	return nil
}

// SetSimulation configures the simulation result for a specific swap
func (m *MockDEXClient) SetSimulation(fromToken, toToken, amount string, simulation *dex.SwapSimulation) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s", fromToken, toToken, amount)
	m.Simulations[key] = simulation
}

// SetSimulateSwapErr configures SimulateSwap to return an error
func (m *MockDEXClient) SetSimulateSwapErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SimulateSwapErr = err
}

// Reset clears all mock data and call tracking
func (m *MockDEXClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Simulations = make(map[string]*dex.SwapSimulation)
	m.SimulateSwapCalls = make(map[string]int)
	m.SimulateSwapErr = nil
	m.CloseErr = nil
	m.CloseCalls = 0
}

// GetCallCount returns the number of times SimulateSwap was called for a specific swap
func (m *MockDEXClient) GetCallCount(fromToken, toToken, amount string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s:%s:%s", fromToken, toToken, amount)
	return m.SimulateSwapCalls[key]
}

// Verify that MockDEXClient implements dex.Client interface
var _ dex.Client = (*MockDEXClient)(nil)
