package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/Nicconike/AutomatedGo/v2/pkg"
)

func runGetCurrentVersionTest(t *testing.T, filePath, directVersion, expectedResult string, expectError bool) {
	t.Helper()
	result, err := pkg.GetCurrentVersion(filePath, directVersion)
	if (err != nil) != expectError {
		t.Errorf("GetCurrentVersion() error = %v, expectError %v", err, expectError)
		return
	}
	if result != expectedResult {
		t.Errorf("GetCurrentVersion() = %v, want %v", result, expectedResult)
	}
}

func createTempFile(t *testing.T, content string) *os.File {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "version")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	return tmpfile
}

func TestGetCurrentVersion(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		directVersion  string
		expectedResult string
		expectError    bool
	}{
		{"Direct version", "", "1.16.5", "1.16.5", false},
		{"No input", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runGetCurrentVersionTest(t, tt.filePath, tt.directVersion, tt.expectedResult, tt.expectError)
		})
	}

	t.Run("Valid file path", func(t *testing.T) {
		tmpfile := createTempFile(t, "go 1.18")
		defer os.Remove(tmpfile.Name())

		runGetCurrentVersionTest(t, tmpfile.Name(), "", "1.18", false)
	})

	t.Run("Invalid file path", func(t *testing.T) {
		runGetCurrentVersionTest(t, "/non/existent/path", "", "", true)
	})
}

func TestExtractGoVersion(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"Dockerfile", "FROM golang:1.17.5", "1.17.5"},
		{"go.mod", "go 1.17", "1.17"},
		{"JSON", `{"go_version": "1.18.0"}`, "1.18.0"},
		{"No version", "Some random content", ""},
		{"Version with equals", "go_version = 1.17.1", "1.17.1"},
		{"Golang version", "golang_version: 1.18.0", "1.18.0"},
		{"Version without prefix", "1.19.0", "1.19.0"},
		{"Dockerfile ARG", "ARG GO_VERSION=1.20.0", "1.20.0"},
		{"Dockerfile ENV", "ENV GO_VERSION=1.21.0", "1.21.0"},
		{"JSON with goVersion", `{"goVersion": "1.22.0"}`, "1.22.0"},
		{"JSON with golangVersion", `{"golangVersion": "1.23.0"}`, "1.23.0"},
		{"JSON with GO_VERSION", `{"GO_VERSION": "1.24.0"}`, "1.24.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pkg.ExtractGoVersion(tt.content)
			if result != tt.expected {
				t.Errorf("ExtractGoVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReadVersionFromFile(t *testing.T) {
	// Create a temporary file for testing
	content := []byte("go 1.17")
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test reading from the file
	version, err := pkg.ReadVersionFromFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != "1.17" {
		t.Errorf("Expected version 1.17, got %s", version)
	}

	// Test reading from non-existent file
	_, err = pkg.ReadVersionFromFile("non_existent_file.txt")
	if err == nil {
		t.Error("Expected an error for non-existent file, got nil")
	}

	// Test file with no extractable version
	noVersionFile, err := os.CreateTemp("", "noversion")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(noVersionFile.Name())

	if _, err := noVersionFile.Write([]byte("No version here")); err != nil {
		t.Fatal(err)
	}
	noVersionFile.Close()

	result, err := pkg.ReadVersionFromFile(noVersionFile.Name())
	if result != "" || (err == nil || !strings.Contains(err.Error(), "unable to extract Go version from file")) {
		t.Errorf("Expected 'unable to extract Go version from file' error, got %v with result %s", err, result)
	}
}
