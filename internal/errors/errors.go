package errors

import "fmt"

// ErrorCode represents a specific error type in the bot
type ErrorCode string

// Error codes organized by category (from design.md Error Handling section)
const (
	// Authentication errors
	AUTH_001 ErrorCode = "AUTH_001" // Access denied
	AUTH_002 ErrorCode = "AUTH_002" // Session expired

	// Rate limiting errors
	RATE_001 ErrorCode = "RATE_001" // Too many requests

	// Wallet connection errors
	WALLET_001 ErrorCode = "WALLET_001" // Wallet not connected
	WALLET_002 ErrorCode = "WALLET_002" // Wallet connection timeout
	WALLET_003 ErrorCode = "WALLET_003" // Invalid wallet address format

	// Balance check errors
	BALANCE_001 ErrorCode = "BALANCE_001" // Insufficient balance
	BALANCE_002 ErrorCode = "BALANCE_002" // Failed to fetch balance

	// Swap simulation errors
	SWAP_001 ErrorCode = "SWAP_001" // Failed to simulate swap

	// Swap execution errors
	SWAP_002 ErrorCode = "SWAP_002" // Slippage too high
	SWAP_003 ErrorCode = "SWAP_003" // Transaction failed
	SWAP_004 ErrorCode = "SWAP_004" // Transaction timeout
	SWAP_005 ErrorCode = "SWAP_005" // Minimum balance not met

	// Price check errors
	PRICE_001 ErrorCode = "PRICE_001" // Failed to fetch price from exchange
	PRICE_002 ErrorCode = "PRICE_002" // Price data unavailable

	// Arbitrage errors
	ARB_001 ErrorCode = "ARB_001" // No arbitrage opportunity found
	ARB_002 ErrorCode = "ARB_002" // Arbitrage failed
	ARB_003 ErrorCode = "ARB_003" // Arbitrage partially complete

	// GalaChain service errors
	GALA_001 ErrorCode = "GALA_001" // GalaChain service unavailable
	GALA_002 ErrorCode = "GALA_002" // gswap rate limit reached

	// Database errors
	DB_001 ErrorCode = "DB_001" // Failed to save data
	DB_002 ErrorCode = "DB_002" // Data retrieval failed

	// Network errors
	NET_001 ErrorCode = "NET_001" // Network error
	NET_002 ErrorCode = "NET_002" // Service timeout

	// Validation errors
	VAL_001 ErrorCode = "VAL_001" // Invalid command format
	VAL_002 ErrorCode = "VAL_002" // Invalid amount
	VAL_003 ErrorCode = "VAL_003" // Invalid token symbol

	// Encryption errors
	ENC_001 ErrorCode = "ENC_001" // Failed to decrypt wallet key

	// Unknown errors
	UNK_001 ErrorCode = "UNK_001" // Unexpected error occurred
)

// BotError represents an error that can be shown to users
type BotError struct {
	Code          ErrorCode
	Message       string
	Params        map[string]string // For parameterized messages
	Cause         error             // Original error for logging
	CorrelationID string            // Correlation ID for request tracing
}

