package tests

import (
	"bytes"
	"errors"
	"testing"

	"github.com/Nicconike/AutomatedGo/v2/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockVersionChecker struct {
	mock.Mock
}

func (m *MockVersionChecker) GetCurrentVersion(versionFile, currentVersion string) (string, error) {
	args := m.Called(versionFile, currentVersion)
	return args.String(0), args.Error(1)
}

func (m *MockVersionChecker) GetLatestVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockVersionChecker) IsNewer(latestVersion, currentVersion string) bool {
	args := m.Called(latestVersion, currentVersion)
	return args.Bool(0)
}

func (m *MockVersionChecker) DownloadGo(version, os, arch, path string) error {
	args := m.Called(version, os, arch, path)
	return args.Error(0)
}

func TestRun(t *testing.T) {
	const version = "version.txt"
	tests := []struct {
		name           string
		versionFile    string
		currentVersion string
		targetOS       string
		targetArch     string
		input          string
		expectedOutput string
		expectedError  error
		mockSetup      func(m *MockVersionChecker)
	}{
		{
			name:          "No version specified",
			expectedError: errors.New("error: Either -file (-f) or -version (-v) must be specified"),
			mockSetup: func(m *MockVersionChecker) {
				// No specific setup required for this test case
			},
		},
		{
			name:           "Latest version available, user agrees to download",
			versionFile:    version,
			currentVersion: "",
			targetOS:       "linux",
			targetArch:     "amd64",
			input:          "yes\nyes\n", // Simulate user input for agreement and download path choice
			expectedOutput: "Current version: 1.0.0\nLatest version: 1.1.0\nA newer version is available\nDo you want to download the latest version? (yes/no): Do you want to download it to the current working directory? (yes/no): ",
			expectedError:  nil,
			mockSetup: func(m *MockVersionChecker) {
				m.On("GetCurrentVersion", version, "").Return("1.0.0", nil)
				m.On("GetLatestVersion").Return("1.1.0", nil)
				m.On("IsNewer", "1.1.0", "1.0.0").Return(true)
				m.On("DownloadGo", "1.1.0", "linux", "amd64", mock.Anything).Return(nil)
			},
		},
		{
			name:           "Latest version available, user declines to download",
			versionFile:    version,
			currentVersion: "",
			targetOS:       "linux",
			targetArch:     "amd64",
			input:          "no\n",
			expectedOutput: "Current version: 1.0.0\nLatest version: 1.1.0\nA newer version is available\nDo you want to download the latest version? (yes/no): Download aborted by user\n",
			expectedError:  nil,
			mockSetup: func(m *MockVersionChecker) {
				m.On("GetCurrentVersion", version, "").Return("1.0.0", nil)
				m.On("GetLatestVersion").Return("1.1.0", nil)
				m.On("IsNewer", "1.1.0", "1.0.0").Return(true)
			},
		},
		{
			name:           "No new version available",
			versionFile:    version,
			currentVersion: "",
			targetOS:       "linux",
			targetArch:     "amd64",
			input:          "",
			expectedOutput: "Current version: 1.1.0\nLatest version: 1.1.0\nYou have the latest version\n",
			expectedError:  nil,
			mockSetup: func(m *MockVersionChecker) {
				m.On("GetCurrentVersion", version, "").Return("1.1.0", nil)
				m.On("GetLatestVersion").Return("1.1.0", nil)
				m.On("IsNewer", "1.1.0", "1.1.0").Return(false)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockVersionChecker)
			tt.mockSetup(mockService)

			input := bytes.NewBufferString(tt.input)
			output := new(bytes.Buffer)

			err := pkg.Run(mockService, tt.versionFile, tt.currentVersion, tt.targetOS, tt.targetArch, input, output)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
