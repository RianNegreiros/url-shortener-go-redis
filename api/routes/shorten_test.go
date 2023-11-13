package routes

import (
	"errors"
	"testing"
)

func TestValidateURL(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
		err      error
	}{
		{"http://example.com", "http://example.com", nil},
		{"https://example.com", "https://example.com", nil},
		{"example.com", "", errors.New("invalid url")},
	}

	for _, tc := range testCases {
		result, err := validateURL(tc.url)
		if result != tc.expected || (err != nil && err.Error() != tc.err.Error()) {
			t.Errorf("validateURL(%s) = %s, %v, want %s, %v", tc.url, result, err, tc.expected, tc.err)
		}
	}
}

func TestGenerateID(t *testing.T) {
	testCases := []struct {
		customShort string
		expectedLen int
	}{
		{"", 6},
		{"custom", 6},
	}

	for _, tc := range testCases {
		result := generateID(tc.customShort)
		if len(result) != tc.expectedLen {
			t.Errorf("generateID(%s) = %d, want %d", tc.customShort, len(result), tc.expectedLen)
		}
	}
}
