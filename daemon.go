package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed assets/launchd/com.github.ljagiello.airdash.plist
var plistTemplate string

const (
	launchAgentLabel = "com.github.ljagiello.airdash"
	plistFilename    = "com.github.ljagiello.airdash.plist"
)

// getPlistPath returns the path to the LaunchAgent plist file.
func getPlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", plistFilename), nil
}

// getInstallBinaryPath returns the path where the binary should be installed.
func getInstallBinaryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".local", "bin", "airdash"), nil
}

// installDaemon installs airdash as a LaunchAgent.
func installDaemon() error {
	// Get current executable path
	currentExec, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	// Resolve symlinks to get real path
	currentExec, err = filepath.EvalSymlinks(currentExec)
	if err != nil {
		return fmt.Errorf("resolving executable symlinks: %w", err)
	}

	// Check if config exists
	configPath := getDefaultConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found at %s\nPlease create it first - see README for instructions", configPath)
	}

	// Check if already installed
	plistPath, err := getPlistPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(plistPath); err == nil {
		return fmt.Errorf("daemon already installed\nPlist exists at: %s\nRun 'airdash uninstall' first to reinstall", plistPath)
	}

	// Determine installation path
	var installPath string
	if isRunningFromAppBundle(currentExec) {
		// Running from /Applications/AirDash.app - use that path directly
		installPath = currentExec
		logger.Info("Detected app bundle installation", "path", installPath)
	} else {
		// Running from standalone binary - copy to ~/.local/bin
		installPath, err = getInstallBinaryPath()
		if err != nil {
			return err
		}

		// Create ~/.local/bin directory if it doesn't exist
		installDir := filepath.Dir(installPath)
		if err := os.MkdirAll(installDir, 0o755); err != nil { //nolint:gosec // Standard macOS directory permissions
			return fmt.Errorf("creating install directory %s: %w", installDir, err)
		}

		// Copy binary to install location
		if err := copyFile(currentExec, installPath); err != nil {
			return fmt.Errorf("copying binary: %w", err)
		}

		// Make it executable
		if err := os.Chmod(installPath, 0o755); err != nil { //nolint:gosec // Binary needs to be executable
			return fmt.Errorf("making binary executable: %w", err)
		}
		logger.Info("Copied binary to install location", "path", installPath)
	}

	// Create LaunchAgents directory if it doesn't exist
	launchAgentsDir := filepath.Dir(plistPath)
	if err := os.MkdirAll(launchAgentsDir, 0o755); err != nil { //nolint:gosec // Standard macOS directory permissions
		return fmt.Errorf("creating LaunchAgents directory: %w", err)
	}

	// Prepare log file paths
	home, _ := os.UserHomeDir()
	logsDir := filepath.Join(home, "Library", "Logs")
	stdoutLog := filepath.Join(logsDir, "airdash.log")
	stderrLog := filepath.Join(logsDir, "airdash.error.log")

	// Create Logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0o755); err != nil { //nolint:gosec // Standard macOS directory permissions
		return fmt.Errorf("creating Logs directory: %w", err)
	}

	// Generate plist content from template
	plistContent := plistTemplate
	plistContent = strings.ReplaceAll(plistContent, "{{BINARY_PATH}}", installPath)
	plistContent = strings.ReplaceAll(plistContent, "{{CONFIG_PATH}}", configPath)
	plistContent = strings.ReplaceAll(plistContent, "{{STDOUT_LOG}}", stdoutLog)
	plistContent = strings.ReplaceAll(plistContent, "{{STDERR_LOG}}", stderrLog)

	// Write plist file
	if err := os.WriteFile(plistPath, []byte(plistContent), 0o644); err != nil { //nolint:gosec // Standard macOS file permissions
		return fmt.Errorf("writing plist file: %w", err)
	}

	// Load the LaunchAgent
	cmd := exec.Command("launchctl", "load", plistPath) //nolint:gosec // Intentional launchctl command with validated path
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up plist on failure
		_ = os.Remove(plistPath)
		return fmt.Errorf("loading LaunchAgent with launchctl: %w\nOutput: %s", err, string(output))
	}

	logger.Info("Daemon installed successfully",
		"binary", installPath,
		"plist", plistPath,
		"logs", stdoutLog,
	)

	return nil
}

// uninstallDaemon removes the airdash LaunchAgent.
func uninstallDaemon() error {
	plistPath, err := getPlistPath()
	if err != nil {
		return err
	}

	// Check if installed
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return fmt.Errorf("daemon not installed\nPlist not found at: %s", plistPath)
	}

	// Unload the LaunchAgent
	cmd := exec.Command("launchctl", "unload", plistPath) //nolint:gosec // Intentional launchctl command with validated path
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Don't fail if unload fails - service might not be running
		fmt.Printf("Warning: launchctl unload failed: %s\nContinuing with removal...\n", string(output))
	}

	// Remove plist file
	if err := os.Remove(plistPath); err != nil {
		return fmt.Errorf("removing plist file: %w", err)
	}

	// Get current executable to determine if we're in an app bundle
	currentExec, _ := os.Executable()
	currentExec, _ = filepath.EvalSymlinks(currentExec)
	inAppBundle := isRunningFromAppBundle(currentExec)

	// Print success message
	home, _ := os.UserHomeDir()
	fmt.Printf("Successfully uninstalled airdash daemon\n\n")
	fmt.Printf("Removed:\n")
	fmt.Printf("  Plist: %s\n", plistPath)

	// Only remove binary if not in app bundle
	if !inAppBundle {
		installPath, err := getInstallBinaryPath()
		if err == nil {
			if err := os.Remove(installPath); err != nil {
				if !os.IsNotExist(err) {
					fmt.Printf("Warning: failed to remove binary at %s: %v\n", installPath, err)
				}
			} else {
				fmt.Printf("  Binary: %s\n", installPath)
			}
		}
	} else {
		fmt.Printf("  Note: App bundle remains at its current location\n")
	}

	fmt.Printf("\nLog files remain at ~/Library/Logs/airdash*.log\n")
	fmt.Printf("Config remains at %s\n", filepath.Join(home, ".airdash", "config.yaml"))

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src) //nolint:gosec // Intentional file read from validated path
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			logger.Error("Closing source file", "error", closeErr)
		}
	}()

	destFile, err := os.Create(dst) //nolint:gosec // Intentional file write to validated path
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			logger.Error("Closing destination file", "error", closeErr)
		}
	}()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("copying file contents: %w", err)
	}

	return nil
}

// getDefaultConfigPath returns the default config file path.
func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".airdash", "config.yaml")
}

// isDaemonInstalled checks if the daemon is already installed.
func isDaemonInstalled() bool {
	plistPath, err := getPlistPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(plistPath)
	return err == nil
}

// isRunningFromAppBundle checks if the binary is running from an .app bundle.
func isRunningFromAppBundle(execPath string) bool {
	// Check if path contains .app/Contents/MacOS/
	return strings.Contains(execPath, ".app/Contents/MacOS/")
}
