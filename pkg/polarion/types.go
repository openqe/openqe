package polarion

// Config represents the Polarion configuration
type Config struct {
	Polarion      PolarionConfig `yaml:"polarion"`
	TestCasesFile string         `yaml:"test_cases_file"`
	DryRun        bool           `yaml:"dry_run"`
	Verbose       bool           `yaml:"verbose"`
}

// PolarionConfig contains Polarion server and project configuration
type PolarionConfig struct {
	ServerURL string         `yaml:"server_url"`
	ProjectID string         `yaml:"project_id"`
	Auth      AuthConfig     `yaml:"auth"`
	TestCase  TestCaseConfig `yaml:"test_case"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	APIToken string `yaml:"api_token"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// TestCaseConfig contains test case configuration
type TestCaseConfig struct {
	WorkItemType string                 `yaml:"work_item_type"`
	Defaults     map[string]interface{} `yaml:"defaults"`
}

// TestCasesData represents the structure of the test cases YAML file
type TestCasesData struct {
	TestCases []TestCase `yaml:"test_cases"`
}

// TestCase represents a single test case
type TestCase struct {
	ID          string            `yaml:"id"`
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Priority    string            `yaml:"priority"`
	Category    string            `yaml:"category"`
	Type        string            `yaml:"type"`
	Component   string            `yaml:"component"`
	Status      string            `yaml:"status"`
	Level       string            `yaml:"level"`
	TestType    string            `yaml:"test_type"`
	Steps       []TestStep        `yaml:"steps"`
	Attributes  map[string]string `yaml:"attributes,omitempty"`
}

// TestStep represents a single test step
type TestStep struct {
	Description string `yaml:"description"`
	Expected    string `yaml:"expected"`
}

// WorkItemPayload represents the Polarion work item creation payload
type WorkItemPayload struct {
	Data []WorkItemData `json:"data"`
}

// WorkItemData represents a single work item in the payload
type WorkItemData struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id,omitempty"` // Required for PATCH, omitted for POST
	Attributes map[string]interface{} `json:"attributes"`
}

// TextContent represents text content with type
type TextContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// TestStepsPayload represents the test steps creation payload
type TestStepsPayload struct {
	Data []TestStepData `json:"data"`
}

// TestStepsResponse represents the response from getting test steps
type TestStepsResponse struct {
	Data []TestStepData `json:"data"`
}

// TestStepsDeletePayload represents the payload for deleting test steps
type TestStepsDeletePayload struct {
	Data []TestStepDeleteData `json:"data"`
}

// TestStepDeleteData represents a single test step to delete
type TestStepDeleteData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// TestStepData represents a single test step in the payload
type TestStepData struct {
	Type       string             `json:"type"`
	ID         string             `json:"id,omitempty"`
	Attributes TestStepAttributes `json:"attributes,omitempty"`
}

// TestStepAttributes represents test step attributes
type TestStepAttributes struct {
	Keys   []string      `json:"keys"`
	Values []interface{} `json:"values"`
}

// WorkItemResponse represents the response from work item creation
type WorkItemResponse struct {
	Data []WorkItemResponseData `json:"data"`
}

// WorkItemResponseData represents a single work item in the response
type WorkItemResponseData struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
}
