// Package api provides task management functionality for LimaCharlie.
// This file implements task-related operations including:
// - Sending PUT commands to sensors
// - Running shell commands on sensors
// - Creating reliable tasks with retry mechanisms
// - Managing task metadata and responses
//
// The package supports both one-time tasks and reliable tasks through
// the ext-reliable-tasking extension.
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

// PutCommand sends a PUT command to a sensor to upload a file.
// This is a convenience wrapper around TaskSensor for file uploads.
//
// Parameters:
//   - creds: Authentication credentials for the API
//   - sensorID: ID of the target sensor
//   - path: Destination path on the sensor
//   - content: Content to write to the file
//   - investigationID: Optional investigation ID for tracking
//
// Returns:
//   - *TaskResponse: Response from the task execution
//   - error: Any error that occurred during the operation
func PutCommand(creds *auth.Credentials, sensorID string, path string, content string, investigationID string) (*TaskResponse, error) {
	task := fmt.Sprintf("put %s %s", path, content)
	return TaskSensor(creds, sensorID, []string{task}, investigationID)
}

// RunCommand sends a RUN command to execute a shell command on a sensor.
// This is a convenience wrapper around TaskSensor for command execution.
//
// Parameters:
//   - creds: Authentication credentials for the API
//   - sensorID: ID of the target sensor
//   - command: Shell command to execute
//   - investigationID: Optional investigation ID for tracking
//
// Returns:
//   - *TaskResponse: Response from the task execution
//   - error: Any error that occurred during the operation
func RunCommand(creds *auth.Credentials, sensorID string, command string, investigationID string) (*TaskResponse, error) {
	// Use --shell-command flag for running shell commands
	task := fmt.Sprintf(`run --shell-command '%s'`, command)
	return TaskSensor(creds, sensorID, []string{task}, investigationID)
}

// TaskSensor sends a task to a sensor. This is the core function for
// sending any type of task to a sensor.
//
// Parameters:
//   - creds: Authentication credentials for the API
//   - sensorID: ID of the target sensor
//   - tasks: List of tasks to execute
//   - investigationID: Optional investigation ID for tracking
//
// Returns:
//   - *TaskResponse: Response from the task execution
//   - error: Any error that occurred during the operation
func TaskSensor(creds *auth.Credentials, sensorID string, tasks []string, investigationID string) (*TaskResponse, error) {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/v1/%s", baseURL, sensorID))
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	// Prepare form data
	form := url.Values{}
	form.Add("tasks", strings.Join(tasks, ","))
	if investigationID != "" {
		form.Add("investigation_id", investigationID)
	}

	// Create request
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set API key in Authorization header
	authHeader, err := creds.GetAuthHeader()
	if err != nil {
		return nil, fmt.Errorf("error getting auth header: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	var response TaskResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("task error: %s", response.Error)
	}

	return &response, nil
}

// CreateExtensionRequest sends a request to a LimaCharlie extension.
// This is used for advanced functionality like reliable tasking.
//
// Parameters:
//   - creds: Authentication credentials for the API
//   - extensionName: Name of the extension to use
//   - action: Action to perform
//   - data: JSON-encoded data for the action
//
// Returns:
//   - error: Any error that occurred during the operation
func CreateExtensionRequest(creds *auth.Credentials, extensionName string, action string, data interface{}) error {
	// Build URL
	u, err := url.Parse(fmt.Sprintf("%s/v1/extension/request/%s", baseURL, extensionName))
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}

	// Add required query parameters
	q := u.Query()
	q.Set("oid", creds.OID)
	q.Set("action", action)
	u.RawQuery = q.Encode()

	// Convert data to map if it's a string
	var taskData map[string]interface{}
	switch v := data.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &taskData); err != nil {
			return fmt.Errorf("error parsing task data: %w", err)
		}
	case map[string]interface{}:
		taskData = v
	default:
		return fmt.Errorf("unsupported data type for task data")
	}

	// Convert task data to JSON string
	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("error encoding task data: %w", err)
	}

	// Prepare form data
	form := url.Values{}
	form.Add("data", string(jsonData))

	// Create request
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set API key in Authorization header
	authHeader, err := creds.GetAuthHeader()
	if err != nil {
		return fmt.Errorf("error getting auth header: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CreateReliableTask creates a task that will be retried until successful.
// Uses the ext-reliable-tasking extension to ensure task delivery.
//
// Parameters:
//   - creds: Authentication credentials for the API
//   - sensorID: ID of the target sensor
//   - command: Command to execute
//   - context: Optional context for tracking retries
//   - ttl: Time-to-live in seconds for the task
//
// Returns:
//   - error: Any error that occurred during the operation
func CreateReliableTask(creds *auth.Credentials, sensorID string, command string, context string, ttl int64) error {
	// Prepare the task data
	taskData := map[string]interface{}{
		"task": command,
		"ttl":  ttl,
		"sid":  sensorID,
	}

	// Add context if provided
	if context != "" {
		taskData["context"] = context
	}

	// Send the request to the reliable tasking extension
	return CreateExtensionRequest(creds, "ext-reliable-tasking", "task", taskData)
}
