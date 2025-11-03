package polarion

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateRenderer_RenderEnv(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	renderer := NewTemplateRenderer()

	// Test environment variable expansion
	result, err := renderer.Render("{{ env.TEST_VAR }}", nil)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", result)
}

func TestTemplateRenderer_RenderPlainText(t *testing.T) {
	renderer := NewTemplateRenderer()

	// Test plain text (no template)
	result, err := renderer.Render("plain text", nil)
	assert.NoError(t, err)
	assert.Equal(t, "plain text", result)
}

func TestTemplateRenderer_InvalidTemplate(t *testing.T) {
	renderer := NewTemplateRenderer()

	// Test invalid template syntax
	_, err := renderer.Render("{{ invalid syntax", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse template")
}

func TestTemplateRenderer_RenderWithParams(t *testing.T) {
	renderer := NewTemplateRenderer()

	params := map[string]interface{}{
		"custom_var": "custom_value",
	}

	result, err := renderer.Render("{{ custom_var }}", params)
	assert.NoError(t, err)
	assert.Equal(t, "custom_value", result)
}

func TestTemplateRenderer_KeyringFilterSyntax(t *testing.T) {
	renderer := NewTemplateRenderer()

	// This test demonstrates the correct syntax but will fail without keyring setup
	// To test manually with keyring setup, uncomment and set up keyring:
	// secret-tool store --label='Polarion API Key' service polarion username api_key
	// Then enter your token when prompted

	// Test that the template parses correctly (execution may fail without keyring)
	template := "{{ ''|keyring:'polarion,api_key' }}"
	_, err := renderer.Render(template, nil)

	// We expect either success (if keyring is set up) or a specific keyring error
	// The important thing is that parsing should not fail
	if err != nil {
		// Should be a keyring error, not a template parsing error
		assert.NotContains(t, err.Error(), "failed to parse template")
	}
}
