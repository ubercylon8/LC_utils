// Package api provides types and functions for interacting with the LimaCharlie API.
// It includes functionality for managing sensors, tasks, payloads, and authentication.
//
// The package is organized into several main components:
// - Sensor management (listing, tagging, status)
// - Task execution (running commands, uploading files)
// - Payload management (uploading and managing payloads)
// - Authentication and authorization
//
// This file contains the core types and constants used throughout the package.
package api

import (
	"fmt"
)

const baseURL = "https://api.limacharlie.io"

// Platform constants
const (
	PlatformWindows = "windows"
	PlatformMacOS   = "mac"
	PlatformLinux   = "linux"
)

// Architecture constants
const (
	ArchX86   = "x86"
	ArchX64   = "x64"
	ArchARM   = "arm"
	ArchARM64 = "arm64"
)

// ListOptions contains parameters for filtering and paginating sensor lists
type ListOptions struct {
	// Limit the number of results
	Limit int
	// Include sensor tags in response
	WithTags bool
	// Filter by IP address
	WithIP string
	// Filter by hostname prefix
	WithHostnamePrefix string
	// Only return online sensors
	OnlyOnline bool
	// Continuation token for pagination
	ContinuationToken string
	// Filter by tag
	FilterTag string
}

// Sensor represents a LimaCharlie sensor with all its attributes.
// It contains information about the sensor's identity, platform,
// status, and configuration.
type Sensor struct {
	// SID is the unique sensor identifier
	SID string `json:"sid"`
	// InstallationID is the unique installation identifier
	InstallationID string `json:"did"`
	// PlatformID identifies the operating system platform
	PlatformID int64 `json:"plat"`
	// Architecture identifies the CPU architecture
	Architecture int `json:"arch"`
	// Hostname is the sensor's hostname
	Hostname string `json:"hostname"`
	// OID is the organization identifier
	OID string `json:"oid"`
	// Tags are the sensor's assigned tags
	Tags []string `json:"tags,omitempty"`
	// IsOnline indicates if the sensor is currently connected
	IsOnline bool `json:"is_online"`
	// LastSeen is the timestamp of last contact
	LastSeen string `json:"alive"`
	// EnrollmentTime is when the sensor was enrolled
	EnrollmentTime string `json:"enroll"`
	// ExternalIP is the sensor's external IP address
	ExternalIP string `json:"ext_ip"`
	// InternalIP is the sensor's internal IP address
	InternalIP string `json:"int_ip"`
	// Version is the sensor software version
	Version string `json:"version,omitempty"`
	// MacAddr is the sensor's MAC address
	MacAddr string `json:"mac_addr"`
	// IsIsolated indicates if the sensor is network isolated
	IsIsolated bool `json:"is_isolated"`
	// ShouldIsolate indicates if isolation is pending
	ShouldIsolate bool `json:"should_isolate"`
	// IsSealed indicates if the sensor is sealed
	IsSealed bool `json:"sealed"`
	// ShouldSeal indicates if sealing is pending
	ShouldSeal bool `json:"should_seal"`
	// KernelAvailable indicates if kernel mode is available
	KernelAvailable bool `json:"is_kernel_available"`
	// ExternalPlatform is the platform ID from external sources
	ExternalPlatform int64 `json:"ext_plat"`
	// InstallerVersion is the version of the installer used
	InstallerVersion string `json:"installer_version,omitempty"`
}

// SensorList represents a list of sensors response
type SensorList struct {
	Sensors []Sensor `json:"sensors"`
}

// OnlineStatusResponse represents the online status response
type OnlineStatusResponse struct {
	Online map[string]bool `json:"online"`
}

// TaskResponse represents the response from a task operation
type TaskResponse struct {
	Error string `json:"error,omitempty"`
	ID    string `json:"id,omitempty"`
}

// GetPlatformString returns a human-readable platform name based on
// the sensor's PlatformID. It handles all known platform types and
// returns "Unknown" with the hex value for unknown platforms.
func (s *Sensor) GetPlatformString() string {
	// Platform IDs from LimaCharlie
	// Windows: 0x10000000 (268435456)
	// macOS:   0x30000000 (805306368)
	// Linux:   0x20000000 (536870912)
	// Cloud:   0x90000000 (2415919104)
	switch s.PlatformID {
	case 268435456:
		return "Windows"
	case 805306368:
		return "macOS"
	case 536870912:
		return "Linux"
	case 2415919104:
		return "Cloud"
	default:
		return fmt.Sprintf("Unknown (0x%x)", s.PlatformID)
	}
}

// GetArchitectureString returns a human-readable architecture name
// based on the sensor's Architecture value. It handles all known
// architectures and returns "Unknown" with the value for unknown types.
func (s *Sensor) GetArchitectureString() string {
	// Architecture IDs from LimaCharlie
	// x86: 1
	// x64: 2
	// ARM: 3
	// ARM64: 4
	// Cloud: 9
	switch s.Architecture {
	case 1:
		return "x86"
	case 2:
		return "x64"
	case 3:
		return "ARM"
	case 4:
		return "ARM64"
	case 9:
		return "Cloud"
	default:
		return fmt.Sprintf("Unknown (%d)", s.Architecture)
	}
}

// GetLastSeenString returns the last seen time of the sensor.
// Returns "Never" if the sensor has not been seen.
func (s *Sensor) GetLastSeenString() string {
	if s.LastSeen == "" {
		return "Never"
	}
	return s.LastSeen
}

// GetEnrollmentTimeString returns the enrollment time of the sensor.
// Returns "Never" if the sensor has not been enrolled.
func (s *Sensor) GetEnrollmentTimeString() string {
	if s.EnrollmentTime == "" {
		return "Never"
	}
	return s.EnrollmentTime
}
