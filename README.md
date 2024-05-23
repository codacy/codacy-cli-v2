# codacy-cli-v2

This is a POC for what could a new CLI for us. The idea is to rely on the native tools and SARIF format instead of relying on docker.

## Download

### MacOS brew

```bash
brew install codacy/codacy-cli-v2/codacy-cli-v2
```

### Linux

For linux we rely on `codacy-cli.sh` script on the root, to download the CLI you can:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/codacy/codacy-cli-v2/main/codacy-cli.sh)
```

You can either put the downloaded script on a specific file, or an alias that will download the script and look for changes:
```bash
alias codacy-cli-v2="bash <(curl -Ls https://raw.githubusercontent.com/codacy/codacy-cli-v2/main/codacy-cli.sh)"
```

## Important concepts

### `.codacy/.codacy.yaml`

Our CLI relies on the `.codacy/.codacy.yaml` to install and run the specified `node` and `eslint` versions.

Meaning that on the root of the repository that you want to analyse, you should create a `.codacy/.codacy.yaml` such as:

```yaml
runtimes:
    - node@22.2.0
tools:
    - eslint@9.3.0
```

### `codacy-cli-v2 install`

Before running, you need to do an install command, so it downloads the specified `node` and `eslint` versions.
To install the tools specified on your `.codacy/.codacy.yaml` you should:

```bash
  codacy-cli-v2 install
```

## Run

For now, we only support ESLint, to run and output the results to the terminal:

```bash
codacy-cli-v2 analyze --tool eslint
```

Alternatively, you can store the results as SARIF to a file:

```bash
codacy-cli-v2 analyze -t eslint -o eslint.sarif
```

## Example repo

As an example, that as an action that relies on this CLI, you can check <https://github.com/troubleshoot-codacy/eslint-test-examples>
