# AirDash

> Display AirGradient air quality measurements in your macOS menu bar

[![Build Status](https://github.com/ljagiello/airdash/actions/workflows/go.yml/badge.svg)](https://github.com/ljagiello/airdash/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/v/release/ljagiello/airdash)](https://github.com/ljagiello/airdash/releases/latest)
[![License](https://img.shields.io/github/license/ljagiello/airdash)](LICENSE)

![Screenshot](screenshot.png)

## Features

- 🌡️ Real-time temperature, PM2.5, humidity, and CO2 levels in your menu bar
- 🔄 Auto-refresh every 60 seconds (configurable)
- 🎨 Clean, minimal menu bar interface
- ℹ️ About window with version info and GitHub link
- 🔐 Signed and notarized for macOS security
- 🍎 Native macOS app using AppKit

## Requirements

- macOS 11 (Big Sur) or later
- Apple Silicon (M1/M2/M3/M4)
- AirGradient sensor with API access

## Installation

### Download Pre-built Binary (Recommended)

1. Download the latest release from [Releases](https://github.com/ljagiello/airdash/releases/latest)
2. Extract the archive:
   ```bash
   tar -xzf airdash_0.0.1_darwin_arm64.tar.gz
   ```
3. Run the binary:
   ```bash
   ./airdash
   ```

The binary is signed and notarized - macOS will trust it automatically.

### Build from Source

```bash
git clone https://github.com/ljagiello/airdash.git
cd airdash
go build
./airdash
```

**Requirements:** Go 1.25.4+ and Xcode Command Line Tools

## Configuration

Create `~/.airdash/config.yaml`:

```yaml
# Required: Your AirGradient API token
token: your-secret-token-here

# Optional: Specific location ID (0 = all locations, default)
locationId: 0

# Optional: Update interval in seconds (default: 60)
interval: 60

# Optional: Temperature unit - "C" or "F" (default: "C")
tempUnit: F
```

### Getting Your API Token

1. Log in to [AirGradient Dashboard](https://app.airgradient.com/)
2. Navigate to **Settings** → **API**
3. Generate a new API token
4. Copy the token to your `config.yaml`

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `token` | string | *required* | Your AirGradient API token |
| `locationId` | int | `0` | Specific sensor location (0 = all) |
| `interval` | int | `60` | Update interval in seconds |
| `tempUnit` | string | `"C"` | Temperature unit: "C" or "F" |

## Usage

1. **Start AirDash:**
   ```bash
   ./airdash
   ```

2. **Menu Bar:** Look for your current measurements displayed in the menu bar

3. **Menu Options:**
   - **About AirDash**: View version, build info, and GitHub link
   - **Quit**: Exit the application

The app runs in the background and updates automatically. It will fetch data immediately on startup and then refresh based on your configured interval.

## Troubleshooting

### No measurements showing

**Check your configuration:**
- Verify `~/.airdash/config.yaml` exists with a valid API token
- Ensure `locationId` matches your sensor (or use `0` for all sensors)

**Check the logs:**
```bash
# View recent logs
log show --predicate 'process == "airdash"' --last 5m

# View with more detail
log show --predicate 'process == "airdash"' --info --last 1h
```

**Common issues:**
- Invalid or expired API token
- Sensor offline or not reporting data
- Network connectivity issues

### App won't open or shows security warning

The app is signed and notarized. If you see a security warning:
1. Right-click the app and select **Open**
2. Click **Open** in the security dialog
3. Alternatively: System Settings → Privacy & Security → Allow

### HTTP/API Errors

If you see HTTP errors in logs:
- Check your internet connection
- Verify AirGradient API status at https://status.airgradient.com
- Confirm your API token is valid and hasn't expired
- Try accessing the API directly:
  ```bash
  curl "https://api.airgradient.com/public/api/v1/locations/measures/current?token=YOUR_TOKEN"
  ```

### App crashes or freezes

- Ensure you're on macOS 11 or later
- Check for updates: [Releases](https://github.com/ljagiello/airdash/releases)
- Report issues with crash logs: [GitHub Issues](https://github.com/ljagiello/airdash/issues)

## Development

### Local Development

```bash
# Clone the repository
git clone https://github.com/ljagiello/airdash.git
cd airdash

# Install dependencies
go mod download

# Run tests
go test -v ./...

# Run linter
golangci-lint run

# Build locally
go build -o airdash .

# Run
./airdash
```

### Project Structure

```
airdash/
├── main.go           # macOS UI and app entry point
├── airgradient.go    # AirGradient API client
├── config.go         # Configuration loading
├── log.go            # Structured logging
├── assets/           # Embedded assets (logo)
├── testdata/         # Test fixtures
└── .github/          # CI/CD workflows
```

### Running Tests with Coverage

```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Creating a Release

Releases are automated via GitHub Actions. To create a new release:

1. Ensure all changes are committed and pushed to `main`
2. Create and push a version tag:
   ```bash
   git tag -a v0.0.2 -m "Release v0.0.2"
   git push origin v0.0.2
   ```

The GitHub Actions workflow will automatically:
- Build the arm64 binary
- Sign with Apple Developer ID
- Notarize with Apple
- Create a GitHub release with downloadable artifacts

**Release workflow:** `.github/workflows/release.yml`

### Required GitHub Secrets (for maintainers)

For signing and notarization, configure these secrets in your repository:

| Secret | Description |
|--------|-------------|
| `MACOS_SIGN_P12_BASE64` | Base64-encoded Developer ID certificate (.p12) |
| `MACOS_SIGN_PASSWORD` | Password for the .p12 certificate |
| `MACOS_NOTARY_KEY_BASE64` | Base64-encoded App Store Connect API key (.p8) |
| `MACOS_NOTARY_KEY_ID` | App Store Connect API Key ID |
| `MACOS_NOTARY_ISSUER_ID` | App Store Connect Issuer ID (UUID) |

See [Apple Developer documentation](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution) for details on obtaining certificates and keys.

## API

AirDash uses the [AirGradient Public API v1](https://api.airgradient.com/public/docs/api/v1/).

**Endpoints used:**
- `GET /locations/measures/current` - All locations
- `GET /locations/{locationId}/measures/current` - Specific location

## Contributing

Contributions are welcome! Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

**Guidelines:**
- Write tests for new features
- Run `golangci-lint run` before submitting
- Follow existing code style
- Update documentation as needed

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [DarwinKit](https://github.com/progrium/darwinkit) for macOS AppKit bindings
- Air quality data from [AirGradient](https://www.airgradient.com/)
- Logo design inspired by air quality monitoring
