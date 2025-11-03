package polarion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// Client represents a Polarion API client
type Client struct {
	config     *Config
	baseURL    string
	httpClient *http.Client
	logger     *log.Logger
}

// NewClient creates a new Polarion API client
func NewClient(config *Config, verbose bool) (*Client, error) {
	client := &Client{
		config:     config,
		baseURL:    buildBaseURL(config.Polarion.ServerURL),
		httpClient: &http.Client{},
	}

	// Setup logger
	if verbose {
		client.logger = log.New(log.Writer(), "[POLARION] ", log.LstdFlags)
	} else {
		client.logger = log.New(io.Discard, "", 0)
	}

	return client, nil
}

// buildBaseURL constructs the Polarion REST API base URL
func buildBaseURL(serverURL string) string {
	serverURL = strings.TrimSuffix(serverURL, "/")
	return fmt.Sprintf("%s/polarion/rest/v1", serverURL)
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, url string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)

		// Log request payload in verbose mode
		c.logger.Printf("Request Payload: %s", string(jsonData))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set authentication
	if c.config.Polarion.Auth.APIToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.Polarion.Auth.APIToken))
	} else if c.config.Polarion.Auth.Username != "" {
		req.SetBasicAuth(c.config.Polarion.Auth.Username, c.config.Polarion.Auth.Password)
	}

	c.logger.Printf("%s %s", method, url)

	return c.httpClient.Do(req)
}

// TestConnection tests the connection to Polarion
func (c *Client) TestConnection() error {
	url := fmt.Sprintf("%s/projects/%s", c.baseURL, c.config.Polarion.ProjectID)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ Successfully connected to Polarion project: %s\n", c.config.Polarion.ProjectID)
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("connection test failed with status %d: %s", resp.StatusCode, string(bodyBytes))
}

// GetWorkItem retrieves a work item from Polarion by ID
func (c *Client) GetWorkItem(workItemID string) (*WorkItemResponseData, error) {
	url := fmt.Sprintf("%s/projects/%s/workitems/%s", c.baseURL, c.config.Polarion.ProjectID, workItemID)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get work item: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// If not found, return nil without error (so caller can handle it)
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get work item (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var responseContent map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &responseContent); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	data, _ := json.Marshal(responseContent["data"])
	var result WorkItemResponseData
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse data from response: %w", err)
	}

	return &result, nil
}

// CreateWorkItem creates a work item in Polarion (without test steps)
func (c *Client) CreateWorkItem(payload *WorkItemPayload) (*WorkItemResponse, error) {
	url := fmt.Sprintf("%s/projects/%s/workitems", c.baseURL, c.config.Polarion.ProjectID)

	resp, err := c.doRequest("POST", url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create work item: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create work item (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result WorkItemResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetTestSteps retrieves test steps for a work item
func (c *Client) GetTestSteps(workItemID string) (*TestStepsResponse, error) {
	url := fmt.Sprintf("%s/projects/%s/workitems/%s/teststeps",
		c.baseURL, c.config.Polarion.ProjectID, workItemID)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get test steps: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// If not found, return empty response
	if resp.StatusCode == http.StatusNotFound {
		return &TestStepsResponse{Data: []TestStepData{}}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get test steps (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result TestStepsResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// DeleteTestSteps deletes test steps from a work item
// If existingSteps is provided, it will delete those specific steps
// If existingSteps is nil or empty, it will try to delete without a body (may not work per API spec)
func (c *Client) DeleteTestSteps(workItemID string, existingSteps *TestStepsResponse) error {
	url := fmt.Sprintf("%s/projects/%s/workitems/%s/teststeps",
		c.baseURL, c.config.Polarion.ProjectID, workItemID)

	var payload *TestStepsDeletePayload

	// Build delete payload from existing steps if provided
	if existingSteps != nil && len(existingSteps.Data) > 0 {
		deleteData := make([]TestStepDeleteData, 0, len(existingSteps.Data))
		for _, step := range existingSteps.Data {
			if step.ID != "" {
				deleteData = append(deleteData, TestStepDeleteData{
					Type: "teststeps",
					ID:   step.ID,
				})
			}
		}

		if len(deleteData) > 0 {
			payload = &TestStepsDeletePayload{
				Data: deleteData,
			}
			c.logger.Printf("Deleting %d test steps with IDs\n", len(deleteData))
		}
	}

	resp, err := c.doRequest("DELETE", url, payload)
	if err != nil {
		return fmt.Errorf("failed to delete test steps: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Accept both OK and No Content as success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete test steps (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// AddTestSteps adds test steps to an existing work item
func (c *Client) AddTestSteps(workItemID string, payload *TestStepsPayload) error {
	url := fmt.Sprintf("%s/projects/%s/workitems/%s/teststeps",
		c.baseURL, c.config.Polarion.ProjectID, workItemID)

	resp, err := c.doRequest("POST", url, payload)
	if err != nil {
		return fmt.Errorf("failed to add test steps: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to add test steps (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
