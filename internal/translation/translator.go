package translation

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Translator defines the interface for translation services
type Translator interface {
	Translate(text, targetLang string) (string, error)
}

// DBInterface defines the minimal database interface needed for proxy settings
type DBInterface interface {
	GetSetting(key string) (string, error)
}

// buildProxyURL constructs a proxy URL from settings
func buildProxyURL(proxyType, proxyHost, proxyPort, username, password string) string {
	if proxyHost == "" || proxyPort == "" {
		return ""
	}

	// Build auth string if username is provided
	auth := ""
	if username != "" {
		if password != "" {
			auth = username + ":" + password + "@"
		} else {
			auth = username + "@"
		}
	}

	return fmt.Sprintf("%s://%s%s:%s", proxyType, auth, proxyHost, proxyPort)
}

// createHTTPClient creates an HTTP client with optional proxy support
func createHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Configure proxy if provided
	if proxyURL != "" {
		parsedProxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedProxy)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return client, nil
}

// CreateHTTPClientWithProxy creates an HTTP client with global proxy settings if enabled
func CreateHTTPClientWithProxy(db DBInterface, timeout time.Duration) (*http.Client, error) {
	var proxyURL string

	// Check if global proxy is enabled
	proxyEnabled, _ := db.GetSetting("proxy_enabled")
	if proxyEnabled == "true" {
		// Build proxy URL from global settings
		proxyType, _ := db.GetSetting("proxy_type")
		proxyHost, _ := db.GetSetting("proxy_host")
		proxyPort, _ := db.GetSetting("proxy_port")
		proxyUsername, _ := db.GetSetting("proxy_username")
		proxyPassword, _ := db.GetSetting("proxy_password")
		proxyURL = buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
	}

	// Create HTTP client with or without proxy
	return createHTTPClient(proxyURL, timeout)
}

// MockTranslator is a simple translator for demonstration
type MockTranslator struct{}

func NewMockTranslator() *MockTranslator {
	return &MockTranslator{}
}

func (t *MockTranslator) Translate(text, targetLang string) (string, error) {
	// In a real application, this would call an external API (Google, DeepL, etc.)
	// For now, we simulate translation by appending the language code.
	// We can also do some simple word replacements to make it look "translated"

	prefix := fmt.Sprintf("[%s] ", strings.ToUpper(targetLang))
	if strings.HasPrefix(text, prefix) {
		return text, nil
	}

	return prefix + text, nil
}
