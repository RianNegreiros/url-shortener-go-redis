package helpers

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestEnforceHTTP(t *testing.T) {
    assert := assert.New(t)

    // Test with a URL that does not start with "http"
    url := "example.com"
    expected := "http://example.com"
    result := EnforceHTTP(url)
    assert.Equal(expected, result, "They should be equal")

    // Test with a URL that starts with "http"
    url = "http://example.com"
    expected = "http://example.com"
    result = EnforceHTTP(url)
    assert.Equal(expected, result, "They should be equal")
}