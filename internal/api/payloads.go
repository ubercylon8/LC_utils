// Package api provides payload management functionality for LimaCharlie.
// This file implements payload-related operations including:
// - Uploading payloads to LimaCharlie
// - Finding executable files in directories
// - Managing payload metadata
//
// The package uses secure upload mechanisms and handles authentication
// through the auth package.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"LC_utils/internal/auth"
)

const (
	payloadEndpoint = "https://api.limacharlie.io/v1/payload"
)

// PayloadUploadResponse represents the response from the payload upload request.
// It contains the URL where the payload should be uploaded.
type PayloadUploadResponse struct {
	// PutURL is the pre-signed URL for uploading the payload
	PutURL string `json:"put_url"`
}

// UploadPayload uploads a payload file to LimaCharlie.
// The function handles the two-step upload process:
// 1. Get a pre-signed upload URL from LimaCharlie
// 2. Upload the file contents to the provided URL
//
// Parameters:
//   - orgID: Organization ID
//   - apiKey: API Key for authentication
//   - filePath: Path to the file to upload
//
// Returns:
//   - error: Any error that occurred during the operation
func UploadPayload(orgID, apiKey, filePath string) error {
	// Create credentials
	creds := auth.NewCredentials(orgID, apiKey)
	if err := creds.ValidateCredentials(); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	fileName := filepath.Base(filePath)
	url := fmt.Sprintf("%s/%s/%s", payloadEndpoint, orgID, fileName)

	// Step 1: Get upload URL
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Get JWT token and set authorization header
	req.Header.Set("Authorization", creds.GetAuthHeader())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error getting upload URL: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var uploadResp PayloadUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	// Step 2: Upload file to the provided URL
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	uploadReq, err := http.NewRequest("PUT", uploadResp.PutURL, bytes.NewReader(fileContent))
	if err != nil {
		return fmt.Errorf("error creating upload request: %w", err)
	}

	uploadReq.Header.Set("Content-Type", "application/octet-stream")

	resp2, err := client.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		return fmt.Errorf("error uploading file: status=%d, body=%s", resp2.StatusCode, string(body))
	}

	return nil
}

// FindExecutableFiles finds all executable files in the given directory
// and its subdirectories. Currently only looks for .exe files.
//
// Parameters:
//   - basePath: Root directory to start searching from
//
// Returns:
//   - []string: List of paths to executable files found
//   - error: Any error that occurred during the operation
func FindExecutableFiles(basePath string) ([]string, error) {
	var files []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".exe") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
