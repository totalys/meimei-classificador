package main

import (
	"testing"
	"time"
)

func TestStripNonNumeric(t *testing.T) {
	// Test cases
	tests := []struct {
		input    string
		expected string
	}{
		{"(123) 456-7890", "1234567890"},
		{"+1 (987) 654-3210", "19876543210"},
		{"  555-123-4567  ", "5551234567"},
		{"+55(11)97644-4240", "5511976444240"},
		{"11976444240", "11976444240"},
		{"", ""},
	}

	for _, test := range tests {
		result := stripNonNumeric(test.input)
		if result != test.expected {
			t.Errorf("stripNonNumeric(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestGetAge(t *testing.T) {
	// Test cases

	// date to test the expectations:
	now, err := time.Parse("2006-01-02", "2023-08-01")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		age      int
		dobStr   string
		expected string
	}{
		{25, "", "25"},
		{0, "1990-08-01", "33"},
		{0, "2000-03-15", "23"},
		{0, "2000-12-31", "22"},
		{40, "1985-12-31", "40"},
	}

	for _, test := range tests {
		result := getAge(now, test.age, test.dobStr)
		if result != test.expected {
			t.Errorf("getAge(%q, %q) = %q; expected %q", test.age, test.dobStr, result, test.expected)
		}
	}
}

func TestGradeAverage(t *testing.T) {
	tests := []struct {
		pt, mat, log, red, expected string
	}{
		// Test scenarios with valid inputs
		{"4,0", "6,0", "1,0", "8,0", "5.43"},
		{"1,5", "6,0", "2,0", "4,5", "4.00"},
		{"10", "10", "5", "10", "10.00"}, // The sum is 35, so the result is 35*10/35 = 10.00
		{"0", "0", "0", "0", "0.00"},

		// Test scenarios with empty inputs (will be considered as 0)
		{"", "2.7", "3.9", "4.1", "3.06"},
		{"1.5", "", "3.9", "4.1", "2.71"},
		{"1.5", "2.7", "", "4.1", "1.60"},
		{"1.5", "2.7", "3.9", "", "1.54"},
	}

	for i, test := range tests {
		result := getGradeAverage(test.pt, test.mat, test.log, test.red)
		if result != test.expected {
			t.Errorf("Test case %d failed: expected '%s', got '%s'", i+1, test.expected, result)
		}
	}
}

func TestAverage(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		// Test scenarios with valid inputs
		{[]string{"1,5", "2,7", "3,9", "4,1"}, "3.05"},
		{[]string{"0,5", "1,2", "0,3", "0,0"}, "0.67"},
		{[]string{"10", "10", "5", "10"}, "8.75"}, // The sum is 35, so the result is 35/4 = 8.75

		// Test scenarios with inputs exceeding the range (will be considered as 0)
		{[]string{"12,5", "2,7", "3,9", "4,1"}, "5.80"},
		{[]string{"1,5", "11,2", "3,9", "4,1"}, "5.17"},
		{[]string{"1,5", "2,7", "8,3", "4,1"}, "4.15"},
		{[]string{"1,5", "2,7", "3,9", "15,1"}, "5.80"},

		// Test scenarios with all elements invalid (return 0)
		{[]string{"invalid", "abc", "xyz"}, "0.00"},
	}

	for i, test := range tests {
		result := average(test.input)
		if result != test.expected {
			t.Errorf("Test case %d failed: expected '%s', got '%s'", i+1, test.expected, result)
		}
	}
}
