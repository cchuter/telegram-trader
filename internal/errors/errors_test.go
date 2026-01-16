package errors

import (
	stderrors "errors"
	"testing"
)

func TestBotError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *BotError
		wantCode ErrorCode
	}{
		{
			name:     "AUTH_001 error",
			err:      ErrAccessDenied(),
			wantCode: AUTH_001,
		},
		{
			name:     "WALLET_001 error",
			err:      ErrWalletNotConnected(),
			wantCode: WALLET_001,
		},
		{
			name:     "VAL_002 error",
			err:      ErrInvalidAmount(),
			wantCode: VAL_002,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Expected error code %s, got %s", tt.wantCode, tt.err.Code)
			}
			if tt.err.GetUserMessage() == "" {
				t.Error("User message should not be empty")
			}
		})
	}
}

func TestBotErrorWithParams(t *testing.T) {
	err := ErrInsufficientBalance("10.5", "5.0", "TON")
	if err.Code != BALANCE_001 {
		t.Errorf("Expected error code %s, got %s", BALANCE_001, err.Code)
	}

	msg := err.GetUserMessage()
	if msg == "" {
		t.Error("User message should not be empty")
	}
	// Message should contain the parameters
	// Note: This is a basic check - full string matching would be brittle
	t.Logf("Generated message: %s", msg)
}

func TestBotErrorWithCause(t *testing.T) {
	cause := stderrors.New("network timeout")
	err := ErrWalletTimeout(cause)

	if err.Cause != cause {
		t.Error("Cause should be preserved")
	}

	errorString := err.Error()
	if errorString == "" {
		t.Error("Error string should not be empty")
	}
	t.Logf("Error string: %s", errorString)
}

func TestAllErrorCodes(t *testing.T) {
	// Test that all error codes have user messages
	errorCodes := []ErrorCode{
		AUTH_001, AUTH_002,
		RATE_001,
		WALLET_001, WALLET_002, WALLET_003,
		BALANCE_001, BALANCE_002,
		SWAP_001, SWAP_002, SWAP_003, SWAP_004, SWAP_005,
		PRICE_001, PRICE_002,
		ARB_001, ARB_002, ARB_003,
		GALA_001, GALA_002,
		DB_001, DB_002,
		NET_001, NET_002,
		VAL_001, VAL_002, VAL_003,
		ENC_001,
		UNK_001,
	}

	for _, code := range errorCodes {
		t.Run(string(code), func(t *testing.T) {
			err := NewBotError(code, nil)
			if err.GetUserMessage() == "" {
				t.Errorf("Error code %s has no user message", code)
			}
		})
	}
}
