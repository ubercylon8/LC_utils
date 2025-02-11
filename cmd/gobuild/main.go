package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	sourcePath   string
	outputPath   string
	targetOS     string
	targetArch   string
	certPath     string
	certPassword string
	verbose      bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gobuild",
		Short: "A utility to build and sign Go binaries",
		Long: `gobuild is a command line utility that compiles Go source code into platform-specific binaries
and optionally code-signs Windows executables with provided certificates.

Example:
  gobuild --source ./myapp --output ./bin --target-os windows --target-arch amd64 --cert cert.pfx --cert-pass mypass`,
	}

	// Build command
	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build and optionally sign a Go binary",
		Run:   runBuild,
	}

	// Sign command
	var signCmd = &cobra.Command{
		Use:   "sign",
		Short: "Sign an existing binary",
		Long: `Sign an existing Windows binary using a code signing certificate.
		
Example:
  gobuild sign --binary ./myapp.exe --cert cert.pfx --cert-pass mypass`,
		Run: runSign,
	}

	// Build command flags
	buildCmd.PersistentFlags().StringVarP(&sourcePath, "source", "s", "", "Path to Go source code (required)")
	buildCmd.PersistentFlags().StringVarP(&outputPath, "output", "o", "", "Output path for the binary (required)")
	buildCmd.PersistentFlags().StringVar(&targetOS, "target-os", runtime.GOOS, "Target OS (windows, linux, darwin)")
	buildCmd.PersistentFlags().StringVar(&targetArch, "target-arch", runtime.GOARCH, "Target architecture (amd64, 386)")
	buildCmd.PersistentFlags().StringVar(&certPath, "cert", "", "Path to code signing certificate (PFX format)")
	buildCmd.PersistentFlags().StringVar(&certPassword, "cert-pass", "", "Certificate password")
	buildCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Sign command flags
	signCmd.PersistentFlags().StringVarP(&outputPath, "binary", "b", "", "Path to the binary to sign (required)")
	signCmd.PersistentFlags().StringVar(&certPath, "cert", "", "Path to code signing certificate (PFX format) (required)")
	signCmd.PersistentFlags().StringVar(&certPassword, "cert-pass", "", "Certificate password (required)")
	signCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Mark required flags
	buildCmd.MarkPersistentFlagRequired("source")
	buildCmd.MarkPersistentFlagRequired("output")
	signCmd.MarkPersistentFlagRequired("binary")
	signCmd.MarkPersistentFlagRequired("cert")
	signCmd.MarkPersistentFlagRequired("cert-pass")

	// Add commands to root
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(signCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runBuild(cmd *cobra.Command, args []string) {
	// Validate input
	if err := validateBuildInput(); err != nil {
		log.Fatalf("Input validation failed: %v", err)
	}

	// Create builder instance
	builder := NewBuilder(sourcePath, outputPath, targetOS, targetArch, verbose)

	// Build binary
	fmt.Printf("Building binary for %s/%s...\n", targetOS, targetArch)
	bar := progressbar.Default(100)

	if err := builder.Build(); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
	bar.Finish()

	// Sign Windows binary if certificate provided
	if targetOS == "windows" && certPath != "" {
		fmt.Println("Signing Windows binary...")
		signer := NewSigner(certPath, certPassword)
		if err := signer.Sign(outputPath); err != nil {
			log.Fatalf("Signing failed: %v", err)
		}
	}

	fmt.Printf("\nBuild completed successfully!\nBinary location: %s\n", outputPath)
}

func runSign(cmd *cobra.Command, args []string) {
	// Validate input
	if err := validateSignInput(); err != nil {
		log.Fatalf("Input validation failed: %v", err)
	}

	// Sign the binary
	fmt.Println("Signing binary...")
	signer := NewSigner(certPath, certPassword)
	if err := signer.Sign(outputPath); err != nil {
		log.Fatalf("Signing failed: %v", err)
	}

	fmt.Printf("\nSigning completed successfully!\nSigned binary location: %s\n", outputPath)
}

func validateBuildInput() error {
	// Check if source path exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Validate target OS
	validOS := map[string]bool{"windows": true, "linux": true, "darwin": true}
	if !validOS[targetOS] {
		return fmt.Errorf("invalid target OS: %s", targetOS)
	}

	// Validate target architecture
	validArch := map[string]bool{"amd64": true, "386": true, "arm64": true}
	if !validArch[targetArch] {
		return fmt.Errorf("invalid target architecture: %s", targetArch)
	}

	// Validate certificate if provided for Windows
	if targetOS == "windows" && certPath != "" {
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			return fmt.Errorf("certificate file does not exist: %s", certPath)
		}
		if certPassword == "" {
			return fmt.Errorf("certificate password is required when using a certificate")
		}
	}

	return nil
}

func validateSignInput() error {
	// Check if binary exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("binary file does not exist: %s", outputPath)
	}

	// Check if certificate exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("certificate file does not exist: %s", certPath)
	}

	// Validate certificate password
	if certPassword == "" {
		return fmt.Errorf("certificate password is required")
	}

	return nil
}
