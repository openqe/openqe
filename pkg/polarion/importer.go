package polarion

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"os"
	"strings"
)

// Importer handles importing test cases to Polarion
type Importer struct {
	config        *Config
	client        *Client
	logger        *log.Logger
	testCasesFile string
}

// NewImporter creates a new Polarion importer
func NewImporter(configFile string, verbose bool) (*Importer, error) {
	// Load configuration
	config, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	// Override verbose setting if provided
	if verbose {
		config.Verbose = true
	}

	// Create API client
	client, err := NewClient(config, config.Verbose)
	if err != nil {
		return nil, err
	}

	importer := &Importer{
		config:        config,
		client:        client,
		testCasesFile: config.TestCasesFile,
	}

	// Setup logger
	if config.Verbose {
		importer.logger = log.New(os.Stdout, "[IMPORTER] ", log.LstdFlags)
	} else {
		importer.logger = log.New(os.Stdout, "", 0)
	}

	return importer, nil
}

// SetTestCasesFile sets the test cases file path
func (i *Importer) SetTestCasesFile(file string) {
	i.testCasesFile = file
}

// TestConnection tests the connection to Polarion
func (i *Importer) TestConnection() error {
	i.logger.Println("Testing connection to Polarion server...")
	return i.client.TestConnection()
}

// ImportAll imports all test cases
func (i *Importer) ImportAll(dryRun bool) error {
	// Test connection first (skip in dry-run mode)
	if !dryRun {
		if err := i.TestConnection(); err != nil {
			return fmt.Errorf("connection test failed: %w", err)
		}
	} else {
		i.logger.Println("DRY RUN MODE: Skipping connection test")
	}

	// Load test cases
	testCases, err := LoadTestCases(i.testCasesFile)
	if err != nil {
		return err
	}

	i.logger.Printf("Loaded %d test cases\n", len(testCases))

	// Import each test case
	successCount := 0
	failCount := 0

	for _, testCase := range testCases {
		if err := i.createTestCase(&testCase, dryRun); err != nil {
			i.logger.Printf("✗ Failed to create test case %s: %v\n", testCase.ID, err)
			failCount++
		} else {
			successCount++
		}
	}

	// Summary
	fmt.Println("\n" + "============================================================")
	fmt.Println("Import Summary:")
	fmt.Printf("  Total:   %d\n", len(testCases))
	fmt.Printf("  Success: %d\n", successCount)
	fmt.Printf("  Failed:  %d\n", failCount)
	fmt.Println("============================================================")

	if failCount > 0 {
		return fmt.Errorf("import completed with %d failures", failCount)
	}

	return nil
}

// createTestCase creates a single test case in Polarion
func (i *Importer) createTestCase(testCase *TestCase, dryRun bool) error {
	i.logger.Printf("Creating test case: %s - %s\n", testCase.ID, testCase.Title)

	// Build work item payload (without test steps)
	workItemPayload := i.buildWorkItemPayload(testCase)

	if dryRun {
		i.logger.Println("DRY RUN: Would create test case with payload:")
		payloadJSON, _ := json.MarshalIndent(workItemPayload, "", "  ")
		fmt.Println(string(payloadJSON))

		// Also show test steps payload
		if len(testCase.Steps) > 0 {
			testStepsPayload := i.buildTestStepsPayload(testCase.Steps)
			i.logger.Println("\nDRY RUN: Would add test steps with payload:")
			stepsJSON, _ := json.MarshalIndent(testStepsPayload, "", "  ")
			fmt.Println(string(stepsJSON))
		}

		return nil
	}

	// Create work item
	response, err := i.client.CreateWorkItem(workItemPayload)
	if err != nil {
		return err
	}

	if len(response.Data) == 0 {
		return fmt.Errorf("no work item returned in response")
	}

	workItemID := response.Data[0].ID
	i.logger.Printf("✓ Successfully created work item: %s (ID: %s)\n", testCase.ID, workItemID)

	// Add test steps if present
	if len(testCase.Steps) > 0 {
		testStepsPayload := i.buildTestStepsPayload(testCase.Steps)
		// Extract just the ID part (e.g., "OCP-85835" from "OSE/OCP-85835")
		workItemIDOnly := extractWorkItemID(workItemID)
		if err := i.client.AddTestSteps(workItemIDOnly, testStepsPayload); err != nil {
			return fmt.Errorf("work item created but failed to add test steps: %w", err)
		}
		i.logger.Printf("✓ Successfully added %d test steps\n", len(testCase.Steps))
	}

	return nil
}

