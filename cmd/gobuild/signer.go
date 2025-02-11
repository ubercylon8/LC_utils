package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Signer handles binary code signing
type Signer struct {
	certPath     string
	certPassword string
}

// NewSigner creates a new Signer instance
func NewSigner(certPath, certPassword string) *Signer {
	return &Signer{
		certPath:     certPath,
		certPassword: certPassword,
	}
}

// Sign performs code signing on the binary
func (s *Signer) Sign(binaryPath string) error {
	// Ensure the binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary file does not exist: %s", binaryPath)
	}

	// Get absolute paths
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	absCertPath, err := filepath.Abs(s.certPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for certificate: %v", err)
	}

	if runtime.GOOS == "windows" {
		return s.signWithSigntool(absPath)
	}
	return s.signWithOsslsigncode(absPath, absCertPath)
}

// signWithSigntool signs using Windows signtool.exe (native)
func (s *Signer) signWithSigntool(binaryPath string) error {
	args := []string{
		"sign",
		"/f", s.certPath,
		"/p", s.certPassword,
		"/tr", "http://timestamp.digicert.com",
		"/td", "sha256",
		"/fd", "sha256",
		binaryPath,
	}

	cmd := exec.Command("signtool.exe", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("signing with signtool failed: %v\nOutput: %s", err, output)
	}

	return nil
}

// signWithOsslsigncode signs using osslsigncode (for non-Windows platforms)
func (s *Signer) signWithOsslsigncode(binaryPath, certPath string) error {
	// Check if osslsigncode is installed
	if _, err := exec.LookPath("osslsigncode"); err != nil {
		return fmt.Errorf("osslsigncode not found. Please install it first (e.g., 'brew install osslsigncode' on macOS)")
	}

	// Create a temporary file for the signed output
	tmpDir := filepath.Dir(binaryPath)
	signedPath := filepath.Join(tmpDir, "signed_"+filepath.Base(binaryPath))

	// Sign the binary
	args := []string{
		"sign",
		"-pkcs12", certPath,
		"-pass", s.certPassword,
		"-n", "F0RT1KA CST Binary",
		"-i", "http://timestamp.digicert.com",
		"-h", "sha256",
		"-in", binaryPath,
		"-out", signedPath,
	}

	cmd := exec.Command("osslsigncode", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("signing with osslsigncode failed: %v\nOutput: %s", err, output)
	}

	// Replace original file with signed one
	if err := os.Rename(signedPath, binaryPath); err != nil {
		return fmt.Errorf("failed to replace original file with signed one: %v", err)
	}

	return nil
}

// signWithCodesign signs using macOS codesign command
func (s *Signer) signWithCodesign(binaryPath string) error {
	// Create a temporary directory for keychain operations
	tmpDir, err := os.MkdirTemp("", "codesign")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create paths for temporary files
	keychain := filepath.Join(tmpDir, "build.keychain")

	// Create and configure keychain
	identityName, err := s.setupKeychain(keychain)
	if err != nil {
		return fmt.Errorf("failed to setup keychain: %v", err)
	}

	if identityName == "" {
		return fmt.Errorf("no valid signing identity found in the certificate")
	}

	// Sign the binary using codesign
	args := []string{
		"-s", identityName,
		"-v",
		"--keychain", keychain,
		"--timestamp",
		"--force",
		binaryPath,
	}

	cmd := exec.Command("codesign", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("signing with codesign failed: %v\nOutput: %s", err, output)
	}

	return nil
}

// setupKeychain creates and configures a temporary keychain
func (s *Signer) setupKeychain(keychain string) (string, error) {
	// Create a new keychain
	createCmd := exec.Command("security", "create-keychain", "-p", s.certPassword, keychain)
	if output, err := createCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create keychain: %v\nOutput: %s", err, output)
	}

	// Set keychain settings
	settingsCmd := exec.Command("security", "set-keychain-settings", "-t", "3600", "-l", keychain)
	if output, err := settingsCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to set keychain settings: %v\nOutput: %s", err, output)
	}

	// Unlock the keychain
	unlockCmd := exec.Command("security", "unlock-keychain", "-p", s.certPassword, keychain)
	if output, err := unlockCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to unlock keychain: %v\nOutput: %s", err, output)
	}

	// Import the certificate
	importCmd := exec.Command("security", "import", s.certPath,
		"-k", keychain,
		"-P", s.certPassword,
		"-T", "/usr/bin/codesign",
		"-f", "pkcs12")
	if output, err := importCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to import certificate: %v\nOutput: %s", err, output)
	}

	// Allow codesign to access the keychain without prompting
	authCmd := exec.Command("security", "set-key-partition-list",
		"-S", "apple-tool:,apple:,codesign:",
		"-s", "-k", s.certPassword,
		keychain)
	if output, err := authCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to set key partition list: %v\nOutput: %s", err, output)
	}

	// Get the identity name from the keychain
	findCmd := exec.Command("security", "find-identity", "-p", "codesigning", "-v", keychain)
	output, err := findCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to find identity: %v\nOutput: %s", err, output)
	}

	// Parse the output to get the identity name
	// Output format: 1) <hash> "<identity name>"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "\"") {
			parts := strings.SplitN(line, "\"", 3)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("no valid signing identity found in keychain")
}
