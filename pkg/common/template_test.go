package common

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
	// To test manually with keyring setup, use secret-tool:
	// secret-tool store --label='Test API Key' service myservice username api_key

	// Test that the template parses correctly (execution may fail without keyring)
	template := "{{ ''|keyring:'myservice,api_key' }}"
	_, err := renderer.Render(template, nil)

	// We expect either success (if keyring is set up) or a specific keyring error
	// The important thing is that parsing should not fail
	if err != nil {
		// Should be a keyring error, not a template parsing error
		assert.NotContains(t, err.Error(), "failed to parse template")
	}
}

func TestTemplateRenderer_MultipleEnvVars(t *testing.T) {
	// Set multiple test environment variables
	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	defer func() {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")
	}()

	renderer := NewTemplateRenderer()

	// Test multiple environment variable expansion
	result, err := renderer.Render("{{ env.VAR1 }}-{{ env.VAR2 }}", nil)
	assert.NoError(t, err)
	assert.Equal(t, "value1-value2", result)
}

func TestTemplateRenderer_MixedContext(t *testing.T) {
	// Set environment variable
	os.Setenv("ENV_VAR", "from_env")
	defer os.Unsetenv("ENV_VAR")

	renderer := NewTemplateRenderer()

	params := map[string]interface{}{
		"param_var": "from_param",
	}

	// Test mixing env variables and params
	result, err := renderer.Render("{{ env.ENV_VAR }}-{{ param_var }}", params)
	assert.NoError(t, err)
	assert.Equal(t, "from_env-from_param", result)
}
