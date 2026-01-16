package wallet

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

// generateTestKey generates a random 32-byte key for testing
func generateTestKey() string {
	key := make([]byte, 32)
	rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}

func TestNewEncryptionService(t *testing.T) {
	tests := []struct {
		name      string
		masterKey string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid 32-byte key",
			masterKey: generateTestKey(),
			wantErr:   false,
		},
		{
			name:      "invalid base64",
			masterKey: "not-valid-base64!!!",
			wantErr:   true,
			errMsg:    "failed to decode master key",
		},
		{
			name:      "wrong key length - 16 bytes",
			masterKey: base64.StdEncoding.EncodeToString(make([]byte, 16)),
			wantErr:   true,
			errMsg:    "invalid key length",
		},
		{
			name:      "wrong key length - 24 bytes",
			masterKey: base64.StdEncoding.EncodeToString(make([]byte, 24)),
			wantErr:   true,
			errMsg:    "invalid key length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewEncryptionService(tt.masterKey)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEncryptionService() expected error but got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewEncryptionService() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("NewEncryptionService() unexpected error = %v", err)
				return
			}
			if svc == nil {
				t.Errorf("NewEncryptionService() returned nil service")
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	masterKey := generateTestKey()
	svc, err := NewEncryptionService(masterKey)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "Hello, World!",
		},
		{
			name:      "private key format",
			plaintext: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "long text",
			plaintext: strings.Repeat("a", 1000),
		},
		{
			name:      "special characters",
			plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "unicode characters",
			plaintext: "Hello 世界 🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := svc.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Verify ciphertext is not empty
			if ciphertext == "" {
				t.Errorf("Encrypt() returned empty ciphertext")
			}

			// Verify ciphertext is different from plaintext
			if ciphertext == tt.plaintext {
				t.Errorf("Encrypt() ciphertext equals plaintext")
			}

			// Decrypt
			decrypted, err := svc.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			// Verify decrypted text matches original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptionUniqueness(t *testing.T) {
	masterKey := generateTestKey()
	svc, err := NewEncryptionService(masterKey)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	plaintext := "test private key"

	// Encrypt the same plaintext multiple times
	ciphertext1, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("First Encrypt() error = %v", err)
	}

	ciphertext2, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Second Encrypt() error = %v", err)
	}

	// Verify that ciphertexts are different (due to random nonce)
	if ciphertext1 == ciphertext2 {
		t.Errorf("Encrypt() produced same ciphertext for same plaintext - nonce not unique")
	}

	// Verify both decrypt to the same plaintext
	decrypted1, err := svc.Decrypt(ciphertext1)
	if err != nil {
		t.Fatalf("First Decrypt() error = %v", err)
	}

	decrypted2, err := svc.Decrypt(ciphertext2)
	if err != nil {
		t.Fatalf("Second Decrypt() error = %v", err)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Errorf("Decrypt() failed to recover original plaintext")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	masterKey := generateTestKey()
	svc, err := NewEncryptionService(masterKey)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "invalid base64",
			ciphertext: "not-valid-base64!!!",
			wantErr:    true,
			errMsg:     "failed to decode ciphertext",
		},
		{
			name:       "empty string",
			ciphertext: "",
			wantErr:    true,
			errMsg:     "ciphertext too short",
		},
		{
			name:       "too short ciphertext",
			ciphertext: base64.StdEncoding.EncodeToString([]byte("short")),
			wantErr:    true,
			errMsg:     "ciphertext too short",
		},
		{
			name:       "corrupted ciphertext",
			ciphertext: base64.StdEncoding.EncodeToString(make([]byte, 32)),
			wantErr:    true,
			errMsg:     "failed to decrypt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Decrypt(tt.ciphertext)
			if !tt.wantErr {
				if err != nil {
					t.Errorf("Decrypt() unexpected error = %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("Decrypt() expected error but got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Decrypt() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

func TestDecryptWithDifferentKey(t *testing.T) {
	// Create two services with different keys
	masterKey1 := generateTestKey()
	svc1, err := NewEncryptionService(masterKey1)
	if err != nil {
		t.Fatalf("Failed to create first encryption service: %v", err)
	}

	masterKey2 := generateTestKey()
	svc2, err := NewEncryptionService(masterKey2)
	if err != nil {
		t.Fatalf("Failed to create second encryption service: %v", err)
	}

	plaintext := "secret data"

	// Encrypt with first service
	ciphertext, err := svc1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Try to decrypt with second service (different key)
	_, err = svc2.Decrypt(ciphertext)
	if err == nil {
		t.Errorf("Decrypt() with different key should fail but succeeded")
	}
}

func TestEncryptionServiceIsReusable(t *testing.T) {
	masterKey := generateTestKey()
	svc, err := NewEncryptionService(masterKey)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	// Test that the same service can be used multiple times
	for i := 0; i < 10; i++ {
		plaintext := "test data " + string(rune(i))

		ciphertext, err := svc.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt() iteration %d error = %v", i, err)
		}

		decrypted, err := svc.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Decrypt() iteration %d error = %v", i, err)
		}

		if decrypted != plaintext {
			t.Errorf("Iteration %d: Decrypt() = %v, want %v", i, decrypted, plaintext)
		}
	}
}
