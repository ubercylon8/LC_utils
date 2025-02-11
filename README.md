# LimaCharlie Utilities

A collection of command-line utilities for interacting with the LimaCharlie platform.

## Installation

1. Clone the repository:
```bash
git clone https://github.com/f0rt1ka/lc-utils.git
cd lc-utils
```

2. Build the utilities:
```bash
# For macOS
GOOS=darwin GOARCH=amd64 go build -o bin/lc-sensors cmd/lc-sensors/main.go

# For Linux
GOOS=linux GOARCH=amd64 go build -o bin/lc-sensors cmd/lc-sensors/main.go

# For Windows
GOOS=windows GOARCH=amd64 go build -o bin/lc-sensors.exe cmd/lc-sensors/main.go
```

## Environment Variables

The following environment variables can be used instead of command-line flags:

- `LC_ORG_ID`: LimaCharlie Organization ID (overrides --oid flag)
- `LC_API_KEY`: LimaCharlie API Key (overrides --api-key flag)

Example using environment variables:
```bash
# Set environment variables
export LC_ORG_ID=your_org_id
export LC_API_KEY=your_api_key

# Now you can run commands without specifying -o and -k flags
lc-sensors list
lc-sensors list --online -t
lc-sensors tag-multiple --filter-platform windows --filter-hostname "*web*" --add-tags web-server
```

## Available Utilities

### lc-sensors

Lists all sensors in your LimaCharlie organization with detailed information including platform, architecture, version, and status.

#### Usage

```bash
# List all sensors
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY

# List online sensors with tags
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY --online -t

# List Windows sensors with specific hostname pattern
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY --filter-platform windows --filter-hostname "*web*"

# Filter sensors by tag (supports wildcards)
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY --filter-tag "web*" -t

# Add tags to a sensor
lc-sensors tag -o YOUR_ORG_ID -k YOUR_API_KEY --sensor-id SENSOR_ID --add-tags tag1,tag2

# Remove tags from a sensor
lc-sensors tag -o YOUR_ORG_ID -k YOUR_API_KEY --sensor-id SENSOR_ID --remove-tags tag1,tag2

# Tag multiple sensors based on filters
lc-sensors tag-multiple -o YOUR_ORG_ID -k YOUR_API_KEY --filter-platform windows --filter-hostname "*web*" --add-tags web-server

# Output in JSON format
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY -f json

# Output in CSV format
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY -f csv

# Just show the cool banner
lc-sensors --fun

# Show Matrix-style animation
lc-sensors --matrix

# Show hacking animation
lc-sensors --hack

# Use different theme
lc-sensors --theme cyberpunk --fun

# Show available commands and help
lc-sensors --help

# Show help for a specific command
lc-sensors list --help
lc-sensors tag --help
```

#### Global Options

- `-o, --oid`: LimaCharlie Organization ID (required)
- `-k, --api-key`: LimaCharlie API Key (required)
- `--fun`: Just show the cool banner
- `--matrix`: Show Matrix-style animation
- `--hack`: Show hacking animation
- `--theme`: Visual theme (matrix, hacker, cyberpunk, retro)

#### List Command Options

- `-f, --output`: Output format (text/json/csv) (default: "text")
- `-l, --limit`: Limit the number of results
- `-t, --tags`: Include sensor tags in output
- `-i, --ip`: Filter sensors by IP address
- `-n, --hostname`: Filter sensors by hostname prefix
- `--filter-hostname`: Filter by hostname (supports wildcards *)
- `--filter-platform`: Filter by platform (windows, macos)
- `--filter-tag`: Filter by tag
- `--online`: Show only online sensors

#### Tag Command Options

- `--sensor-id`: Sensor ID to tag (required)
- `--add-tags`: Tags to add (comma-separated)
- `--remove-tags`: Tags to remove (comma-separated)

#### Examples

```bash
# List all sensors with text output (default)
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY

# List online sensors with tags
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY --online -t

# List Windows sensors with specific hostname pattern
lc-sensors list -o YOUR_ORG_ID -k YOUR_API_KEY --filter-platform windows --filter-hostname "*web*"

# Add tags to a sensor
lc-sensors tag -o YOUR_ORG_ID -k YOUR_API_KEY --sensor-id SENSOR_ID --add-tags production,web-server

# Remove tags from a sensor
lc-sensors tag -o YOUR_ORG_ID -k YOUR_API_KEY --sensor-id SENSOR_ID --remove-tags old,deprecated

# Add and remove tags in a single command
lc-sensors tag -o YOUR_ORG_ID -k YOUR_API_KEY --sensor-id SENSOR_ID --add-tags new,active --remove-tags old,inactive
```

#### Output Formats

1. **Text (default)**: Human-readable format with colored output
2. **JSON**: Machine-readable JSON format
3. **CSV**: Comma-separated values format with headers

## Security Note

- Never share your API key
- Store API keys securely
- Use environment variables for sensitive information in production environments
- The API key should have appropriate permissions in LimaCharlie

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

# Task Command Options

- `--sensor-id`: Sensor ID to task (required)
- `--path`: Path on the sensor where to write the file (required for put)
- `--content`: Content to write to the file (required for put)
- `--command`: Command to execute (required for run)
- `--filter-hostname`: Filter sensors by hostname (supports wildcards *)
- `--filter-tag`: Filter sensors by tag (supports wildcards *)
- `--filter-platform`: Filter by platform (windows, macos, linux)
- `--reliable`: Use reliable tasking (will retry if sensor is offline)
- `--random-delay`: Add random delay between commands (5-15 seconds)
- `--investigation-id`: Investigation ID to tag the task with
- `--context`: Context value for reliable tasking (only used with --reliable)

Examples:
```bash
# Upload a file to all Windows sensors with hostname matching "web-*"
lc-sensors task put -o YOUR_ORG_ID -k YOUR_API_KEY --filter-platform windows --filter-hostname "web-*" --payload-name file.txt --payload-path "/tmp/file.txt"

# Upload a file to all macOS sensors with a specific tag
lc-sensors task put -o YOUR_ORG_ID -k YOUR_API_KEY --filter-platform macos --filter-tag "developer" --payload-name file.txt --payload-path "/tmp/file.txt"

# Upload multiple files from a command list with random delay
lc-sensors task put -o YOUR_ORG_ID -k YOUR_API_KEY --filter-platform windows --filter-hostname "web-*" --command-list uploads.txt --random-delay
``` 