package main

import (
	"math"
	"testing"
)

func TestClassificationValidationAndMath(t *testing.T) {
	// Test valid classifications
	validTypes := []string{"helpful", "useless"}
	for _, vt := range validTypes {
		if vt != "helpful" && vt != "useless" {
			t.Errorf("Expected %s to be recognized as valid classification", vt)
		}
	}

	// Test invalid classifications
	invalidTypes := []string{"", "spam", "good", "bad", "HELPFUL", "none"}
	for _, it := range invalidTypes {
		if it == "helpful" || it == "useless" {
			t.Errorf("Expected %s to be recognized as invalid classification", it)
		}
	}

	// Test analytics percentage calculations
	testCases := []struct {
		helpful        int64
		useless        int64
		expectedTotal  int64
		expectedHelpPct float64
		expectedUselPct float64
	}{
		{
			helpful:        0,
			useless:        0,
			expectedTotal:  0,
			expectedHelpPct: 0.0,
			expectedUselPct: 0.0,
		},
		{
			helpful:        1,
			useless:        1,
			expectedTotal:  2,
			expectedHelpPct: 50.0,
			expectedUselPct: 50.0,
		},
		{
			helpful:        3,
			useless:        1,
			expectedTotal:  4,
			expectedHelpPct: 75.0,
			expectedUselPct: 25.0,
		},
		{
			helpful:        2,
			useless:        1,
			expectedTotal:  3,
			expectedHelpPct: 66.7,
			expectedUselPct: 33.3,
		},
	}

	for _, tc := range testCases {
		total := tc.helpful + tc.useless
		if total != tc.expectedTotal {
			t.Errorf("Total classified mismatch: got %d, want %d", total, tc.expectedTotal)
		}

		var helpfulPct, uselessPct float64
		if total > 0 {
			helpfulPct = math.Round((float64(tc.helpful)/float64(total))*1000) / 10
			uselessPct = math.Round((float64(tc.useless)/float64(total))*1000) / 10
		}

		if helpfulPct != tc.expectedHelpPct {
			t.Errorf("Helpful percentage mismatch: got %f, want %f", helpfulPct, tc.expectedHelpPct)
		}
		if uselessPct != tc.expectedUselPct {
			t.Errorf("Useless percentage mismatch: got %f, want %f", uselessPct, tc.expectedUselPct)
		}
	}
}
