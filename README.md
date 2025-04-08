# codacy-cli-v2

The Codacy CLI (version 2) is a command-line tool that supports analyzing code using tools like ESLint and uploading the results in SARIF format to Codacy. 
The tool is invoked using the `codacy-cli` command, and provides two main commands: analyze and upload.

### Commands

- **`analyze` Command**: Runs ESLint analysis on the codebase.
    - `--output, -o`: Output file for the results.
    - `--tool, -t`: Specifies the tool to run analysis with (e.g., ESLint).
    - `--format`: Output format (use 'sarif' for SARIF format to terminal).
    - `--fix`: Automatically fixes issues when possible.

- **`upload` Command With Project Token**: Uploads a SARIF file containing analysis results to Codacy.
    - `--sarif-path, -s`: Path to the SARIF report.
    - `--commit-uuid, -c`: Commit UUID.
    - `--project-token, -t`: Project token for Codacy API.

- **`upload` Command With API Token**: Uploads a SARIF file containing analysis results to Codacy.
    - `--sarif-path, -s`: Path to the SARIF report.
    - `--commit-uuid, -c`: Commit UUID.
    - `--api-token, -a`: User level token for Codacy API.
    - `--provider, -p`: Provider name (e.g., gh, gl, bb).
    - `--owner, -o`: Repository owner.
    - `--repository, -r`: Repository name.

### Important Concepts

- **`.codacy/codacy.yaml`**: Configuration file to specify runtimes and tools versions for the CLI.
  ```yaml
  runtimes:
      - node@22.2.0
  tools:
      - eslint@9.3.0
  
- **`codacy-cli install`**: Command to install the specified node and eslint versions before running analysis.

## Download

### MacOS (brew)

To install `codacy-cli` using Homebrew:

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
alias codacy-cli="bash <(curl -Ls https://raw.githubusercontent.com/codacy/codacy-cli-v2/main/codacy-cli.sh)"
```

## Installation

Before running the analysis, install the specified tools:

```bash
codacy-cli install
```

### Run Analysis

To run ESLint and output the results to the terminal:

```bash
codacy-cli analyze --tool eslint
```

To output results in SARIF format to the terminal:

```bash
codacy-cli analyze --tool eslint --format sarif
```

To store the results as SARIF in a file:

```bash
codacy-cli analyze -t eslint --format sarif -o eslint.sarif
```

## Upload Results

To upload a SARIF file to Codacy:

```bash
codacy-cli upload -s path/to/your.sarif -c your-commit-uuid -t your-project-token
```

### Example Repository

As an example, you can check https://github.com/troubleshoot-codacy/eslint-test-examples for a repository that has an action relying on this CLI.
