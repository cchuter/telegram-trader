package validation

import (
	"math/big"
	"testing"
)

func TestValidateTonAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid EQ address",
			address: "EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV",
			wantErr: false,
		},
		{
			name:    "invalid prefix",
			address: "XQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV",
			wantErr: true,
		},
		{
			name:    "invalid base64",
			address: "EQ!!!invalid!!!",
			wantErr: true,
		},
		{
			name:    "too short",
			address: "EQBadmO",
			wantErr: true,
		},
		{
			name:    "empty string",
			address: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTonAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTonAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGalaAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid 64 char hex",
			address: "2954bd1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34",
			wantErr: false,
		},
		{
			name:    "too short",
			address: "2954bd1e38c2dff11a2ee8798e2f59b99c206e64",
			wantErr: true,
		},
		{
			name:    "too long",
			address: "2954bd1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34ff",
			wantErr: true,
		},
		{
			name:    "non-hex characters",
			address: "zzzzzz1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34",
			wantErr: true,
		},
		{
			name:    "empty string",
			address: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGalaAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGalaAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name      string
		amount    string
		wantErr   bool
		wantValue string // expected value as string for comparison
	}{
		{
			name:      "valid integer",
			amount:    "100",
			wantErr:   false,
			wantValue: "100",
		},
		{
			name:      "valid decimal",
			amount:    "123.456",
			wantErr:   false,
			wantValue: "123.456",
		},
		{
			name:      "18 decimals",
			amount:    "1.123456789012345678",
			wantErr:   false,
			wantValue: "1.123456789012345678",
		},
		{
			name:    "19 decimals",
			amount:  "1.1234567890123456789",
			wantErr: true,
		},
		{
			name:    "zero",
			amount:  "0",
			wantErr: true,
		},
		{
			name:    "negative",
			amount:  "-100",
			wantErr: true,
		},
		{
			name:    "invalid format",
			amount:  "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			amount:  "",
			wantErr: true,
		},
		{
			name:      "small decimal",
			amount:    "0.000001",
			wantErr:   false,
			wantValue: "0.000001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				want, _ := new(big.Float).SetString(tt.wantValue)
				if got.Cmp(want) != 0 {
					t.Errorf("ValidateAmount() = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestValidateTokenSymbol(t *testing.T) {
	tests := []struct {
		name    string
		symbol  string
		wantErr bool
	}{
		{
			name:    "TON uppercase",
			symbol:  "TON",
			wantErr: false,
		},
		{
			name:    "ton lowercase",
			symbol:  "ton",
			wantErr: false,
		},
		{
			name:    "GALA uppercase",
			symbol:  "GALA",
			wantErr: false,
		},
		{
			name:    "gala lowercase",
			symbol:  "gala",
			wantErr: false,
		},
		{
			name:    "GTON uppercase",
			symbol:  "GTON",
			wantErr: false,
		},
		{
			name:    "gton lowercase",
			symbol:  "gton",
			wantErr: false,
		},
		{
			name:    "invalid symbol",
			symbol:  "BTC",
			wantErr: true,
		},
		{
			name:    "empty string",
			symbol:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTokenSymbol(tt.symbol)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTokenSymbol() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "valid hex",
			s:    "deadbeef",
			want: true,
		},
		{
			name: "valid hex uppercase",
			s:    "DEADBEEF",
			want: true,
		},
		{
			name: "valid hex mixed case",
			s:    "DeAdBeEf",
			want: true,
		},
		{
			name: "invalid hex with letters",
			s:    "xyz",
			want: false,
		},
		{
			name: "empty string",
			s:    "",
			want: true, // empty string is valid hex (decodes to empty bytes)
		},
		{
			name: "odd length hex",
			s:    "abc",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHex(tt.s)
			if got != tt.want {
				t.Errorf("isHex() = %v, want %v", got, tt.want)
			}
		})
	}
}
