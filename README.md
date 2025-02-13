# LC_utils - LimaCharlie Sensor Management Tool

A modern command-line interface (CLI) tool for managing LimaCharlie sensors with advanced features and a cyberpunk-inspired interface.

## Features

- 🔍 List and filter sensors by various criteria
- 🏷️ Manage sensor tags (add/remove)
- 📤 Upload files to sensors
- 🖥️ Execute commands on sensors
- 🔄 Support for reliable tasking
- 🎨 Multiple visual themes (Matrix, Hacker, Cyberpunk, Retro)
- 📊 Multiple output formats (Text, JSON, CSV)

## Installation

```bash
go install github.com/ubercylon8/LC_utils/cmd/lc-sensors@latest
```

Or build from source:

```bash
git clone https://github.com/ubercylon8/LC_utils.git
cd LC_utils
go build ./cmd/lc-sensors
```

## Quick Start

1. Set your LimaCharlie credentials:
```bash
export LC_ORG_ID="your-org-id"
export LC_API_KEY="your-api-key"
```

2. List all sensors:
```bash
lc-sensors list
```

3. Filter sensors by platform and hostname:
```bash
lc-sensors list --filter-platform windows --filter-hostname "web-*"
```

4. Add tags to sensors:
```bash
lc-sensors tag --sensor-id SID --add-tags web-server,production
```

## Usage Examples

### List Sensors
```bash
# List all sensors in JSON format
lc-sensors list -f json

# Show only online sensors
lc-sensors list --online

# Filter by platform and show tags
lc-sensors list --filter-platform windows --tags
```

### Execute Commands
```bash
# Run a command on specific sensors
lc-sensors task run --filter-hostname "web-*" --command "whoami"

# Upload a file to sensors
lc-sensors task put --filter-hostname "db-*" --payload-name config.yaml --payload-path "/etc/config.yaml"
```

### Manage Tags
```bash
# Tag multiple sensors
lc-sensors tag-multiple --filter-platform windows --add-tags windows-servers

# Remove tags from a sensor
lc-sensors tag --sensor-id SID --remove-tags old-config
```

## Security

- All sensitive operations require proper authentication
- API keys are handled securely and never logged
- Support for investigation IDs for audit trails
- Secure file upload mechanisms

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details 