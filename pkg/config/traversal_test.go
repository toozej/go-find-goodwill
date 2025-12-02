package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryTraversalProtection(t *testing.T) {
	tests := []struct {
		name        string
		yamlPath    string
		expectError bool
	}{
		{
			name:        "Normal YAML file should work",
			yamlPath:    "config.yaml",
			expectError: false,
		},
		{
			name:        "Directory traversal attempt should fail",
			yamlPath:    "../../../../etc/passwd",
			expectError: true,
		},
		{
			name:        "Invalid file extension should fail",
			yamlPath:    "config.txt",
			expectError: true,
		},
		{
			name:        "Empty path should fail",
			yamlPath:    "",
			expectError: true,
		},
		{
			name:        "Path with null bytes should fail",
			yamlPath:    "config\000.yaml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir := t.TempDir()

			// For valid test cases, create the YAML file
			if !tt.expectError && tt.yamlPath != "" {
				yamlContent := `shopgoodwill:
  username: "testuser"
  password: "testpass"
  api_base_url: "https://api.shopgoodwill.com"
  max_retries: 3
  request_timeout: "30s"`

				filePath := filepath.Join(tmpDir, tt.yamlPath)
				err := os.WriteFile(filePath, []byte(yamlContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(filePath)

				// Change to temp directory for the test
				originalDir, err := os.Getwd()
				if err != nil {
					t.Fatalf("Failed to get current directory: %v", err)
				}
				defer func() {
					if err := os.Chdir(originalDir); err != nil {
						t.Logf("Failed to restore original directory: %v", err)
					}
				}()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("Failed to change to temp directory: %v", err)
				}

				// Test the function
				_, err = LoadYAMLConfig(tt.yamlPath)
				if tt.expectError && err == nil {
					t.Errorf("Expected error but got none")
				} else if !tt.expectError && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			} else {
				// For invalid cases, just test the validation
				_, err := LoadYAMLConfig(tt.yamlPath)
				if tt.expectError && err == nil {
					t.Errorf("Expected error but got none")
				} else if !tt.expectError && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
