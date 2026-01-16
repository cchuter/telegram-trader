package arbitrage

import (
	"testing"
)

func TestCalculatePositionSize_BuyTonSellGala_Valid(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has 100 TON and 1000 GALA
	// Should use 50% of GALA (500 GALA) for the trade
	// After trade: 100 TON (meets 1.5 TON minimum), 500 GALA (meets 15 GALA minimum)
	tonBalance := 100.0
	galaBalance := 1000.0

	result := calc.CalculatePositionSize(tonBalance, galaBalance, BuyTonSellGala)

	if !result.Valid {
		t.Fatalf("Expected valid position, got invalid: %s", result.Reason)
	}
	if result.GALAAmount != 500.0 {
		t.Errorf("Expected GALA amount 500.0, got %.2f", result.GALAAmount)
	}
	if result.TONAmount != 0 {
		t.Errorf("Expected TON amount 0 (will receive from swap), got %.2f", result.TONAmount)
	}

	// Verify minimum balance constraint
	remainingGALA := galaBalance - result.GALAAmount
	if remainingGALA < MinGALABalance*SafetyMargin {
		t.Errorf("Remaining GALA %.2f is below safety minimum %.2f", remainingGALA, MinGALABalance*SafetyMargin)
	}
}

func TestCalculatePositionSize_BuyTonSellGala_InsufficientGALA(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has 10 TON but only 10 GALA (below 15 GALA safety minimum)
	tonBalance := 10.0
	galaBalance := 10.0

	result := calc.CalculatePositionSize(tonBalance, galaBalance, BuyTonSellGala)

	if result.Valid {
		t.Fatal("Expected invalid position due to insufficient GALA, got valid")
	}
	if result.Reason == "" {
		t.Error("Expected reason for invalid position, got empty string")
	}
}

func TestCalculatePositionSize_BuyTonSellGala_InsufficientTON(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has only 1 TON (below 1.5 TON safety minimum) and 1000 GALA
	tonBalance := 1.0
	galaBalance := 1000.0

	result := calc.CalculatePositionSize(tonBalance, galaBalance, BuyTonSellGala)

	if result.Valid {
		t.Fatal("Expected invalid position due to insufficient TON, got valid")
	}
	if result.Reason == "" {
		t.Error("Expected reason for invalid position, got empty string")
	}
}

func TestCalculatePositionSize_BuyGalaSellTon_Valid(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has 100 TON and 1000 GALA
	// Should use 50% of TON (50 TON) for the trade
	// After trade: 50 TON (meets 1.5 TON minimum), 1000 GALA (meets 15 GALA minimum)
	tonBalance := 100.0
	galaBalance := 1000.0

	result := calc.CalculatePositionSize(tonBalance, galaBalance, BuyGalaSellTon)

	if !result.Valid {
		t.Fatalf("Expected valid position, got invalid: %s", result.Reason)
	}
	if result.TONAmount != 50.0 {
		t.Errorf("Expected TON amount 50.0, got %.2f", result.TONAmount)
	}
	if result.GALAAmount != 0 {
		t.Errorf("Expected GALA amount 0 (will receive from swap), got %.2f", result.GALAAmount)
	}

	// Verify minimum balance constraint
	remainingTON := tonBalance - result.TONAmount
	if remainingTON < MinTONBalance*SafetyMargin {
		t.Errorf("Remaining TON %.2f is below safety minimum %.2f", remainingTON, MinTONBalance*SafetyMargin)
	}
}

func TestCalculatePositionSize_BuyGalaSellTon_InsufficientTON(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has only 1 TON (below 1.5 TON safety minimum) and 1000 GALA
	tonBalance := 1.0
	galaBalance := 1000.0

	result := calc.CalculatePositionSize(tonBalance, galaBalance, BuyGalaSellTon)

	if result.Valid {
		t.Fatal("Expected invalid position due to insufficient TON, got valid")
	}
	if result.Reason == "" {
		t.Error("Expected reason for invalid position, got empty string")
	}
}

func TestCalculatePositionSize_BuyGalaSellTon_InsufficientGALA(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has 100 TON but only 10 GALA (below 15 GALA safety minimum)
	tonBalance := 100.0
	galaBalance := 10.0

	result := calc.CalculatePositionSize(tonBalance, galaBalance, BuyGalaSellTon)

	if result.Valid {
		t.Fatal("Expected invalid position due to insufficient GALA, got valid")
	}
	if result.Reason == "" {
		t.Error("Expected reason for invalid position, got empty string")
	}
}

func TestCalculatePositionSize_EdgeCase_ExactlyAtMinimum(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has exactly at safety minimums: 1.5 TON and 15 GALA
	// Should return invalid because we can't trade any amount
	tonBalance := 1.5
	galaBalance := 15.0

	result1 := calc.CalculatePositionSize(tonBalance, galaBalance, BuyTonSellGala)
	if result1.Valid {
		t.Error("Expected invalid position when at exact minimum for BuyTonSellGala")
	}

	result2 := calc.CalculatePositionSize(tonBalance, galaBalance, BuyGalaSellTon)
	if result2.Valid {
		t.Error("Expected invalid position when at exact minimum for BuyGalaSellTon")
	}
}

