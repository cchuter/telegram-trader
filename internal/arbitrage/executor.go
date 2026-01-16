package arbitrage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cchuter/telegram-trader/internal/blockchain/ton"
	"github.com/cchuter/telegram-trader/internal/galachain/pb"
)

// ExecutionResult represents the result of an arbitrage execution
type ExecutionResult struct {
	// Success indicates if the arbitrage was fully successful
	Success bool
	// PartialSuccess indicates if one leg succeeded and one failed
	PartialSuccess bool
	// StonfiResult is the result from the ston.fi swap
	StonfiResult *ton.SwapResult
	// GswapResult is the result from the gswap swap
	GswapResult *pb.SwapResponse
	// Profit is the calculated profit after fees (in percentage)
	Profit float64
	// Error contains any error that occurred during execution
	Error error
	// ExecutionTime is the total time taken to execute both legs
	ExecutionTime time.Duration
}

// TonClient defines the interface for TON blockchain operations needed by the executor
type TonClient interface {
	SendTransaction(ctx context.Context, walletAddr string, privateKeyHex string, msg *ton.TransactionBuilder) (string, error)
	WaitForTransaction(ctx context.Context, walletAddr, txHash string, maxWaitTime time.Duration) (*ton.SwapResult, error)
}

// GalaSwapClient defines the interface for GalaChain swap operations needed by the executor
type GalaSwapClient interface {
	ExecuteSwap(ctx context.Context, req *pb.SwapRequest) (*pb.SwapResponse, error)
}

// Executor handles arbitrage execution
type Executor struct {
	tonClient  TonClient
	galaClient GalaSwapClient
	engine     *Engine
}

// NewExecutor creates a new arbitrage executor
func NewExecutor(tonClient TonClient, galaClient GalaSwapClient, engine *Engine) *Executor {
	return &Executor{
		tonClient:  tonClient,
		galaClient: galaClient,
		engine:     engine,
	}
}

// ExecuteArbitrage executes an arbitrage opportunity by running both legs concurrently
// It takes an opportunity, position size, and wallet credentials
// Returns an ExecutionResult with details about both transactions
func (e *Executor) ExecuteArbitrage(ctx context.Context, opportunity *Opportunity, position *PositionSize, userID int64, walletAddr string, privateKey string) *ExecutionResult {
	startTime := time.Now()
	result := &ExecutionResult{}

	// Create a WaitGroup to wait for both goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Channel for collecting errors
	errChan := make(chan error, 2)

	// Execute both legs concurrently based on direction
	switch opportunity.Direction {
	case BuyTonSellGala:
		// Buy TON on ston.fi (TON -> GALA), sell on gswap (GALA -> GTON)

		// Execute ston.fi swap in goroutine
		go func() {
			defer wg.Done()
			stonfiResult, err := e.executeStonfiSwap(ctx, position.TONAmount, walletAddr, privateKey)
			if err != nil {
				errChan <- fmt.Errorf("ston.fi swap failed: %w", err)
				return
			}
			result.StonfiResult = stonfiResult
		}()

		// Execute gswap swap in goroutine
		go func() {
			defer wg.Done()
			gswapResult, err := e.executeGswapSwap(ctx, position.GALAAmount, userID)
			if err != nil {
				errChan <- fmt.Errorf("gswap swap failed: %w", err)
				return
			}
			result.GswapResult = gswapResult
		}()

	case BuyGalaSellTon:
		// Buy GALA on gswap (GTON -> GALA), sell on ston.fi (GALA -> TON)

		// Execute gswap swap in goroutine
		go func() {
			defer wg.Done()
			gswapResult, err := e.executeGswapSwap(ctx, position.GALAAmount, userID)
			if err != nil {
				errChan <- fmt.Errorf("gswap swap failed: %w", err)
				return
			}
			result.GswapResult = gswapResult
		}()

		// Execute ston.fi swap in goroutine
		go func() {
			defer wg.Done()
			stonfiResult, err := e.executeStonfiSwap(ctx, position.TONAmount, walletAddr, privateKey)
			if err != nil {
				errChan <- fmt.Errorf("ston.fi swap failed: %w", err)
				return
			}
			result.StonfiResult = stonfiResult
		}()

	default:
		result.Error = fmt.Errorf("unknown arbitrage direction: %s", opportunity.Direction)
		return result
	}

	// Wait for both goroutines to complete
	wg.Wait()
	close(errChan)

	// Calculate execution time
	result.ExecutionTime = time.Since(startTime)

	// Collect errors from channel
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	// Determine success status
	if len(errors) == 0 {
		// Both legs succeeded
		result.Success = true
		result.Profit = e.calculateActualProfit(result.StonfiResult, result.GswapResult, opportunity)
	} else if len(errors) == 1 {
		// Partial success - one leg succeeded, one failed
		result.PartialSuccess = true
		result.Error = errors[0]
		// Calculate partial profit (may be negative due to one leg failing)
		result.Profit = e.calculateActualProfit(result.StonfiResult, result.GswapResult, opportunity)
	} else {
		// Both legs failed
		result.Success = false
		result.Error = fmt.Errorf("both legs failed: %v", errors)
	}

	return result
}

// executeStonfiSwap executes a swap on ston.fi
func (e *Executor) executeStonfiSwap(ctx context.Context, amount float64, walletAddr string, privateKey string) (*ton.SwapResult, error) {
	// For POC: Return mock result
	// In production, this would:
	// 1. Build swap transaction using TransactionBuilder
	// 2. Send transaction via tonClient.SendTransaction()
	// 3. Wait for confirmation via tonClient.WaitForTransaction()

	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	return &ton.SwapResult{
		TxHash:       fmt.Sprintf("stonfi_%d", time.Now().UnixNano()),
		Status:       "success",
		OutputAmount: fmt.Sprintf("%.2f", amount*850), // Mock output based on typical price
		Fee:          "0.05",
		Error:        "",
	}, nil
}

// executeGswapSwap executes a swap on gswap
func (e *Executor) executeGswapSwap(ctx context.Context, amount float64, userID int64) (*pb.SwapResponse, error) {
	// For POC: Return mock result
	// In production, this would call galaClient.ExecuteSwap() with proper SwapRequest

	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	return &pb.SwapResponse{
		TxHash:       fmt.Sprintf("gswap_%d", time.Now().UnixNano()),
		AmountIn:     fmt.Sprintf("%.2f", amount),
		AmountOut:    fmt.Sprintf("%.2f", amount*855), // Mock output based on typical price
		Fee:          "0.03",
		Status:       "success",
		ErrorMessage: "",
	}, nil
}

// calculateActualProfit calculates the actual profit after fees from completed transactions
func (e *Executor) calculateActualProfit(stonfiResult *ton.SwapResult, gswapResult *pb.SwapResponse, opportunity *Opportunity) float64 {
	// Handle partial success cases
	if stonfiResult == nil || gswapResult == nil {
		return 0.0
	}

	// Check if both succeeded
	if stonfiResult.Status != "success" || gswapResult.Status != "success" {
		return 0.0
	}

	// For POC: Return the opportunity spread as profit estimate
	// In production, this would:
	// 1. Parse actual output amounts from both transactions
	// 2. Subtract fees from both legs
	// 3. Calculate net profit as: (output_value - input_value) / input_value * 100

	// Return absolute value of spread as profit percentage
	if opportunity.Spread > 0 {
		return opportunity.Spread
	}
	return -opportunity.Spread
}
