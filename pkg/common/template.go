package common

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

func init() {
	// Register keyring as a global filter in pongo2
	pongo2.RegisterFilter("keyring", keyringFilter)
}

// keyringFilter is a pongo2 filter that retrieves secrets from keyring
// Usage: {{ ”|keyring:'service,secret' }}
// If not provided, defaults to 'default,api_key'
func keyringFilter(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	serviceName := "default"
	secretName := "api_key"

	// Get the parameters - pongo2 filters can receive parameters as comma-separated values
	if param.IsString() && param.String() != "" {
		// Parse parameters if provided as "service,secret"
		parts := strings.Split(param.String(), ",")
		if len(parts) > 0 && parts[0] != "" {
			serviceName = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 && parts[1] != "" {
			secretName = strings.TrimSpace(parts[1])
		}
	}

	secret, err := GetKeyringSecret(serviceName, secretName)
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:keyring",
			OrigError: err,
		}
	}

	return pongo2.AsValue(secret), nil
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
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	result, err := tpl.Execute(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result, nil
}

// GetKeyringSecret retrieves a secret from the system keyring
func GetKeyringSecret(serviceName, secretName string) (string, error) {
	if serviceName == "" {
		serviceName = "default"
	}
	if secretName == "" {
		secretName = "api_key"
	}

	secret, err := keyring.Get(serviceName, secretName)
	if err != nil {
		return "", fmt.Errorf("failed to get secret from keyring (service=%s, key=%s): %w", serviceName, secretName, err)
	}

	return secret, nil
}
