// Package api provides authentication functionality for LimaCharlie.
// This file implements API-specific authentication operations including:
// - Validating API credentials
// - Managing authentication headers
// - Ensuring proper authentication for API calls
//
// The package uses the auth package for core authentication functionality
// and provides additional API-specific authentication features.
package api

import (
	"fmt"

	"LC_utils/internal/auth"
)

// ensureAuthenticated ensures we have a valid API key and returns an error
// if authentication is not possible. This is used internally by API functions
// to validate credentials before making requests.
//
// Parameters:
//   - creds: The credentials to validate
//
// Returns:
//   - error: An error if authentication fails, nil otherwise
func ensureAuthenticated(creds *auth.Credentials) error {
	if creds.GetAPIKey() == "" {
		return fmt.Errorf("API key is empty")
	}
	return nil
}
