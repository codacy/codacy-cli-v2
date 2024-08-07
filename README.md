# codacy-cli-v2

This is a POC for what could be a new CLI for us. The idea is to rely on the native tools and SARIF format instead of relying on Docker.

## Overview

The `codacy-cli-v2` is a command-line tool for Codacy that supports analyzing code using ESLint and uploading the results in SARIF format to Codacy. It provides two main commands: `analyze` and `upload`.

### Commands

- **`analyze` Command**: Runs ESLint analysis on the codebase.
    - `--output, -o`: Output file for the results.
    - `--tool, -t`: Specifies the tool to run analysis with (e.g., ESLint).
    - `--fix, -f`: Automatically fixes issues when possible.
    - `--new-pr`: Creates a new GitHub PR with fixed issues.

- **`upload` Command**: Uploads a SARIF file containing analysis results to Codacy.
    - `--sarif-path, -s`: Path to the SARIF report.
    - `--commit-uuid, -c`: Commit UUID.
    - `--project-token, -t`: Project token for Codacy API.

### Important Concepts

- **`.codacy/.codacy.yaml`**: Configuration file to specify `node` and `eslint` versions for the CLI.
  ```yaml
  runtimes:
      - node@22.2.0
  tools:
      - eslint@9.3.0
  
- **`codacy-cli-v2 install`**: Command to install the specified node and eslint versions before running analysis.
- 
## Download

### MacOS (brew)

To install `codacy-cli-v2` using Homebrew:

```bash
brew install codacy/codacy-cli-v2/codacy-cli-v2
```

## Linux

For Linux, we rely on the codacy-cli.sh script in the root. To download the CLI, run:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/codacy/codacy-cli-v2/main/codacy-cli.sh)
```
You can either put the downloaded script in a specific file or create an alias that will download the script and look for changes:

```bash
alias codacy-cli-v2="bash <(curl -Ls https://raw.githubusercontent.com/codacy/codacy-cli-v2/main/codacy-cli.sh)"
```

## Installation

Before running the analysis, install the specified tools:

```bash
codacy-cli-v2 install
```

### Run Analysis

To run ESLint and output the results to the terminal:

```bash
codacy-cli-v2 analyze --tool eslint
```

To store the results as SARIF in a file:

```bash
codacy-cli-v2 analyze -t eslint -o eslint.sarif
```

## Upload Results

To upload a SARIF file to Codacy:

```bash
codacy-cli-v2 upload -s path/to/your.sarif -c your-commit-uuid -t your-project-token
```

### Example Repository

As an example, you can check https://github.com/troubleshoot-codacy/eslint-test-examples for a repository that has an action relying on this CLI.

