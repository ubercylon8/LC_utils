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
		if opts.WithTags {
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
	req.Header.Set("Authorization", creds.GetAuthHeader())

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

	// Filter online sensors if requested
	if opts != nil && opts.OnlyOnline {
		var onlineSensors []Sensor
		for _, sensor := range sensorList.Sensors {
			if sensor.IsOnline {
				onlineSensors = append(onlineSensors, sensor)
			}
		}
		return onlineSensors, nil
	}

	return sensorList.Sensors, nil
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
	u, err := url.Parse(fmt.Sprintf("%s/v1/online/%s", baseURL, creds.OID))
	if err != nil {
		return &OnlineStatusResponse{Online: make(map[string]bool)}, fmt.Errorf("error parsing URL: %w", err)
	}

	// Add query parameters for sensor IDs
	q := u.Query()
	for _, sid := range sensorIDs {
		q.Add("sids", sid)
	}
	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return &OnlineStatusResponse{Online: make(map[string]bool)}, fmt.Errorf("error creating request: %w", err)
	}

	// Set API key in Authorization header
	req.Header.Set("Authorization", creds.GetAuthHeader())

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &OnlineStatusResponse{Online: make(map[string]bool)}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &OnlineStatusResponse{Online: make(map[string]bool)}, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &OnlineStatusResponse{Online: make(map[string]bool)}, fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var response OnlineStatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return &OnlineStatusResponse{Online: make(map[string]bool)}, fmt.Errorf("error decoding response: %w", err)
	}

	return &response, nil
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
	// Handle adding tags
	if len(tags.AddTags) > 0 {
		// Build URL for adding tags
		u, err := url.Parse(fmt.Sprintf("%s/v1/%s/tags", baseURL, sensorID))
		if err != nil {
			return fmt.Errorf("error parsing URL: %w", err)
		}

		// Add tags as a single comma-separated parameter
		q := u.Query()
		q.Set("tags", strings.Join(tags.AddTags, ","))
		q.Set("oid", creds.OID)
		u.RawQuery = q.Encode()

		// Create request
		req, err := http.NewRequest("POST", u.String(), nil)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		// Set API key in Authorization header
		req.Header.Set("Authorization", creds.GetAuthHeader())

		// Make request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
		}
	}

	// Handle removing tags
	if len(tags.RemoveTags) > 0 {
		// Build URL for removing tags
		u, err := url.Parse(fmt.Sprintf("%s/v1/%s/tags", baseURL, sensorID))
		if err != nil {
			return fmt.Errorf("error parsing URL: %w", err)
		}

		// Add tags as a single comma-separated parameter
		q := u.Query()
		q.Set("tags", strings.Join(tags.RemoveTags, ","))
		q.Set("oid", creds.OID)
		u.RawQuery = q.Encode()

		// Create request
		req, err := http.NewRequest("DELETE", u.String(), nil)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		// Set API key in Authorization header
		req.Header.Set("Authorization", creds.GetAuthHeader())

		// Make request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
		}
	}

	return nil
}