// Error implements the error interface
func (e *BotError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewBotError creates a new BotError
func NewBotError(code ErrorCode, cause error) *BotError {
	return &BotError{
		Code:    code,
		Message: getUserMessage(code, nil),
		Params:  make(map[string]string),
		Cause:   cause,
	}
}

// NewBotErrorWithParams creates a new BotError with parameterized message
func NewBotErrorWithParams(code ErrorCode, params map[string]string, cause error) *BotError {
	return &BotError{
		Code:    code,
		Message: getUserMessage(code, params),
		Params:  params,
		Cause:   cause,
	}
}

// GetUserMessage returns the user-friendly message for this error
// If correlation ID is set, it's included in the message for support
func (e *BotError) GetUserMessage() string {
	if e.CorrelationID != "" {
		return fmt.Sprintf("%s\n\nSupport ID: %s", e.Message, e.CorrelationID)
	}
	return e.Message
}

// SetCorrelationID sets the correlation ID for this error
func (e *BotError) SetCorrelationID(correlationID string) *BotError {
	e.CorrelationID = correlationID
	return e
}

// getUserMessage returns a user-friendly message for the given error code
// Messages are designed to be clear, non-technical, and actionable
func getUserMessage(code ErrorCode, params map[string]string) string {
	// Default messages (can be overridden with params)
	messages := map[ErrorCode]string{
		// Authentication
		AUTH_001: "Access denied. Contact @admin for access.",
		AUTH_002: "Session expired. Please use /start to reconnect.",

		// Rate limiting
		RATE_001: "Too many requests. Please wait 60 seconds.",

		// Wallet connection
		WALLET_001: "Wallet not connected. Use /wallet to connect.",
		WALLET_002: "Wallet connection timeout. Please try again.",
		WALLET_003: "Invalid wallet address format.",

		// Balance check
		BALANCE_001: fmt.Sprintf("Insufficient balance. Required: %s %s, Available: %s",
			getParam(params, "required", "X"),
			getParam(params, "token", "TOKEN"),
			getParam(params, "available", "Y")),
		BALANCE_002: "Failed to fetch balance. Please try again.",

		// Swap simulation
		SWAP_001: "Failed to simulate swap. Exchange may be unavailable.",

		// Swap execution
		SWAP_002: fmt.Sprintf("Slippage too high (%s%%). Max allowed: %s%%. Adjust amount or slippage.",
			getParam(params, "slippage", "X"),
			getParam(params, "max_slippage", "Y")),
		SWAP_003: fmt.Sprintf("Transaction failed. Reason: %s",
			getParam(params, "reason", "Unknown error")),
		SWAP_004: "Transaction timeout. Check /orders for status.",
		SWAP_005: fmt.Sprintf("Minimum balance not met. Must keep %s %s.",
			getParam(params, "min", "MIN"),
			getParam(params, "token", "TOKEN")),

		// Price check
		PRICE_001: fmt.Sprintf("Failed to fetch price from %s. Retrying...",
			getParam(params, "exchange", "exchange")),
		PRICE_002: "Price data unavailable. Please try again later.",

		// Arbitrage
		ARB_001: fmt.Sprintf("No arbitrage opportunity found (spread: %s%% < threshold: %s%%).",
			getParam(params, "spread", "X"),
			getParam(params, "threshold", "Y")),
		ARB_002: fmt.Sprintf("Arbitrage failed: %s leg unsuccessful. Reason: %s",
			getParam(params, "leg", "LEG"),
			getParam(params, "reason", "REASON")),
		ARB_003: "Arbitrage partially complete. Buy succeeded, sell failed. Manual intervention needed.",

		// GalaChain service
		GALA_001: "GalaChain service unavailable. TON operations still available.",
		GALA_002: fmt.Sprintf("gswap rate limit reached. Please wait %s seconds.",
			getParam(params, "wait_time", "X")),

		// Database
		DB_001: "Failed to save data. Please try again.",
		DB_002: "Data retrieval failed. Please try again later.",

		// Network
		NET_001: "Network error. Retrying...",
		NET_002: "Service timeout. Please try again.",

		// Validation
		VAL_001: "Invalid command format. Use /help for examples.",
		VAL_002: "Invalid amount. Must be positive number.",
		VAL_003: "Invalid token symbol. Supported: TON, GALA, GTON.",

		// Encryption
		ENC_001: "Failed to decrypt wallet key. Please reconnect wallet.",

		// Unknown
		UNK_001: fmt.Sprintf("An unexpected error occurred. Support has been notified. Error ID: %s",
			getParam(params, "correlation_id", "CORRELATION_ID")),
	}

	msg, ok := messages[code]
	if !ok {
		return "An error occurred. Please try again."
	}
	return msg
}

// getParam safely gets a parameter value or returns a default
func getParam(params map[string]string, key, defaultValue string) string {
	if params == nil {
		return defaultValue
	}
	if val, ok := params[key]; ok {
		return val
	}
	return defaultValue
}

// Helper functions to create common errors

// ErrAccessDenied creates an AUTH_001 error
func ErrAccessDenied() *BotError {
	return NewBotError(AUTH_001, nil)
}

// ErrSessionExpired creates an AUTH_002 error
func ErrSessionExpired() *BotError {
	return NewBotError(AUTH_002, nil)
}

// ErrRateLimitExceeded creates a RATE_001 error
func ErrRateLimitExceeded() *BotError {
	return NewBotError(RATE_001, nil)
}

// ErrWalletNotConnected creates a WALLET_001 error
func ErrWalletNotConnected() *BotError {
	return NewBotError(WALLET_001, nil)
}

// ErrWalletTimeout creates a WALLET_002 error
func ErrWalletTimeout(cause error) *BotError {
	return NewBotError(WALLET_002, cause)
}

// ErrInvalidWalletAddress creates a WALLET_003 error
func ErrInvalidWalletAddress() *BotError {
	return NewBotError(WALLET_003, nil)
}

// ErrInsufficientBalance creates a BALANCE_001 error
func ErrInsufficientBalance(required, available, token string) *BotError {
	return NewBotErrorWithParams(BALANCE_001, map[string]string{
		"required":  required,
		"available": available,
		"token":     token,
	}, nil)
}

// ErrBalanceFetchFailed creates a BALANCE_002 error
func ErrBalanceFetchFailed(cause error) *BotError {
	return NewBotError(BALANCE_002, cause)
}

// ErrSwapSimulationFailed creates a SWAP_001 error
func ErrSwapSimulationFailed(cause error) *BotError {
	return NewBotError(SWAP_001, cause)
}

// ErrSlippageTooHigh creates a SWAP_002 error
func ErrSlippageTooHigh(slippage, maxSlippage string) *BotError {
	return NewBotErrorWithParams(SWAP_002, map[string]string{
		"slippage":     slippage,
		"max_slippage": maxSlippage,
	}, nil)
}

// ErrTransactionFailed creates a SWAP_003 error
func ErrTransactionFailed(reason string, cause error) *BotError {
	return NewBotErrorWithParams(SWAP_003, map[string]string{
		"reason": reason,
	}, cause)
}

// ErrTransactionTimeout creates a SWAP_004 error
func ErrTransactionTimeout() *BotError {
	return NewBotError(SWAP_004, nil)
}

// ErrMinBalanceNotMet creates a SWAP_005 error
func ErrMinBalanceNotMet(min, token string) *BotError {
	return NewBotErrorWithParams(SWAP_005, map[string]string{
		"min":   min,
		"token": token,
	}, nil)
}

// ErrPriceFetchFailed creates a PRICE_001 error
func ErrPriceFetchFailed(exchange string, cause error) *BotError {
	return NewBotErrorWithParams(PRICE_001, map[string]string{
		"exchange": exchange,
	}, cause)
}

// ErrPriceDataUnavailable creates a PRICE_002 error
func ErrPriceDataUnavailable() *BotError {
	return NewBotError(PRICE_002, nil)
}

// ErrNoArbitrageOpportunity creates an ARB_001 error
func ErrNoArbitrageOpportunity(spread, threshold string) *BotError {
	return NewBotErrorWithParams(ARB_001, map[string]string{
		"spread":    spread,
		"threshold": threshold,
	}, nil)
}

// ErrArbitrageFailed creates an ARB_002 error
func ErrArbitrageFailed(leg, reason string, cause error) *BotError {
	return NewBotErrorWithParams(ARB_002, map[string]string{
		"leg":    leg,
		"reason": reason,
	}, cause)
}

// ErrArbitragePartiallyComplete creates an ARB_003 error
func ErrArbitragePartiallyComplete() *BotError {
	return NewBotError(ARB_003, nil)
}

// ErrGalaChainUnavailable creates a GALA_001 error
func ErrGalaChainUnavailable(cause error) *BotError {
	return NewBotError(GALA_001, cause)
}

// ErrGalaChainRateLimit creates a GALA_002 error
func ErrGalaChainRateLimit(waitTime string) *BotError {
	return NewBotErrorWithParams(GALA_002, map[string]string{
		"wait_time": waitTime,
	}, nil)
}

// ErrDatabaseSaveFailed creates a DB_001 error
func ErrDatabaseSaveFailed(cause error) *BotError {
	return NewBotError(DB_001, cause)
}

// ErrDatabaseRetrievalFailed creates a DB_002 error
func ErrDatabaseRetrievalFailed(cause error) *BotError {
	return NewBotError(DB_002, cause)
}

// ErrNetworkError creates a NET_001 error
func ErrNetworkError(cause error) *BotError {
	return NewBotError(NET_001, cause)
}

// ErrServiceTimeout creates a NET_002 error
func ErrServiceTimeout(cause error) *BotError {
	return NewBotError(NET_002, cause)
}

// ErrInvalidCommandFormat creates a VAL_001 error
func ErrInvalidCommandFormat() *BotError {
	return NewBotError(VAL_001, nil)
}

// ErrInvalidAmount creates a VAL_002 error
func ErrInvalidAmount() *BotError {
	return NewBotError(VAL_002, nil)
}

// ErrInvalidTokenSymbol creates a VAL_003 error
func ErrInvalidTokenSymbol() *BotError {
	return NewBotError(VAL_003, nil)
}

// ErrEncryptionFailed creates an ENC_001 error
func ErrEncryptionFailed(cause error) *BotError {
	return NewBotError(ENC_001, cause)
}

// ErrUnknown creates a UNK_001 error
func ErrUnknown(correlationID string, cause error) *BotError {
	return NewBotErrorWithParams(UNK_001, map[string]string{
		"correlation_id": correlationID,
	}, cause)
}
