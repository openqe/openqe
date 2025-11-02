package polarion

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from a YAML file with Jinja2 templating support
func LoadConfig(configFile string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configFile)
	}

	// Read the config file
	configContent, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Process through Jinja2 template with keyring support
	renderer := NewTemplateRenderer()
	renderedContent, err := renderer.Render(string(configContent), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to render config template: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal([]byte(renderedContent), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Set defaults if not specified
	if config.TestCasesFile == "" {
		config.TestCasesFile = "test_cases.yaml"
	}

	return &config, nil
}

// LoadTestCases loads test cases from a YAML file
func LoadTestCases(testCasesFile string) ([]TestCase, error) {
	// Support relative paths - resolve relative to current directory
	if !filepath.IsAbs(testCasesFile) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		testCasesFile = filepath.Join(cwd, testCasesFile)
	}

	// Check if file exists
	if _, err := os.Stat(testCasesFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("test cases file not found: %s", testCasesFile)
	}

	// Read the file
	content, err := os.ReadFile(testCasesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read test cases file: %w", err)
	}

	// Parse YAML
	var data TestCasesData
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML test cases: %w", err)
	}

	return data.TestCases, nil
}
