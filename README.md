# Codacy CLI v2

The Codacy CLI (version 2) is a command-line tool that helps you analyze code quality and security issues in your codebase and upload the results to Codacy. It supports multiple analysis tools including ESLint, Trivy, PMD, PyLint, and DartAnalyzer.

## Installation

### macOS (Homebrew)

```bash
brew install codacy/codacy-cli-v2/codacy-cli-v2
```

### Linux and macOS (Script)

```bash
bash <(curl -Ls https://raw.githubusercontent.com/codacy/codacy-cli-v2/main/codacy-cli.sh)
```

You can create an alias for easy access:

```bash
alias codacy-cli="bash <(curl -Ls https://raw.githubusercontent.com/codacy/codacy-cli-v2/main/codacy-cli.sh)"
```

### Specific Version

To use a specific version of the CLI:

```bash
export CODACY_CLI_V2_VERSION="1.0.0-main.133.3607792"
```

## Getting Started

### Initialize Your Project

Set up your project with a configuration file:

```bash
# With Codacy API token (recommended)
codacy-cli init --api-token YOUR_API_TOKEN --provider gh --organization YOUR_ORG --repository YOUR_REPO

# Or use local configuration only
codacy-cli init
```

### Install Required Tools

Install the runtimes and tools specified in your configuration:

```bash
codacy-cli install
```

## Commands

### Initialize Project (`init`)

```bash
# With API token
codacy-cli init --api-token YOUR_API_TOKEN --provider gh --organization YOUR_ORG --repository YOUR_REPO

# Local mode (without Codacy integration)
codacy-cli init
```

This command creates a `.codacy` directory with configuration files, including `codacy.yaml` that defines the tools and runtime versions to use.

### Install Tools (`install`)

```bash
codacy-cli install
```

Installs all the runtimes and tools specified in your configuration file.

### Run Analysis (`analyze`)

```bash
# Run analysis with a specific tool
codacy-cli analyze --tool eslint

# Run analysis and output results in SARIF format to terminal
codacy-cli analyze --tool eslint --format sarif

# Run analysis and save results to a file
codacy-cli analyze --tool eslint --format sarif --output results.sarif

# Run analysis with auto-fix (when supported)
codacy-cli analyze --tool eslint --fix
```

Supported tools:
- ESLint (JavaScript/TypeScript)
- Trivy (Security scanning)
- PMD (Java, etc.)
- PyLint (Python)
- DartAnalyzer (Dart)

### Upload Results (`upload`)

Using project token:
```bash
codacy-cli upload --sarif-path results.sarif --commit-uuid YOUR_COMMIT_UUID --project-token YOUR_PROJECT_TOKEN
```

Using API token:
```bash
codacy-cli upload --sarif-path results.sarif --commit-uuid YOUR_COMMIT_UUID --api-token YOUR_API_TOKEN --provider gh --owner YOUR_ORG --repository YOUR_REPO
```

## Configuration File

The Codacy CLI uses a YAML configuration file (`.codacy/codacy.yaml`) to specify runtimes and tools:

```yaml
runtimes:
  - node@22.2.0
  - python@3.11.11
  - dart@3.7.2
tools:
  - eslint@9.3.0
  - trivy@0.59.1
  - pylint@3.3.6
  - pmd@6.55.0
  - dartanalyzer@3.7.2
```

## CI/CD Integration

To use Codacy CLI in your CI/CD pipeline, add it to your workflow. Example for GitHub Actions:

```yaml
- name: Install Codacy CLI
  run: bash <(curl -Ls https://raw.githubusercontent.com/codacy/codacy-cli-v2/main/codacy-cli.sh) download

- name: Initialize Codacy CLI
  run: codacy-cli init

- name: Install tools
  run: codacy-cli install

- name: Run analysis
  run: codacy-cli analyze --tool eslint --format sarif --output results.sarif

- name: Upload results
  run: codacy-cli upload --sarif-path results.sarif --commit-uuid ${{ github.sha }} --project-token ${{ secrets.CODACY_PROJECT_TOKEN }}
```

## Example Repository

For a complete example of using the Codacy CLI, check out: https://github.com/troubleshoot-codacy/eslint-test-examples

## Troubleshooting

### Docker Credential Helper Errors

If you encounter errors related to `docker-credential-osxkeychain` when running analysis, install the Docker credential helper:

```bash
# macOS
brew install docker-credential-helper
```

### Configuration File Not Found

If you see an error about missing configuration:

```
No configuration file was found, execute init command first.
```

Run `codacy-cli init` to create the necessary configuration files.

### Tool Installation Issues

If you encounter issues with tool installation, try:

1. Check your internet connection
2. Verify you have appropriate permissions
3. Run `codacy-cli install` with verbose output to see detailed errors

## Support

For issues or questions, please file an issue on the [GitHub repository](https://github.com/codacy/codacy-cli-v2/issues).