func TestCalculatePositionSize_EdgeCase_JustAboveMinimum(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has just above safety minimums: 3 TON and 30 GALA
	// Should use 50% (1.5 TON or 15 GALA) but limited by available after minimums
	tonBalance := 3.0
	galaBalance := 30.0

	result1 := calc.CalculatePositionSize(tonBalance, galaBalance, BuyTonSellGala)
	if !result1.Valid {
		t.Errorf("Expected valid position for BuyTonSellGala, got invalid: %s", result1.Reason)
	}
	// 50% of 30 GALA = 15, but available after 15 GALA minimum = 15, so min(15, 15) = 15
	if result1.GALAAmount > galaBalance-MinGALABalance*SafetyMargin {
		t.Errorf("GALA amount %.2f exceeds available after minimums", result1.GALAAmount)
	}

	result2 := calc.CalculatePositionSize(tonBalance, galaBalance, BuyGalaSellTon)
	if !result2.Valid {
		t.Errorf("Expected valid position for BuyGalaSellTon, got invalid: %s", result2.Reason)
	}
	// 50% of 3 TON = 1.5, but available after 1.5 TON minimum = 1.5, so min(1.5, 1.5) = 1.5
	if result2.TONAmount > tonBalance-MinTONBalance*SafetyMargin {
		t.Errorf("TON amount %.2f exceeds available after minimums", result2.TONAmount)
	}
}

func TestCalculatePositionSize_SafetyMarginApplied(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has 10 TON and 100 GALA
	// Verify safety margin is applied (1.5x minimum balances)
	tonBalance := 10.0
	galaBalance := 100.0

	result := calc.CalculatePositionSize(tonBalance, galaBalance, BuyTonSellGala)

	if !result.Valid {
		t.Fatalf("Expected valid position, got invalid: %s", result.Reason)
	}

	// After spending GALA, should have at least MinGALABalance * SafetyMargin = 15 GALA
	remainingGALA := galaBalance - result.GALAAmount
	expectedMin := MinGALABalance * SafetyMargin
	if remainingGALA < expectedMin {
		t.Errorf("Safety margin not applied. Remaining GALA %.2f < expected min %.2f", remainingGALA, expectedMin)
	}
}

func TestCalculatePositionSize_FiftyPercentRule(t *testing.T) {
	calc := NewCalculator()

	// Setup: User has plenty of balance: 100 TON and 1000 GALA
	tonBalance := 100.0
	galaBalance := 1000.0

	result1 := calc.CalculatePositionSize(tonBalance, galaBalance, BuyTonSellGala)
	if !result1.Valid {
		t.Fatalf("Expected valid position, got invalid: %s", result1.Reason)
	}
	// Should use exactly 50% of GALA
	expectedGALA := galaBalance * PositionSizePercent
	if result1.GALAAmount != expectedGALA {
		t.Errorf("Expected %.2f GALA (50%%), got %.2f", expectedGALA, result1.GALAAmount)
	}

	result2 := calc.CalculatePositionSize(tonBalance, galaBalance, BuyGalaSellTon)
	if !result2.Valid {
		t.Fatalf("Expected valid position, got invalid: %s", result2.Reason)
	}
	// Should use exactly 50% of TON
	expectedTON := tonBalance * PositionSizePercent
	if result2.TONAmount != expectedTON {
		t.Errorf("Expected %.2f TON (50%%), got %.2f", expectedTON, result2.TONAmount)
	}
}

func TestCalculatePositionSize_UnknownDirection(t *testing.T) {
	calc := NewCalculator()

	tonBalance := 100.0
	galaBalance := 1000.0

	// Test with invalid direction
	result := calc.CalculatePositionSize(tonBalance, galaBalance, Direction("InvalidDirection"))

	if result.Valid {
		t.Error("Expected invalid position for unknown direction, got valid")
	}
	if result.Reason == "" {
		t.Error("Expected reason for invalid position, got empty string")
	}
}

func TestCalculatePositionSize_ZeroBalances(t *testing.T) {
	calc := NewCalculator()

	result1 := calc.CalculatePositionSize(0, 0, BuyTonSellGala)
	if result1.Valid {
		t.Error("Expected invalid position for zero balances (BuyTonSellGala)")
	}

	result2 := calc.CalculatePositionSize(0, 0, BuyGalaSellTon)
	if result2.Valid {
		t.Error("Expected invalid position for zero balances (BuyGalaSellTon)")
	}
}

func TestCalculatePositionSize_NegativeBalances(t *testing.T) {
	calc := NewCalculator()

	result1 := calc.CalculatePositionSize(-10, 100, BuyTonSellGala)
	if result1.Valid {
		t.Error("Expected invalid position for negative TON balance")
	}

	result2 := calc.CalculatePositionSize(100, -10, BuyGalaSellTon)
	if result2.Valid {
		t.Error("Expected invalid position for negative GALA balance")
	}
}
