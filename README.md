# OpenQE

Aims to provide utilities to speed up preparation for QE testing.

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/openqe/openqe/releases).

#### Linux (x86_64)
```bash
# Download the latest release
curl -LO https://github.com/openqe/openqe/releases/latest/download/openqe_<VERSION>_Linux_x86_64.tar.gz

# Extract the binary
tar -xzf openqe_<VERSION>_Linux_x86_64.tar.gz

# Move to a directory in your PATH
sudo mv openqe /usr/local/bin/

# Verify installation
openqe --help
```

#### macOS (x86_64)
```bash
# Download the latest release
curl -LO https://github.com/openqe/openqe/releases/latest/download/openqe_<VERSION>_Darwin_x86_64.tar.gz

# Extract the binary
tar -xzf openqe_<VERSION>_Darwin_x86_64.tar.gz

# Move to a directory in your PATH
sudo mv openqe /usr/local/bin/

# Verify installation
openqe --help
```

#### macOS (ARM64/Apple Silicon)
```bash
# Download the latest release
curl -LO https://github.com/openqe/openqe/releases/latest/download/openqe_<VERSION>_Darwin_arm64.tar.gz

# Extract the binary
tar -xzf openqe_<VERSION>_Darwin_arm64.tar.gz

# Move to a directory in your PATH
sudo mv openqe /usr/local/bin/

# Verify installation
openqe --help
```

#### Windows
1. Download the latest `openqe_<VERSION>_Windows_x86_64.zip` from the [Releases page](https://github.com/openqe/openqe/releases)
2. Extract the ZIP file
3. Add the directory containing `openqe.exe` to your PATH
4. Open a new command prompt and run `openqe --help`

### Build from Source

Prerequisites:
- Go 1.21 or later

```bash
git clone https://github.com/openqe/openqe.git
cd openqe
make build

# The binary will be in bin/openqe
./bin/openqe --help
```

## Usage

OpenQE provides various utilities for QE testing workflows. The main commands are organized into categories:

### Available Commands

Run `openqe --help` to see all available commands:

```bash
openqe --help
```

### Polarion Integration

Import test cases to Polarion:

```bash
# Import test cases from a YAML file
openqe polarion import --config polarion-config.yaml --test-cases test-cases.yaml

# Dry run to preview what would be imported
openqe polarion import --config polarion-config.yaml --test-cases test-cases.yaml --dry-run
```

#### Example Configuration Files

**polarion-config.yaml:**
```yaml
polarion:
  server_url: "https://polarion.example.com"
  project_id: "MyProject"
  auth:
    api_token: "{{ keyring('polarion', 'api_key') }}"
  test_case:
    work_item_type: "testcase"
    defaults:
      priority: "medium"
      case_automation: "notautomated"

test_cases_file: "test-cases.yaml"
dry_run: false
verbose: true
```

**test-cases.yaml:**
```yaml
test_cases:
  - id: "TEST-001"
    title: "Example Test Case"
    description: "Test description here"
    priority: "high"
    category: "Functional"
    type: "functional"
    component: "Auth Module"
    status: "approved"
    level: "component"
    test_type: "automated"

    # Optional custom attributes (snake_case will be converted to camelCase)
    attributes:
      case_automation: "automated"
      case_importance: "critical"

    steps:
      - description: "Step 1"
        expected: "Expected result 1"
      - description: "Step 2"
        expected: "Expected result 2"
```

See `example-polarion-config.yaml` and `example-test-cases.yaml` for more examples.

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

### Code Formatting

```bash
make fmt
```

### Linting

```bash
make vet
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See LICENSE file for details
