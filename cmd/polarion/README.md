# Polarion Test Case Import

This module provides functionality to import test cases into Polarion using the REST API v1.

## Features

- **Two-step POST API**:
  - First POST creates the work item without test steps
  - Second POST adds test steps to the created work item
- **Jinja2 Template Support**: Uses `pongo2` for template rendering with support for:
  - Global environment variables via `{{ env.VARIABLE_NAME }}`
  - Keyring integration for secrets via `{{ keyring('service', 'secret_name') }}`
- **Flexible Configuration**: YAML-based configuration with default values
- **Dry-run Mode**: Preview what would be created without making actual API calls
- **Verbose Logging**: Detailed logging for debugging

## Usage

### Basic Import

```bash
# Import using default config (config.local.yaml)
openqe polarion import

# Import using specific config file
openqe polarion import --config my_config.yaml

# Import using specific test cases file
openqe polarion import --test-cases test_cases.yaml
```

### Dry Run

```bash
# Preview what would be created without actually creating test cases
openqe polarion import --dry-run
```

### Test Connection

```bash
# Test connection to Polarion server
openqe polarion import --test-connection
```

### Verbose Mode

```bash
# Enable detailed logging
openqe polarion import --verbose
```

## Configuration

### Config File Structure

Create a `config.local.yaml` file with the following structure:

```yaml
polarion:
  # Polarion server URL (without /polarion/rest/...)
  server_url: "https://polarion.example.com"

  # Project ID in Polarion
  project_id: "YOUR_PROJECT_ID"

  # Authentication
  auth:
    # Option 1: API Token from environment variable
    api_token: "{{ env.POLARION_API_TOKEN }}"

    # Option 2: API Token from keyring
    # api_token: "{{ keyring('polarion', 'api_key') }}"

    # Option 3: Username/Password
    # username: "your_username"
    # password: "your_password"

  # Test case configuration
  test_case:
    # Work item type (usually "testcase")
    work_item_type: "testcase"

    # Required fields
    required_fields:
      - status
      - priority

    # Default attribute values
    defaults:
      status: "draft"
      priority: "high"
      automation: "automated"
      level: "component"
      component: "YourComponent"
      test_type: "functional"

# Test cases file location
test_cases_file: "test_cases.yaml"

# Options
dry_run: false
verbose: true
```

### Test Cases File Structure

Create a `test_cases.yaml` file:

```yaml
test_cases:
  - id: "TEST-001"
    title: "Test Case Title"
    description: |
      Detailed description of the test case.
      Can span multiple lines.

    priority: "high"
    category: "Category Name"
    automation_script: "test_script_name"
    script_line: 42

    prerequisites: |
      - Prerequisite 1
      - Prerequisite 2

    test_data: |
      - Data item 1
      - Data item 2

    steps:
      - step_num: 1
        description: "Step description"
        command: 'command to execute'
        expected: "Expected result"

      - step_num: 2
        description: "Next step"
        command: 'another command'
        expected: "Another expected result"
```

## Template Support

The configuration file supports Jinja2-style templates via `pongo2`:

### Environment Variables

Access environment variables:

```yaml
api_token: "{{ env.POLARION_API_TOKEN }}"
server_url: "{{ env.POLARION_SERVER }}"
```

### Keyring Integration

Retrieve secrets from system keyring:

```yaml
# Two-argument form: keyring('service_name', 'secret_name')
api_token: "{{ keyring('polarion', 'api_key') }}"

# Single-argument form (uses 'polarion' as service name):
api_token: "{{ keyring('api_key') }}"
```

To store a secret in the keyring, use your system's keyring tool or the Go keyring library.

## API Endpoints

The import process uses two REST API endpoints:

1. **Create Work Item** (without test steps):
   ```
   POST /polarion/rest/v1/projects/{project_id}/workitems
   ```

2. **Add Test Steps**:
   ```
   POST /polarion/rest/v1/projects/{project_id}/workitems/{workItemId}/teststeps
   ```

## Test Steps Payload Format

Test steps are sent in this format:

```json
{
  "data": [
    {
      "type": "teststeps",
      "attributes": {
        "keys": ["step", "expectedResult"],
        "values": [
          {
            "type": "text/html",
            "value": "<p>Step description</p><pre>command</pre>"
          },
          {
            "type": "text/html",
            "value": "<p>Expected result</p>"
          }
        ]
      }
    }
  ]
}
```

## Dependencies

- `github.com/flosch/pongo2/v6` - Jinja2-compatible template engine
- `github.com/zalando/go-keyring` - Cross-platform keyring access
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/spf13/cobra` - CLI framework

## Examples

See the example files in the repository:
- `example-polarion-config.yaml` - Example configuration file
- `example-test-cases.yaml` - Example test cases file

## Implementation Details

### Package Structure

```
cmd/polarion/
├── polarion.go       # Main command
└── import.go         # Import subcommand

pkg/polarion/
├── types.go          # Data structures
├── config.go         # Configuration loading
├── template.go       # Jinja2 template support
├── client.go         # Polarion API client
└── importer.go       # Import logic
```

### Key Features

1. **Template Rendering**: Pre-processes configuration files to resolve template expressions before parsing YAML
2. **HTML Escaping**: Properly escapes HTML special characters in test case data
3. **Error Handling**: Graceful error handling with meaningful error messages
4. **Verbose Logging**: Optional detailed logging for debugging

## Comparison with Python Implementation

This Go implementation provides the same functionality as the Python version with some improvements:

- ✅ Two-step POST API (work item creation + test steps addition)
- ✅ Jinja2 template support via pongo2
- ✅ Keyring integration for secrets
- ✅ Environment variable access in templates
- ✅ Dry-run mode
- ✅ Connection testing
- ✅ Verbose logging
- ✅ HTML escaping
- ✅ Default attribute values
- ✅ Required field validation

Key differences:
- Uses `pongo2` instead of `jinja2` (Go vs Python)
- Uses `go-keyring` instead of Python `keyring`
- Compiled binary (no Python runtime required)
- Integrated into the `openqe` CLI tool
