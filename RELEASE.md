# Release Guide

This document describes how to create signed and notarized macOS releases for AirDash.

## Prerequisites

You need to obtain the following from your Apple Developer Account:

### 1. Developer ID Application Certificate

1. Go to [Apple Developer Certificates](https://developer.apple.com/account/resources/certificates/list)
2. Create a new "Developer ID Application" certificate
3. Download and install it in Keychain Access
4. Export the certificate as a `.p12` file with a password
5. Convert to base64:
   ```bash
   base64 -i certificate.p12 | pbcopy
   ```

### 2. App Store Connect API Key

1. Go to [App Store Connect API Keys](https://appstoreconnect.apple.com/access/api)
2. Create a new API key with "App Manager" role
3. Download the `.p8` file
4. Note the Key ID and Issuer ID
5. Convert to base64:
   ```bash
   base64 -i AuthKey_XXXXXXXXXX.p8 | pbcopy
   ```

## GitHub Secrets Setup

Add the following secrets to your GitHub repository (Settings → Secrets and variables → Actions):

| Secret Name | Description | How to Get |
|-------------|-------------|------------|
| `MACOS_SIGN_P12_BASE64` | Base64-encoded Developer ID Application certificate (.p12) | Export certificate from Keychain, convert to base64 |
| `MACOS_SIGN_PASSWORD` | Password for the .p12 certificate | The password you set when exporting |
| `MACOS_NOTARY_KEY_BASE64` | Base64-encoded App Store Connect API key (.p8) | Download from App Store Connect, convert to base64 |
| `MACOS_NOTARY_KEY_ID` | App Store Connect API Key ID | Found in App Store Connect (e.g., `ABC123DEFG`) |
| `MACOS_NOTARY_ISSUER_ID` | App Store Connect Issuer ID | Found in App Store Connect (UUID format) |

## Creating a Release

1. Ensure all changes are committed and pushed to `main`
2. Create and push a tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. The GitHub Actions workflow will automatically:
   - Build universal binaries (Apple Silicon + Intel)
   - Sign the binaries with your Developer ID
   - Notarize the binaries with Apple
   - Create a GitHub release with the signed artifacts

## Verification

After the release is created, verify the signature:

```bash
# Download and extract the release
tar -xzf airdash_1.0.0_darwin_arm64.tar.gz

# Verify the signature
codesign -vvv --deep --strict airdash
spctl -a -vvv -t install airdash
```

You should see output indicating the binary is properly signed and notarized.

## Troubleshooting

### Notarization Fails

- Verify all secrets are correctly set in GitHub
- Check that the API key has the correct permissions
- Ensure the certificate is a "Developer ID Application" certificate, not a distribution certificate

### Binary Not Signed

- Check the GoReleaser output in the GitHub Actions logs
- Verify the certificate password is correct
- Ensure quill is properly installed

### Universal Binary Issues

- Make sure CGO is enabled for both architectures
- Verify darwinkit dependencies are compatible with both arm64 and amd64

## Manual Local Release (Testing)

To test the release process locally:

```bash
# Install GoReleaser and quill
brew install goreleaser quill

# Export required environment variables
export MACOS_SIGN_P12_PATH=/path/to/certificate.p12
export MACOS_SIGN_PASSWORD="your-password"
export MACOS_NOTARY_KEY_PATH=/path/to/AuthKey.p8
export MACOS_NOTARY_KEY_ID="your-key-id"
export MACOS_NOTARY_ISSUER_ID="your-issuer-id"

# Run GoReleaser in snapshot mode (no push to GitHub)
goreleaser release --snapshot --clean
```
