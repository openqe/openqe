# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OpenQE is a command-line utility suite for QE (Quality Engineering) workflows, built with Go and Cobra. It provides utilities for test case management, OpenShift operations, TLS certificate generation, and authentication helpers.

## Build and Test Commands

```bash
# Build the project (includes fmt and vet)
make build

# Run all tests
make test

# Run tests for a specific package
go test ./pkg/polarion -v

# Run a specific test
go test ./pkg/polarion -v -run TestExtractWorkItemID

# Format code
make fmt

# Run static analysis
make vet

# Run the binary
./bin/openqe --help
```

## Architecture

### Command Structure (Cobra CLI)

The project follows a standard Cobra CLI pattern:

- **main.go**: Entry point that initializes the root Cobra command and adds subcommands
- **cmd/**: Cobra command definitions organized by domain
  - `auth/`: Authentication utilities (htpasswd generation)
  - `core/`: Core utilities (TLS cert generation, documentation)
  - `openshift/`: OpenShift-specific utilities (image registry, pull secrets)
  - `polarion/`: Polarion test case import/management
- **pkg/**: Business logic separated from CLI concerns
  - Matches cmd/ structure: `pkg/auth/`, `pkg/polarion/`, etc.

### Polarion Integration Architecture

The Polarion integration is the most complex subsystem with these key components:

**Data Flow:**
1. **Config Loading** (`config.go`): Loads YAML config with Jinja2-style templating support
2. **Template Rendering** (`template.go`): Processes templates using pongo2 (Django/Jinja2-like engine)
   - Supports `{{ env.VAR }}` for environment variables
   - Supports `{{ ''|keyring:'service,key' }}` filter for system keyring integration
3. **Test Case Import** (`importer.go`): Orchestrates the import process
4. **API Client** (`client.go`): REST API communication with Polarion
5. **Type Definitions** (`types.go`): Data structures for API payloads and configuration

**Key Design Patterns:**

1. **Idempotent Imports**: The importer checks if work items exist before creating them
   - Uses `GetWorkItem()` to check existence by ID
   - If exists: skips creation, only updates test steps
   - If not exists: creates work item, then adds test steps
   - This allows safe re-runs when test step creation fails

2. **Two-Phase Work Item Creation**:
   - Phase 1: Create work item WITHOUT test steps
   - Phase 2: Add test steps separately via `AddTestSteps()`
   - Work item ID extraction: Polarion returns `"PROJECT/ID"` format (e.g., `"PRJ/TEST-123"`), but test steps API expects just `"TEST-123"` - handled by `extractWorkItemID()`

3. **Snake to Camel Case Conversion**:
   - YAML configs use `snake_case` (e.g., `case_automation`)
   - Polarion API expects `camelCase` (e.g., `caseAutomation`)
   - Automatic conversion via `snakeToCamel()` function

4. **Template Rendering for Secrets**:
   - Config files support Jinja2-style templating
   - `{{ env.VAR }}`: Environment variable expansion
   - `{{ ''|keyring:'service,key' }}`: System keyring integration (using pongo2 filter syntax)
   - Registered globally in `init()` function

### Important Implementation Details

**Polarion API Integration:**
- Base URL construction: `{server_url}/polarion/rest/v1`
- Authentication: Supports both API token (Bearer) and username/password (Basic Auth)
- Work item creation returns full ID: `"PROJECT/WORKITEM-ID"`
- Test steps API expects only: `"WORKITEM-ID"` (without project prefix)

**VSCode Debug Configuration:**
- Launch configs are in `.vscode/launch.json`
- Pre-configured for debugging polarion import with dry-run mode
- Uses `config.local.yaml` for local development

**Testing Patterns:**
- Use table-driven tests (see `importer_test.go` for examples)
- Mock HTTP clients are NOT used; tests focus on data transformation logic
- Test helpers use `log.New(os.Stderr, "[TEST] ", 0)` for verbose output

## Key Workflow: Polarion Test Case Import

When implementing or debugging Polarion import:

1. **Check existence first**: Always call `GetWorkItem()` before creating
2. **Extract ID correctly**: Use `extractWorkItemID()` to get just the ID part for test steps
3. **Handle both creation paths**: New work items vs. existing work items
4. **Template rendering**: Errors are now properly propagated (not silently swallowed)

Example flow:
```
User runs: openqe polarion import --config config.yaml --test-cases tests.yaml

1. Load config (with template rendering for secrets)
2. Load test cases from YAML
3. For each test case:
   a. Check if work item exists (GetWorkItem)
   b. If not exists: CreateWorkItem
   c. Extract work item ID (remove project prefix)
   d. AddTestSteps (works for both new and existing)
```

## Configuration Files

**Development Config:**
- Use `config.local.yaml` (gitignored) for local Polarion credentials
- Use environment variables: `export POLARION_API_TOKEN="..."`
- Or use system keyring: `secret-tool store --label='Polarion API Key' service polarion username api_key`

**Test Data:**
- `example-polarion-config.yaml`: Reference configuration
- `example-test-cases.yaml`: Reference test case format
- `test_cases.yaml`: Working test cases file (gitignored)

## Common Patterns

**Adding New Utilities:**
- Follow the established pattern: Create both `cmd/{domain}/` and `pkg/{domain}/`
- Use Cobra's `NewCommand()` pattern for CLI commands
- Separate business logic into pkg/ for testability
- Add unit tests in `{package}_test.go`

**Working with Templates:**
- pongo2 uses filter syntax, NOT function call syntax
- Correct: `{{ ''|keyring:'service,key' }}`
- Wrong: `{{ keyring('service', 'key') }}` (Jinja2 style, not supported)

**Error Handling:**
- Return errors with context using `fmt.Errorf("failed to X: %w", err)`
- Use logger with verbosity control: Check `config.Verbose` or command flags
- Distinguish user errors from system errors in CLI commands
