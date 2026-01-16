package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TradeType represents the type of trade operation
type TradeType string

const (
	// TradeTypeSwap for token swap operations
	TradeTypeSwap TradeType = "swap"
	// TradeTypeArbitrage for arbitrage operations
	TradeTypeArbitrage TradeType = "arbitrage"
)

// TradeStatus represents the status of a trade
type TradeStatus string

const (
	// TradeStatusSuccess for successful trades
	TradeStatusSuccess TradeStatus = "success"
	// TradeStatusFailed for failed trades
	TradeStatusFailed TradeStatus = "failed"
	// TradeStatusPending for pending trades
	TradeStatusPending TradeStatus = "pending"
)

// TradeLogEntry represents a single trade log entry
type TradeLogEntry struct {
	Timestamp       string      `json:"timestamp"`
	UserID          int64       `json:"userId"`
	Type            TradeType   `json:"type"`
	Chain           string      `json:"chain"`
	FromToken       string      `json:"fromToken"`
	ToToken         string      `json:"toToken"`
	AmountIn        string      `json:"amountIn"`
	AmountOut       string      `json:"amountOut"`
	Fee             string      `json:"fee"`
	TxHash          string      `json:"txHash,omitempty"`
	Status          TradeStatus `json:"status"`
	ExecutionTimeMs int64       `json:"executionTimeMs"`
	ErrorMessage    string      `json:"errorMessage,omitempty"`
}

// TradeLogger handles logging trade operations to JSONL files
type TradeLogger struct {
	logDir   string
	file     *os.File
	encoder  *json.Encoder
	mu       sync.Mutex
	filename string
}

// NewTradeLogger creates a new TradeLogger instance
func NewTradeLogger(logDir string) (*TradeLogger, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate log filename with date
	filename := filepath.Join(logDir, fmt.Sprintf("trades-%s.jsonl", time.Now().UTC().Format("2006-01-02")))

	// Open log file in append mode
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open trade log file: %w", err)
	}

	return &TradeLogger{
		logDir:   logDir,
		file:     file,
		encoder:  json.NewEncoder(file),
		filename: filename,
	}, nil
}

// LogSwap logs a swap trade operation
func (tl *TradeLogger) LogSwap(userID int64, chain, fromToken, toToken, amountIn, amountOut, fee, txHash string, status TradeStatus, executionTimeMs int64, errorMsg string) error {
	entry := TradeLogEntry{
		Timestamp:       time.Now().UTC().Format(time.RFC3339Nano),
		UserID:          userID,
		Type:            TradeTypeSwap,
		Chain:           chain,
		FromToken:       fromToken,
		ToToken:         toToken,
		AmountIn:        amountIn,
		AmountOut:       amountOut,
		Fee:             fee,
		TxHash:          txHash,
		Status:          status,
		ExecutionTimeMs: executionTimeMs,
		ErrorMessage:    errorMsg,
	}
	return tl.writeEntry(entry)
}

// LogArbitrage logs an arbitrage trade operation
func (tl *TradeLogger) LogArbitrage(userID int64, fromToken, toToken, amountIn, amountOut, fee string, status TradeStatus, executionTimeMs int64, errorMsg string) error {
	entry := TradeLogEntry{
		Timestamp:       time.Now().UTC().Format(time.RFC3339Nano),
		UserID:          userID,
		Type:            TradeTypeArbitrage,
		Chain:           "multi", // Arbitrage involves multiple chains
		FromToken:       fromToken,
		ToToken:         toToken,
		AmountIn:        amountIn,
		AmountOut:       amountOut,
		Fee:             fee,
		Status:          status,
		ExecutionTimeMs: executionTimeMs,
		ErrorMessage:    errorMsg,
	}
	return tl.writeEntry(entry)
}

// writeEntry writes a trade entry to the log file
func (tl *TradeLogger) writeEntry(entry TradeLogEntry) error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	// Check if we need to rotate the log file (new day)
	expectedFilename := filepath.Join(tl.logDir, fmt.Sprintf("trades-%s.jsonl", time.Now().UTC().Format("2006-01-02")))
	if expectedFilename != tl.filename {
		if err := tl.rotate(expectedFilename); err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	// Write the entry
	if err := tl.encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to write trade log entry: %w", err)
	}

	return nil
}

// rotate closes the current log file and opens a new one
func (tl *TradeLogger) rotate(newFilename string) error {
	// Close current file
	if tl.file != nil {
		if err := tl.file.Close(); err != nil {
			return fmt.Errorf("failed to close current log file: %w", err)
		}
	}

	// Open new file
	file, err := os.OpenFile(newFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %w", err)
	}

	tl.file = file
	tl.encoder = json.NewEncoder(file)
	tl.filename = newFilename

	return nil
}

// Close closes the trade logger and its underlying file
func (tl *TradeLogger) Close() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.file != nil {
		return tl.file.Close()
	}
	return nil
}
