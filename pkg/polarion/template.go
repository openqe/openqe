package polarion

import (
	"fmt"
	"os"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/zalando/go-keyring"
)

// TemplateRenderer handles Jinja2-style template rendering using pongo2
type TemplateRenderer struct {
	context pongo2.Context
}

// keyringFunc is a callable function for pongo2 templates that retrieves secrets from keyring
type keyringFunc struct{}

// Call implements the pongo2 callable interface for keyring function
func (k keyringFunc) Call(args ...interface{}) (interface{}, error) {
	serviceName := "polarion"
	secretName := "api_key"

	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			serviceName = s
		}
	}
	if len(args) > 1 {
		if s, ok := args[1].(string); ok {
			secretName = s
		}
	}

	return GetKeyringSecret(serviceName, secretName)
}

// NewTemplateRenderer creates a new template renderer with global environment variables
func NewTemplateRenderer() *TemplateRenderer {
	renderer := &TemplateRenderer{
		context: make(pongo2.Context),
	}

	// Add environment variables to context
	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		key, value, found := strings.Cut(env, "=")
		if found {
			envMap[key] = value
		}
	}
	renderer.context["env"] = envMap

	// Add keyring function as a callable
	renderer.context["keyring"] = keyringFunc{}

	return renderer
}

// Render renders a template string with the given context
func (r *TemplateRenderer) Render(templateStr string, params map[string]interface{}) (string, error) {
	// Merge params with global context
	ctx := pongo2.Context{}
	for k, v := range r.context {
		ctx[k] = v
	}
	for k, v := range params {
		ctx[k] = v
	}

	// Parse and execute template using pongo2
	tpl, err := pongo2.FromString(templateStr)
	if err != nil {
		// If template parsing fails, return original string
		return templateStr, nil
	}

	result, err := tpl.Execute(ctx)
	if err != nil {
		// If execution fails, return original string
		return templateStr, nil
	}

	return result, nil
}

// GetKeyringSecret retrieves a secret from the system keyring
func GetKeyringSecret(serviceName, secretName string) (string, error) {
	if serviceName == "" {
		serviceName = "polarion"
	}
	if secretName == "" {
		secretName = "api_key"
	}

	secret, err := keyring.Get(serviceName, secretName)
	if err != nil {
		return "", fmt.Errorf("failed to get secret from keyring: %w", err)
	}

	return secret, nil
}
