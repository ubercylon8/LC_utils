// Package api provides sensor management functionality for LimaCharlie.
// This file implements sensor-related operations including:
// - Listing and filtering sensors
// - Checking online status
// - Managing sensor tags
// - Retrieving sensor details
//
// Each function is designed to handle errors gracefully and provide
// meaningful error messages for troubleshooting.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"LC_utils/internal/auth"
)

// ListSensors retrieves all sensors from LimaCharlie platform.
// It supports filtering by various criteria through the ListOptions parameter.
// The function handles pagination and online status filtering internally.
//
// Parameters:
//   - creds: Authentication credentials for the API
//   - opts: Optional filtering and pagination parameters
//
// Returns:
//   - []Sensor: List of sensors matching the criteria
//   - error: Any error that occurred during the operation
func ListSensors(creds *auth.Credentials, opts *ListOptions) ([]Sensor, error) {
	// Build URL with query parameters
	u, err := url.Parse(fmt.Sprintf("%s/v1/sensors/%s", baseURL, creds.OID))
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	// Add query parameters
	q := u.Query()
	if opts != nil {
		if opts.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		// Always fetch tags if we need them for filtering
		if opts.WithTags || opts.FilterTag != "" {
			q.Set("with_tags", "true")
		}
		if opts.WithIP != "" {
			q.Set("with_ip", opts.WithIP)
		}
		if opts.WithHostnamePrefix != "" {
			q.Set("with_hostname_prefix", opts.WithHostnamePrefix)
		}
		if opts.ContinuationToken != "" {
			q.Set("continuation_token", opts.ContinuationToken)
		}
	}
	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set API key in Authorization header
	authHeader, err := creds.GetAuthHeader()
	if err != nil {
		return nil, fmt.Errorf("error getting auth header: %w", err)
	}
	req.Header.Set("Authorization", authHeader)

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var sensorList SensorList
	if err := json.Unmarshal(body, &sensorList); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Filter sensors based on criteria
	var filteredSensors []Sensor
	for _, sensor := range sensorList.Sensors {
		include := true

		// Filter by online status if requested
		if opts != nil && opts.OnlyOnline && !sensor.IsOnline {
			include = false
			continue
		}

		// Filter by tag if specified
		if opts != nil && opts.FilterTag != "" {
			tagFound := false
			for _, tag := range sensor.Tags {
				if tag == opts.FilterTag {
					tagFound = true
					break
				}
			}
			if !tagFound {
				include = false
				continue
			}
		}

		if include {
			filteredSensors = append(filteredSensors, sensor)
		}
	}

	return filteredSensors, nil
}

// GetOnlineStatus retrieves the online status of multiple sensors.
// This is more efficient than checking individual sensors when you need
// to check multiple sensors at once.
//
// Parameters:
//   - creds: Authentication credentials for the API
//   - sensorIDs: List of sensor IDs to check
//
// Returns:
//   - *OnlineStatusResponse: Map of sensor IDs to their online status
//   - error: Any error that occurred during the operation
func GetOnlineStatus(creds *auth.Credentials, sensorIDs []string) (*OnlineStatusResponse, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/v1/sensors/%s/online", baseURL, creds.OID))
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(strings.Join(sensorIDs, ",")))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set API key in Authorization header
	authHeader, err := creds.GetAuthHeader()
	if err != nil {
		return nil, fmt.Errorf("error getting auth header: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "text/plain")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var response OnlineStatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &response, nil
}

// TagSensorRequest represents a request to modify sensor tags
type TagSensorRequest struct {
	AddTags    []string `json:"add_tags,omitempty"`
	RemoveTags []string `json:"remove_tags,omitempty"`
}

// TagSensor adds or removes tags from a specific sensor.
// The function can handle both adding and removing tags in a single call.
// If both operations are requested, adds are performed before removes.
//
// Parameters:
//   - creds: Authentication credentials for the API
//   - sensorID: ID of the sensor to modify
//   - tags: TagSensorRequest containing tags to add and/or remove
//
// Returns:
//   - error: Any error that occurred during the operation
func TagSensor(creds *auth.Credentials, sensorID string, tags TagSensorRequest) error {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/v1/%s/tags", baseURL, sensorID))
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}

	// Add tags as query parameters
	q := u.Query()
	if len(tags.AddTags) > 0 {
		for _, tag := range tags.AddTags {
			q.Add("tags", tag)
		}
	}
	if len(tags.RemoveTags) > 0 {
		for _, tag := range tags.RemoveTags {
			q.Add("remove_tags", tag)
		}
	}
	u.RawQuery = q.Encode()

	fmt.Printf("[DEBUG] TagSensor - URL: %s\n", u.String())

	// Create request
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set API key in Authorization header
	authHeader, err := creds.GetAuthHeader()
	if err != nil {
		return fmt.Errorf("error getting auth header: %w", err)
	}
	fmt.Printf("[DEBUG] TagSensor - Auth Header: %s\n", authHeader[:20]+"...") // Only show first 20 chars for security
	req.Header.Set("Authorization", authHeader)

	// Make request
	client := &http.Client{}
	fmt.Printf("[DEBUG] TagSensor - Sending request for sensor %s...\n", sensorID)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("[DEBUG] TagSensor - Response Status: %d\n", resp.StatusCode)
	fmt.Printf("[DEBUG] TagSensor - Response Body: %s\n", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
