package polarion

import (
	"bufio"
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
	autoConfirm   bool
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

// SetAutoConfirm sets whether to automatically confirm prompts
func (i *Importer) SetAutoConfirm(autoConfirm bool) {
	i.autoConfirm = autoConfirm
}

// TestConnection tests the connection to Polarion
func (i *Importer) TestConnection() error {
	i.logger.Println("Testing connection to Polarion server...")
	return i.client.TestConnection()
}

// InspectWorkItem retrieves and displays a work item's structure
func (i *Importer) InspectWorkItem(workItemID string) error {
	i.logger.Printf("Fetching work item: %s\n", workItemID)

	workItem, err := i.client.GetWorkItem(workItemID)
	if err != nil {
		return fmt.Errorf("failed to get work item: %w", err)
	}

	if workItem == nil {
		return fmt.Errorf("work item not found: %s", workItemID)
	}

	// Pretty print the work item
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("Work Item: %s\n", workItemID)
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("\nType: %s\n", workItem.Type)
	fmt.Printf("ID: %s\n", workItem.ID)

	fmt.Println("\nAttributes:")
	fmt.Println(strings.Repeat("-", 80))

	// Pretty print attributes in JSON format
	attributesJSON, err := json.MarshalIndent(workItem.Attributes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	fmt.Println(string(attributesJSON))
	fmt.Println(strings.Repeat("=", 80))

	// Highlight key fields
	fmt.Println("\nKey Field Values:")
	fmt.Println(strings.Repeat("-", 80))

	keyFields := []string{"component", "level", "testType", "type", "title", "status", "priority"}
	for _, field := range keyFields {
		if value, ok := workItem.Attributes[field]; ok {
			fmt.Printf("  %-15s: %v\n", field, value)
		}
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")

	return nil
}

// ImportResult represents the result of importing a single test case
type ImportResult struct {
	TestCaseID string
	WorkItemID string
	Action     string // "created" or "updated"
	Error      error
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

	// Import each test case and collect results
	var results []ImportResult

	for _, testCase := range testCases {
		result := i.createTestCase(&testCase, dryRun)
		results = append(results, result)
	}

	// Count successes and failures
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		} else {
			failCount++
		}
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("Import Summary:")
	fmt.Printf("  Total:   %d\n", len(testCases))
	fmt.Printf("  Success: %d\n", successCount)
	fmt.Printf("  Failed:  %d\n", failCount)
	fmt.Println(strings.Repeat("=", 80))

	// Display successful imports with links
	if successCount > 0 {
		fmt.Println("\nSuccessful Imports:")
		fmt.Println(strings.Repeat("-", 80))
		for _, r := range results {
			if r.Error == nil {
				workItemURL := i.buildWorkItemURL(r.WorkItemID)
				fmt.Printf("  ✓ %s (%s): %s\n", r.TestCaseID, r.Action, workItemURL)
			}
		}
	}

	// Display failures
	if failCount > 0 {
		fmt.Println("\nFailed Imports:")
		fmt.Println(strings.Repeat("-", 80))
		for _, r := range results {
			if r.Error != nil {
				fmt.Printf("  ✗ %s: %v\n", r.TestCaseID, r.Error)
			}
		}
		fmt.Println(strings.Repeat("=", 80))
		return fmt.Errorf("import completed with %d failures", failCount)
	}

	fmt.Println(strings.Repeat("=", 80))
	return nil
}

// buildWorkItemURL builds the Polarion web UI URL for a work item
func (i *Importer) buildWorkItemURL(workItemID string) string {
	// Extract just the ID part if it has project prefix
	itemID := extractWorkItemID(workItemID)
	// Polarion URL format: https://server/polarion/#/project/PROJECT_ID/workitem?id=ITEM_ID
	return fmt.Sprintf("%s/polarion/#/project/%s/workitem?id=%s",
		strings.TrimSuffix(i.config.Polarion.ServerURL, "/"),
		i.config.Polarion.ProjectID,
		itemID)
}

// createTestCase creates a single test case in Polarion and returns the result
func (i *Importer) createTestCase(testCase *TestCase, dryRun bool) ImportResult {
	i.logger.Printf("Processing test case: %s - %s\n", testCase.ID, testCase.Title)

	result := ImportResult{
		TestCaseID: testCase.ID,
	}

	if dryRun {
		// Build work item payload (without test steps)
		workItemPayload := i.buildWorkItemPayload(testCase, "")
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

		result.WorkItemID = testCase.ID
		result.Action = "dry-run"
		return result
	}

	// Check if work item already exists
	i.logger.Printf("Checking if work item %s already exists...\n", testCase.ID)
	existingWorkItemData, err := i.client.GetWorkItem(testCase.ID)
	if err != nil {
		result.Error = fmt.Errorf("failed to check if work item exists: %w", err)
		i.logger.Printf("✗ Failed to create test case %s: %v\n", testCase.ID, result.Error)
		return result
	}

	var workItemID string
	var workItemCreated bool

	if existingWorkItemData != nil {
		// Work item already exists - update it
		workItemID = existingWorkItemData.ID
		i.logger.Printf("⚠ Work item already exists: %s (ID: %s). Updating work item...\n", testCase.ID, workItemID)

		// Extract just the ID part (e.g., "TEST-123" from "PRJ/TEST-123")
		workItemIDOnly := extractWorkItemID(workItemID)

		// Build work item payload for update (with full ID including project)
		workItemPayload := i.buildWorkItemPayload(testCase, workItemID)

		// Update work item
		_, err := i.client.UpdateWorkItem(workItemIDOnly, workItemPayload)
		if err != nil {
			result.Error = fmt.Errorf("failed to update work item: %w", err)
			i.logger.Printf("✗ Failed to create test case %s: %v\n", testCase.ID, result.Error)
			return result
		}

		i.logger.Printf("✓ Successfully updated work item: %s\n", testCase.ID)
		result.WorkItemID = workItemID
		result.Action = "updated"
		workItemCreated = false
	} else {
		// Work item doesn't exist, create it
		i.logger.Printf("Work item does not exist. Creating new work item...\n")

		// Build work item payload (without test steps, no ID for creation)
		workItemPayload := i.buildWorkItemPayload(testCase, "")

		// Create work item
		response, err := i.client.CreateWorkItem(workItemPayload)
		if err != nil {
			result.Error = err
			i.logger.Printf("✗ Failed to create test case %s: %v\n", testCase.ID, result.Error)
			return result
		}

		if len(response.Data) == 0 {
			result.Error = fmt.Errorf("no work item returned in response")
			i.logger.Printf("✗ Failed to create test case %s: %v\n", testCase.ID, result.Error)
			return result
		}

		workItemID = response.Data[0].ID
		i.logger.Printf("✓ Successfully created work item: %s (ID: %s)\n", testCase.ID, workItemID)
		result.WorkItemID = workItemID
		result.Action = "created"
		workItemCreated = true
	}

	// Add test steps if present
	if len(testCase.Steps) > 0 {
		// Extract just the ID part (e.g., "TEST-123" from "PRJ/TEST-123")
		workItemIDOnly := extractWorkItemID(workItemID)

		// Check if test steps already exist
		i.logger.Printf("Checking for existing test steps...\n")
		existingSteps, err := i.client.GetTestSteps(workItemIDOnly)
		if err != nil {
			result.Error = fmt.Errorf("failed to check existing test steps: %w", err)
			i.logger.Printf("✗ Failed to create test case %s: %v\n", testCase.ID, result.Error)
			return result
		}

		// Handle existing test steps
		if existingSteps != nil && len(existingSteps.Data) > 0 {
			i.logger.Printf("Found %d existing test steps.\n", len(existingSteps.Data))

			// Check if we should skip confirmation
			shouldDelete := i.autoConfirm
			if !i.autoConfirm {
				// Ask user for confirmation before deleting
				confirmMsg := fmt.Sprintf("⚠ Existing test steps will be deleted and replaced. Continue?")
				shouldDelete = confirmAction(confirmMsg)
			} else {
				i.logger.Printf("Auto-confirm enabled - proceeding with deletion\n")
			}

			if !shouldDelete {
				// User declined - print existing steps and skip recreation
				fmt.Println("\nℹ Skipping test steps update. Showing existing test steps:")
				i.printExistingTestSteps(existingSteps)
				i.logger.Printf("✓ Work item processed (test steps unchanged)\n")
				return result
			}

			// User confirmed (or auto-confirmed) - proceed with deletion
			i.logger.Printf("Deleting existing test steps...\n")
			if err := i.client.DeleteTestSteps(workItemIDOnly, existingSteps); err != nil {
				result.Error = fmt.Errorf("failed to delete existing test steps: %w", err)
				i.logger.Printf("✗ Failed to create test case %s: %v\n", testCase.ID, result.Error)
				return result
			}
			i.logger.Printf("✓ Successfully deleted existing test steps\n")
		} else {
			i.logger.Printf("No existing test steps found\n")
		}

		// Add new test steps
		testStepsPayload := i.buildTestStepsPayload(testCase.Steps)

		if workItemCreated {
			i.logger.Printf("Adding %d test steps to newly created work item...\n", len(testCase.Steps))
		} else {
			i.logger.Printf("Adding %d test steps to existing work item...\n", len(testCase.Steps))
		}

		if err := i.client.AddTestSteps(workItemIDOnly, testStepsPayload); err != nil {
			if workItemCreated {
				result.Error = fmt.Errorf("work item created but failed to add test steps: %w", err)
			} else {
				result.Error = fmt.Errorf("failed to add test steps to existing work item: %w", err)
			}
			i.logger.Printf("✗ Failed to create test case %s: %v\n", testCase.ID, result.Error)
			return result
		}
		i.logger.Printf("✓ Successfully added %d test steps\n", len(testCase.Steps))
	}

	return result
}

// buildWorkItemPayload builds the work item creation/update payload (without test steps)
// If workItemID is provided (non-empty), it sets the ID field for PATCH operations
func (i *Importer) buildWorkItemPayload(testCase *TestCase, workItemID string) *WorkItemPayload {
	defaults := i.config.Polarion.TestCase.Defaults

	// Build description HTML
	description := i.buildDescriptionHTML(testCase)

	// Build attributes
	attributes := map[string]interface{}{
		"title": testCase.Title,
		"description": TextContent{
			Type:  "text/html",
			Value: description,
		},
	}

	// For CREATE operations, include type and caseID
	// For UPDATE operations, these are read-only
	if workItemID == "" {
		attributes["type"] = i.config.Polarion.TestCase.WorkItemType
		attributes["caseID"] = testCase.ID
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

	workItemData := WorkItemData{
		Type:       "workitems",
		Attributes: attributes,
	}

	// For PATCH operations, set the ID field
	if workItemID != "" {
		workItemData.ID = workItemID
	}

	return &WorkItemPayload{
		Data: []WorkItemData{workItemData},
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
// For example: "PRJ/TEST-123" -> "TEST-123"
func extractWorkItemID(fullID string) string {
	// If there's a slash, take the part after it
	if idx := strings.LastIndex(fullID, "/"); idx >= 0 {
		return fullID[idx+1:]
	}
	// Otherwise return the full ID as is
	return fullID
}

// confirmAction prompts the user for confirmation
func confirmAction(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/N): ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// printExistingTestSteps prints the existing test steps in a readable format
func (i *Importer) printExistingTestSteps(steps *TestStepsResponse) {
	if steps == nil || len(steps.Data) == 0 {
		fmt.Println("No test steps to display")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Existing Test Steps (%d total):\n", len(steps.Data))
	fmt.Println(strings.Repeat("=", 60))

	for idx, step := range steps.Data {
		fmt.Printf("\nStep %d:\n", idx+1)

		// Extract step description and expected result from values
		if len(step.Attributes.Values) >= 2 {
			// Values are typically [step, expectedResult]
			if stepContent, ok := step.Attributes.Values[0].(map[string]interface{}); ok {
				if value, ok := stepContent["value"].(string); ok {
					fmt.Printf("  Description: %s\n", value)
				}
			}
			if expectedContent, ok := step.Attributes.Values[1].(map[string]interface{}); ok {
				if value, ok := expectedContent["value"].(string); ok {
					fmt.Printf("  Expected:    %s\n", value)
				}
			}
		}
	}

	fmt.Println(strings.Repeat("=", 60) + "\n")
}
