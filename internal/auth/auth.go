// Package auth provides authentication functionality for LimaCharlie API access.
// It handles credential management, validation, and JWT token generation.
//
// The package is designed to be secure and follows best practices for
// handling sensitive credentials. It provides:
// - Secure credential storage and management
// - JWT token generation and validation
// - API key validation
// - Authorization header generation
//
// Example usage:
//
//	creds := auth.NewCredentials(orgID, apiKey)
//	if err := creds.ValidateCredentials(); err != nil {
//	    log.Fatal("Invalid credentials:", err)
//	}
//	authHeader := creds.GetAuthHeader()
package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// JWTResponse represents the response from the JWT endpoint
type JWTResponse struct {
	JWT string `json:"jwt"`
}

// Credentials represents authentication credentials for LimaCharlie.
// It contains the organization ID and API key required for API access.
type Credentials struct {
	// OID is the organization identifier
	OID string
	// apiKey is the API key for authentication (kept private)
	apiKey string
	jwt    string // cached JWT token
}

// NewCredentials creates a new Credentials instance with the provided
// organization ID and API key. It performs basic validation of the
// input parameters.
//
// Parameters:
//   - orgID: The organization identifier
//   - apiKey: The API key for authentication
//
// Returns:
//   - *Credentials: A new credentials instance
func NewCredentials(orgID, apiKey string) *Credentials {
	return &Credentials{
		OID:    orgID,
		apiKey: apiKey,
	}
}

// GetJWT obtains a JWT token from LimaCharlie
func (c *Credentials) GetJWT() (string, error) {
	// Return cached JWT if available
	if c.jwt != "" {
		return c.jwt, nil
	}

	// Build URL and form data
	form := url.Values{}
	form.Add("oid", c.OID)
	form.Add("secret", c.apiKey)

	// Create request
	req, err := http.NewRequest("POST", "https://jwt.limacharlie.io", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var jwtResp JWTResponse
	if err := json.Unmarshal(body, &jwtResp); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	// Cache the JWT token
	c.jwt = jwtResp.JWT
	return c.jwt, nil
}

// GetAuthHeader generates the Authorization header value for API requests.
// The header format follows LimaCharlie's requirements for API authentication.
//
// Returns:
//   - string: The complete Authorization header value
func (c *Credentials) GetAuthHeader() string {
	return fmt.Sprintf("Bearer %s", c.apiKey)
}

// GetAPIKey returns the API key associated with these credentials.
// This method provides controlled access to the private apiKey field.
//
// Returns:
//   - string: The API key
func (c *Credentials) GetAPIKey() string {
	return c.apiKey
}

// ValidateCredentials checks if the credentials are valid by ensuring
// both the organization ID and API key are present and properly formatted.
//
// Returns:
//   - error: An error if validation fails, nil otherwise
func (c *Credentials) ValidateCredentials() error {
	if c.OID == "" {
		return fmt.Errorf("organization ID is required")
	}
	if c.apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	if !strings.HasPrefix(c.apiKey, "lc_") {
		return fmt.Errorf("invalid API key format (should start with 'lc_')")
	}
	return nil
}

// String provides a safe string representation of the credentials,
// masking the API key to prevent accidental exposure in logs.
//
// Returns:
//   - string: A string representation with masked API key
func (c *Credentials) String() string {
	return fmt.Sprintf("Credentials{OID: %s, APIKey: ****}", c.OID)
}
