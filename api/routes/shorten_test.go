package routes

import (
	"errors"
	"testing"
	"time"

	"github.com/RianNegreiros/url-shortener-go-redis/api/database"
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

func TestStoreURL(t *testing.T) {
	testCases := []struct {
		id     string
		url    string
		expiry int
		err    error
	}{
		{"", "http://example.com", 0, nil},
		{"", "http://example.com", 1, nil},
		{"", "http://example.com", 24, nil},
		{"", "http://example.com", 25, nil},
		{"", "http://example.com", 26, nil},
		{"", "http://example.com", 100, nil},
	}

	redisClient := database.CreateClient(0)

	for _, tc := range testCases {
		result := storeURL(redisClient, tc.id, tc.url, time.Duration(tc.expiry)*time.Hour)
		if result != tc.err {
			t.Errorf("storeURL(%s, %s, %d) = %v, want %v", tc.id, tc.url, tc.expiry, result, tc.err)
		}
	}
}
