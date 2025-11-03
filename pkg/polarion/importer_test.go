package polarion

import (
	"log"
	"os"
	"testing"
)

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple snake_case",
			input:    "case_automation",
			expected: "caseAutomation",
		},
		{
			name:     "three word snake_case",
			input:    "case_importance_level",
			expected: "caseImportanceLevel",
		},
		{
			name:     "snake_case with number",
			input:    "sub_type_1",
			expected: "subType1",
		},
		{
			name:     "single word no underscore",
			input:    "priority",
			expected: "priority",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already camelCase",
			input:    "caseAutomation",
			expected: "caseAutomation",
		},
		{
			name:     "two underscores in a row",
			input:    "test__case",
			expected: "testCase",
		},
		{
			name:     "starting with underscore",
			input:    "_test_case",
			expected: "TestCase",
		},
		{
			name:     "ending with underscore",
			input:    "test_case_",
			expected: "testCase",
		},
		{
			name:     "all lowercase multiple words",
			input:    "test_type",
			expected: "testType",
		},
		{
			name:     "long snake_case",
			input:    "my_custom_polarion_field_name",
			expected: "myCustomPolarionFieldName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := snakeToCamel(tt.input)
			if result != tt.expected {
				t.Errorf("snakeToCamel(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildWorkItemPayload(t *testing.T) {
	tests := []struct {
		name            string
		testCase        *TestCase
		defaults        map[string]interface{}
		workItemType    string
		expectedAttrs   map[string]interface{}
		unexpectedAttrs []string
	}{
		{
			name: "basic test case with no defaults",
			testCase: &TestCase{
				ID:          "TEST-001",
				Title:       "Test Login",
				Description: "Test user login functionality",
				Priority:    "high",
				Category:    "Authentication",
			},
			defaults:     map[string]interface{}{},
			workItemType: "testcase",
			expectedAttrs: map[string]interface{}{
				"type":     "testcase",
				"title":    "Test Login",
				"caseID":   "TEST-001",
				"priority": "high",
				"category": "Authentication",
			},
			unexpectedAttrs: []string{},
		},
		{
			name: "test case with all fields",
			testCase: &TestCase{
				ID:          "TEST-002",
				Title:       "Test Logout",
				Description: "Test user logout",
				Priority:    "medium",
				Category:    "Authentication",
				Type:        "functional",
				Component:   "Login Module",
				Status:      "approved",
				Level:       "component",
				TestType:    "automated",
			},
			defaults:     map[string]interface{}{},
			workItemType: "testcase",
			expectedAttrs: map[string]interface{}{
				"type":      "functional",
				"title":     "Test Logout",
				"caseID":    "TEST-002",
				"priority":  "medium",
				"category":  "Authentication",
				"component": "Login Module",
				"status":    "approved",
				"level":     "component",
				"testType":  "automated",
			},
			unexpectedAttrs: []string{},
		},
		{
			name: "test case with custom attributes (snake_case)",
			testCase: &TestCase{
				ID:          "TEST-003",
				Title:       "Test API",
				Description: "Test API endpoints",
				Attributes: map[string]string{
					"case_automation": "automated",
					"case_importance": "critical",
					"sub_type_1":      "compliance",
				},
			},
			defaults:     map[string]interface{}{},
			workItemType: "testcase",
			expectedAttrs: map[string]interface{}{
				"type":           "testcase",
				"title":          "Test API",
				"caseID":         "TEST-003",
				"caseAutomation": "automated",
				"caseImportance": "critical",
				"subType1":       "compliance",
			},
			unexpectedAttrs: []string{},
		},
		{
			name: "defaults applied when fields empty",
			testCase: &TestCase{
				ID:          "TEST-004",
				Title:       "Test with defaults",
				Description: "Test default values",
			},
			defaults: map[string]interface{}{
				"priority":        "low",
				"case_automation": "manual",
				"assignee":        "team@example.com",
			},
			workItemType: "testcase",
			expectedAttrs: map[string]interface{}{
				"type":           "testcase",
				"title":          "Test with defaults",
				"caseID":         "TEST-004",
				"priority":       "low",
				"caseAutomation": "manual",
				"assignee":       "team@example.com",
			},
			unexpectedAttrs: []string{},
		},
		{
			name: "test case fields override defaults",
			testCase: &TestCase{
				ID:          "TEST-005",
				Title:       "Test override",
				Description: "Test override defaults",
				Priority:    "critical",
				Attributes: map[string]string{
					"case_automation": "fully-automated",
				},
			},
			defaults: map[string]interface{}{
				"priority":        "low",
				"case_automation": "manual",
				"assignee":        "default@example.com",
			},
			workItemType: "testcase",
			expectedAttrs: map[string]interface{}{
				"type":           "testcase",
				"title":          "Test override",
				"caseID":         "TEST-005",
				"priority":       "critical",            // overridden
				"caseAutomation": "fully-automated",     // overridden by custom attribute
				"assignee":       "default@example.com", // from defaults
			},
			unexpectedAttrs: []string{},
		},
		{
			name: "defaults with snake_case conversion",
			testCase: &TestCase{
				ID:          "TEST-006",
				Title:       "Test snake case",
				Description: "Test snake_case in defaults",
			},
			defaults: map[string]interface{}{
				"custom_field_one":     "value1",
				"custom_field_two":     "value2",
				"very_long_field_name": "value3",
			},
			workItemType: "testcase",
			expectedAttrs: map[string]interface{}{
				"type":              "testcase",
				"title":             "Test snake case",
				"caseID":            "TEST-006",
				"customFieldOne":    "value1",
				"customFieldTwo":    "value2",
				"veryLongFieldName": "value3",
			},
			unexpectedAttrs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create importer with test config
			importer := &Importer{
				config: &Config{
					Polarion: PolarionConfig{
						TestCase: TestCaseConfig{
							WorkItemType: tt.workItemType,
							Defaults:     tt.defaults,
						},
					},
				},
				logger: log.New(os.Stderr, "[TEST] ", 0),
			}

			// Build payload
			payload := importer.buildWorkItemPayload(tt.testCase, "")

			// Verify payload structure
			if payload == nil {
				t.Fatal("payload is nil")
			}

			if len(payload.Data) != 1 {
				t.Fatalf("expected 1 work item, got %d", len(payload.Data))
			}

			workItem := payload.Data[0]
			if workItem.Type != "workitems" {
				t.Errorf("expected type 'workitems', got %q", workItem.Type)
			}

			// Check expected attributes
			for key, expectedValue := range tt.expectedAttrs {
				actualValue, exists := workItem.Attributes[key]
				if !exists {
					t.Errorf("expected attribute %q to exist", key)
					continue
				}

				// Special handling for TextContent type
				if key == "description" {
					if tc, ok := actualValue.(TextContent); ok {
						// Just verify it's a TextContent, detailed content check not needed
						if tc.Type != "text/html" {
							t.Errorf("description type: expected 'text/html', got %q", tc.Type)
						}
					} else {
						t.Errorf("description should be TextContent type")
					}
					continue
				}

				// For other attributes, compare values
				if actualValue != expectedValue {
					t.Errorf("attribute %q: expected %v, got %v", key, expectedValue, actualValue)
				}
			}

			// Check unexpected attributes don't exist
			for _, key := range tt.unexpectedAttrs {
				if _, exists := workItem.Attributes[key]; exists {
					t.Errorf("attribute %q should not exist", key)
				}
			}

			// Verify description always exists and is TextContent type
			desc, exists := workItem.Attributes["description"]
			if !exists {
				t.Error("description attribute must exist")
			} else if _, ok := desc.(TextContent); !ok {
				t.Error("description must be TextContent type")
			}
		})
	}
}

func TestExtractWorkItemID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Full ID with project prefix",
			input:    "OSE/OCP-85835",
			expected: "OCP-85835",
		},
		{
			name:     "Full ID with different project",
			input:    "PROJECT/ITEM-123",
			expected: "ITEM-123",
		},
		{
			name:     "ID without prefix",
			input:    "OCP-85835",
			expected: "OCP-85835",
		},
		{
			name:     "ID with multiple slashes (take last)",
			input:    "PROJECT/SPACE/OCP-85835",
			expected: "OCP-85835",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single slash at end",
			input:    "PROJECT/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWorkItemID(tt.input)
			if result != tt.expected {
				t.Errorf("extractWorkItemID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
