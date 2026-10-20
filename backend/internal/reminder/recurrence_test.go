package reminder

import (
	"testing"
	"time"
)

func TestParseRecurrenceDays(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"senin,rabu,jumat", []int{1, 3, 5}},
		{"1, 3, 5", []int{1, 3, 5}},
		{"tuesday, friday", []int{2, 5}},
		{"minggu", []int{7}},
		{"invalid, 2, 7", []int{2, 7}},
	}

	for _, tt := range tests {
		got := ParseRecurrenceDays(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("ParseRecurrenceDays(%q) len = %d; want %d", tt.input, len(got), len(tt.expected))
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("ParseRecurrenceDays(%q)[%d] = %d; want %d", tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}

func TestCalculateNextRunDaily(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	baseTime := time.Date(2026, 10, 20, 8, 30, 0, 0, loc)

	next := CalculateNextRun(baseTime, RecurrenceDaily, "", loc)
	expected := time.Date(2026, 10, 21, 8, 30, 0, 0, loc)

	if !next.Equal(expected) {
		t.Errorf("CalculateNextRun Daily = %v; want %v", next, expected)
	}
}

func TestCalculateNextRunOnce(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	baseTime := time.Date(2026, 10, 20, 8, 30, 0, 0, loc)

	next := CalculateNextRun(baseTime, RecurrenceOnce, "", loc)
	if !next.IsZero() {
		t.Errorf("CalculateNextRun Once should be zero; got %v", next)
	}
}

func TestCalculateNextRunSpecificDays(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	// 2026-10-20 is Tuesday (ISO day 2)
	baseTime := time.Date(2026, 10, 20, 8, 30, 0, 0, loc)

	// Next should be Wednesday (ISO day 3)
	next := CalculateNextRun(baseTime, RecurrenceSpecificDays, "1,3,5", loc)
	expected := time.Date(2026, 10, 21, 8, 30, 0, 0, loc) // Wednesday
	if !next.Equal(expected) {
		t.Errorf("CalculateNextRun SpecificDays (Tue->Wed) = %v; want %v", next, expected)
	}

	// From Friday (2026-10-23, ISO day 5), next should be Monday (2026-10-26, ISO day 1)
	friday := time.Date(2026, 10, 23, 8, 30, 0, 0, loc)
	nextFromFriday := CalculateNextRun(friday, RecurrenceSpecificDays, "1,3,5", loc)
	expectedMonday := time.Date(2026, 10, 26, 8, 30, 0, 0, loc)
	if !nextFromFriday.Equal(expectedMonday) {
		t.Errorf("CalculateNextRun SpecificDays (Fri->Mon) = %v; want %v", nextFromFriday, expectedMonday)
	}
}
