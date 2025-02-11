# gobuild

A secure and efficient command-line utility for building and code-signing Go binaries.

## Features

- Cross-platform Go binary compilation
- Windows code signing support (requires signtool.exe or Wine)
- Progress indication for build process
- Verbose mode for debugging
- Input validation and secure handling of certificates

## Prerequisites

- Go 1.21 or later
- For Windows code signing:
  - On Windows: Windows SDK (includes signtool.exe)
  - On Linux/macOS: Wine (optional, for Windows signing)

## Installation

Since this is a local module, you can build and install it directly:

```bash
# From the test_utils directory
go build -o bin/gobuild cmd/gobuild/*.go

# Optional: Install to your GOPATH
go install ./cmd/gobuild
```

## Usage

Basic usage:
```bash
gobuild --source ./path/to/source --output ./path/to/output
```

Cross-compilation example:
```bash
gobuild --source ./myapp --output ./bin/myapp.exe --target-os windows --target-arch amd64
```

Windows code signing:
```bash
gobuild --source ./myapp \
        --output ./bin/myapp.exe \
        --target-os windows \
        --target-arch amd64 \
        --cert ./cert.pfx \
        --cert-pass "your-password"
```

### Available Flags

- `--source, -s`: Path to Go source code (required)
- `--output, -o`: Output path for the binary (required)
- `--target-os`: Target operating system (windows, linux, darwin)
- `--target-arch`: Target architecture (amd64, 386, arm64)
- `--cert`: Path to code signing certificate (PFX format)
- `--cert-pass`: Certificate password
- `--verbose, -v`: Enable verbose output

## Security Considerations

1. Certificate handling:
   - Certificate passwords are never logged or stored
   - PFX files are accessed only during signing
   - Secure error messages that don't leak sensitive information

2. Build process:
   - Input validation for all paths and arguments
   - Secure handling of environment variables
   - Clean error handling without exposing system details

## Error Codes

- 1: General error (invalid arguments, build failure)
- 2: Code signing error
- 3: Permission error
- 4: Invalid certificate or password

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

MIT License - see LICENSE file for details 