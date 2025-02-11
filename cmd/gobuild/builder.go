package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Builder handles the compilation of Go source code
type Builder struct {
	sourcePath string
	outputPath string
	targetOS   string
	targetArch string
	verbose    bool
}

// NewBuilder creates a new Builder instance
func NewBuilder(sourcePath, outputPath, targetOS, targetArch string, verbose bool) *Builder {
	return &Builder{
		sourcePath: sourcePath,
		outputPath: outputPath,
		targetOS:   targetOS,
		targetArch: targetArch,
		verbose:    verbose,
	}
}

// Build compiles the Go source code into a binary
func (b *Builder) Build() error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(b.outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Get absolute paths
	absSource, err := filepath.Abs(b.sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute source path: %v", err)
	}

	absOutput, err := filepath.Abs(b.outputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute output path: %v", err)
	}

	// Set environment variables for cross-compilation
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", b.targetOS))
	env = append(env, fmt.Sprintf("GOARCH=%s", b.targetArch))

	// Prepare build command
	args := []string{"build"}
	if b.verbose {
		args = append(args, "-v")
	}
	args = append(args, "-o", absOutput)
	args = append(args, absSource)

	// Set working directory to source path
	cmd := exec.Command("go", args...)
	cmd.Dir = filepath.Dir(absSource)
	cmd.Env = env

	// If verbose, show build command
	if b.verbose {
		fmt.Printf("Running build command: %v\n", cmd.Args)
		fmt.Printf("Working directory: %s\n", cmd.Dir)
		fmt.Printf("Environment: GOOS=%s GOARCH=%s\n", b.targetOS, b.targetArch)
	}

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v\nOutput: %s", err, output)
	}

	if b.verbose {
		fmt.Printf("Build output:\n%s\n", output)
	}

	// Verify the binary was created
	if _, err := os.Stat(absOutput); os.IsNotExist(err) {
		return fmt.Errorf("build completed but binary was not created at %s", absOutput)
	}

	return nil
}
