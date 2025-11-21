package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultConfigPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, ".airdash", "config.yaml")
	actual := getDefaultConfigPath()

	assert.Equal(t, expected, actual)
}

func TestGetPlistPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, "Library", "LaunchAgents", "com.github.ljagiello.airdash.plist")
	actual, err := getPlistPath()

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGetInstallBinaryPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, ".local", "bin", "airdash")
	actual, err := getInstallBinaryPath()

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestIsDaemonInstalled(t *testing.T) {
	// This test checks the function logic without modifying the actual system
	plistPath, err := getPlistPath()
	require.NoError(t, err)

	// Check if plist exists
	_, statErr := os.Stat(plistPath)
	expectedInstalled := statErr == nil

	actualInstalled := isDaemonInstalled()

	assert.Equal(t, expectedInstalled, actualInstalled)
}

func TestCopyFile(t *testing.T) {
	// Create temporary source file
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "destination.txt")

	testContent := []byte("test content for copy")
	err := os.WriteFile(srcPath, testContent, 0o644) //nolint:gosec // Test file
	require.NoError(t, err)

	// Copy file
	err = copyFile(srcPath, dstPath)
	require.NoError(t, err)

	// Verify destination exists and has same content
	dstContent, err := os.ReadFile(dstPath) //nolint:gosec // Test file read
	require.NoError(t, err)
	assert.Equal(t, testContent, dstContent)
}

func TestCopyFileSourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "nonexistent.txt")
	dstPath := filepath.Join(tmpDir, "destination.txt")

	err := copyFile(srcPath, dstPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "opening source file")
}

func TestCopyFileInvalidDestination(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")

	// Create source file
	err := os.WriteFile(srcPath, []byte("test"), 0o644) //nolint:gosec // Test file
	require.NoError(t, err)

	// Try to copy to invalid destination (directory that doesn't exist)
	dstPath := filepath.Join(tmpDir, "nonexistent", "subdir", "destination.txt")

	err = copyFile(srcPath, dstPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating destination file")
}