// buildWorkItemPayload builds the work item creation payload (without test steps)
func (i *Importer) buildWorkItemPayload(testCase *TestCase) *WorkItemPayload {
	defaults := i.config.Polarion.TestCase.Defaults

	// Build description HTML
	description := i.buildDescriptionHTML(testCase)

	// Build attributes
	attributes := map[string]interface{}{
		"type":  i.config.Polarion.TestCase.WorkItemType,
		"title": testCase.Title,
		"description": TextContent{
			Type:  "text/html",
			Value: description,
		},
		"caseID": testCase.ID,
	}

	// Add category if present
	if testCase.Category != "" {
		attributes["category"] = testCase.Category
	}

	// Add priority (test case specific or default)
	if testCase.Priority != "" {
		attributes["priority"] = testCase.Priority
	} else if priority, ok := defaults["priority"]; ok {
		attributes["priority"] = priority
	}

	// Add type if present
	if testCase.Type != "" {
		attributes["type"] = testCase.Type
	}

	// Add component if present
	if testCase.Component != "" {
		attributes["component"] = testCase.Component
	}

	// Add status if present
	if testCase.Status != "" {
		attributes["status"] = testCase.Status
	}

	// Add level if present
	if testCase.Level != "" {
		attributes["level"] = testCase.Level
	}

	// Add test_type if present
	if testCase.TestType != "" {
		attributes["testType"] = testCase.TestType
	}

	// Add custom attributes (convert snake_case to camelCase)
	for key, value := range testCase.Attributes {
		if value != "" {
			camelKey := snakeToCamel(key)
			attributes[camelKey] = value
			i.logger.Printf("  Adding custom attribute %s (as %s): %v\n", key, camelKey, value)
		}
	}

	// Add all default fields (convert snake_case to camelCase)
	for key, value := range defaults {
		camelKey := snakeToCamel(key)
		if _, exists := attributes[camelKey]; !exists && value != nil {
			attributes[camelKey] = value
			i.logger.Printf("  Adding default %s (as %s): %v\n", key, camelKey, value)
		}
	}

	return &WorkItemPayload{
		Data: []WorkItemData{
			{
				Type:       "workitems",
				Attributes: attributes,
			},
		},
	}
}

// buildDescriptionHTML builds the HTML description for the work item
func (i *Importer) buildDescriptionHTML(testCase *TestCase) string {
	description := `<div class="test-case-description">`

	description += `<h3>Description</h3>`
	description += fmt.Sprintf(`<p>%s</p>`, escapeHTML(testCase.Description))

	description += `</div>`

	return description
}

// buildTestStepsPayload builds the test steps payload according to Polarion API format
func (i *Importer) buildTestStepsPayload(steps []TestStep) *TestStepsPayload {
	var data []TestStepData

	for _, step := range steps {
		// Build step content from description
		stepContent := fmt.Sprintf("<p>%s</p>", escapeHTML(step.Description))

		// Build expected result
		expectedContent := fmt.Sprintf("<p>%s</p>", escapeHTML(step.Expected))

		// Create test step data with the format specified by the user
		stepData := TestStepData{
			Type: "teststeps",
			Attributes: TestStepAttributes{
				Keys: []string{"step", "expectedResult"},
				Values: []interface{}{
					TextContent{
						Type:  "text/html",
						Value: stepContent,
					},
					TextContent{
						Type:  "text/html",
						Value: expectedContent,
					},
				},
			},
		}

		data = append(data, stepData)
	}

	return &TestStepsPayload{
		Data: data,
	}
}

// escapeHTML escapes HTML special characters
func escapeHTML(text string) string {
	return html.EscapeString(text)
}

// snakeToCamel converts snake_case to camelCase
func snakeToCamel(s string) string {
	if s == "" {
		return s
	}

	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		// No underscore, return as is
		return s
	}

	// First part stays lowercase, capitalize first letter of remaining parts
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return result
}

// extractWorkItemID extracts the work item ID from the full ID returned by Polarion
// For example: "OSE/OCP-85835" -> "OCP-85835"
func extractWorkItemID(fullID string) string {
	// If there's a slash, take the part after it
	if idx := strings.LastIndex(fullID, "/"); idx >= 0 {
		return fullID[idx+1:]
	}
	// Otherwise return the full ID as is
	return fullID
}
