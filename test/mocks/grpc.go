package mocks

import (
	"context"
	"sync"

	"github.com/cchuter/telegram-trader/internal/galachain/pb"
)

// MockGRPCClient is a mock implementation of GalaChain gRPC client for testing
type MockGRPCClient struct {
	mu sync.RWMutex

	// Configurable responses
	Balances map[int64]*pb.BalanceResponse // userID -> balance response
	Prices   map[string]*pb.PriceResponse  // pair -> price response

	// Error simulation
	GetBalanceErr error
	GetPriceErr   error
	CloseErr      error

	// Call tracking
	GetBalanceCalls map[int64]int  // userID -> call count
	GetPriceCalls   map[string]int // pair -> call count
	CloseCalls      int
}

// NewMockGRPCClient creates a new mock gRPC client
func NewMockGRPCClient() *MockGRPCClient {
	return &MockGRPCClient{
		Balances:        make(map[int64]*pb.BalanceResponse),
		Prices:          make(map[string]*pb.PriceResponse),
		GetBalanceCalls: make(map[int64]int),
		GetPriceCalls:   make(map[string]int),
	}
}

// GetBalance retrieves the balance for a user
func (m *MockGRPCClient) GetBalance(ctx context.Context, userID int64) (*pb.BalanceResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetBalanceCalls[userID]++

	if m.GetBalanceErr != nil {
		return nil, m.GetBalanceErr
	}

	response, exists := m.Balances[userID]
	if !exists {
		// Return default balance if not configured
		return &pb.BalanceResponse{
			Balances: []*pb.TokenBalance{
				{
					Token:   "GALA",
					Balance: "0.0",
				},
				{
					Token:   "GTON",
					Balance: "0.0",
				},
			},
		}, nil
	}

	// Return a copy to prevent external mutations
	balanceCopy := &pb.BalanceResponse{
		Balances: make([]*pb.TokenBalance, len(response.Balances)),
	}
	for i, b := range response.Balances {
		balanceCopy.Balances[i] = &pb.TokenBalance{
			Token:   b.Token,
			Balance: b.Balance,
		}
	}

	return balanceCopy, nil
}

// GetPrice retrieves the price for a trading pair
func (m *MockGRPCClient) GetPrice(ctx context.Context, pair string) (*pb.PriceResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetPriceCalls[pair]++

	if m.GetPriceErr != nil {
		return nil, m.GetPriceErr
	}

	response, exists := m.Prices[pair]
	if !exists {
		// Return default price if not configured
		return &pb.PriceResponse{
			Price:      "0.0",
			Timestamp:  0,
			Bid:        "0.0",
			Ask:        "0.0",
			Volume_24H: "0.0",
		}, nil
	}

	// Return a copy to prevent external mutations
	return &pb.PriceResponse{
		Price:      response.Price,
		Timestamp:  response.Timestamp,
		Bid:        response.Bid,
		Ask:        response.Ask,
		Volume_24H: response.Volume_24H,
	}, nil
}

// Close closes the gRPC connection
func (m *MockGRPCClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CloseCalls++

	if m.CloseErr != nil {
		return m.CloseErr
	}

	return nil
}

// SetBalance configures the balance response for a specific user
func (m *MockGRPCClient) SetBalance(userID int64, balances []*pb.TokenBalance) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Balances[userID] = &pb.BalanceResponse{
		Balances: balances,
	}
}

// SetPrice configures the price response for a specific pair
func (m *MockGRPCClient) SetPrice(pair string, price *pb.PriceResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Prices[pair] = price
}

// SetGetBalanceErr configures GetBalance to return an error
func (m *MockGRPCClient) SetGetBalanceErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetBalanceErr = err
}

// SetGetPriceErr configures GetPrice to return an error
func (m *MockGRPCClient) SetGetPriceErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.GetPriceErr = err
}

// Reset clears all mock data and call tracking
func (m *MockGRPCClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Balances = make(map[int64]*pb.BalanceResponse)
	m.Prices = make(map[string]*pb.PriceResponse)
	m.GetBalanceCalls = make(map[int64]int)
	m.GetPriceCalls = make(map[string]int)
	m.GetBalanceErr = nil
	m.GetPriceErr = nil
	m.CloseErr = nil
	m.CloseCalls = 0
}

// GetBalanceCallCount returns the number of times GetBalance was called for a user
func (m *MockGRPCClient) GetBalanceCallCount(userID int64) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.GetBalanceCalls[userID]
}

// GetPriceCallCount returns the number of times GetPrice was called for a pair
func (m *MockGRPCClient) GetPriceCallCount(pair string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.GetPriceCalls[pair]
}
