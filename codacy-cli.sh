#!/usr/bin/env bash

set -e +o pipefail

os_name=$(uname)
arch=$(uname -m)

case "$arch" in
"x86_64")
  arch="amd64"
  ;;
"x86")
  arch="386"
  ;;
esac

download_file() {
    local url="$1"

    echo "Download url: ${url}"
    if command -v curl > /dev/null 2>&1; then
        curl -# -LS "$url" -O
    elif command -v wget > /dev/null 2>&1; then
        wget "$url"
    else
        fatal "Could not find curl or wget, please install one."
    fi
}

download() {
    local url="$1"
    local output_folder="$2"

    ( cd "$output_folder" && download_file "$url" )
}

download_cli() {
    # OS name lower case
    suffix=$(echo "$os_name" | tr '[:upper:]' '[:lower:]')

    local bin_folder="$1"
    local bin_path="$2"

    if [ ! -f "$bin_path" ]; then
        echo "Downloading the codacy cli v2 version ($CODACY_CLI_V2_VERSION)"

        remote_file="codacy-cli-v2_${CODACY_CLI_V2_VERSION}_${suffix}_${arch}.tar.gz"
        url="https://github.com/codacy/codacy-cli-v2/releases/download/${CODACY_CLI_V2_VERSION}/${remote_file}"

        download "$url" "$bin_folder"
        tar xzfv "${bin_folder}/${remote_file}" -C "${bin_folder}"
    fi
}

# Temporary folder for downloaded files
if [ -z "$CODACY_CLI_V2_TMP_FOLDER" ]; then
    if [ "$os_name" = "Linux" ]; then
        CODACY_CLI_V2_TMP_FOLDER="$HOME/.cache/codacy/codacy-cli-v2"
    elif [ "$os_name" = "Darwin" ]; then
        CODACY_CLI_V2_TMP_FOLDER="$HOME/Library/Caches/Codacy/codacy-cli-v2"
    else
        CODACY_CLI_V2_TMP_FOLDER=".codacy-cli-v2"
    fi
fi

# if no version is specified, we fetch the latest
if [ -z "$CODACY_CLI_V2_VERSION" ]; then
  if [ -n "$GH_TOKEN" ]; then
    CODACY_CLI_V2_VERSION="$(curl -Lq --header "Authorization: Bearer $GH_TOKEN" "https://api.github.com/repos/codacy/codacy-cli-v2/releases/latest" 2>/dev/null | grep -m 1 tag_name | cut -d'"' -f4)"
  else
    CODACY_CLI_V2_VERSION="$(curl -Lq "https://api.github.com/repos/codacy/codacy-cli-v2/releases/latest" 2>/dev/null | grep -m 1 tag_name | cut -d'"' -f4)"
  fi
fi

# Folder containing the binary
bin_folder="${CODACY_CLI_V2_TMP_FOLDER}/${CODACY_CLI_V2_VERSION}"
# Create the folder if not exists
mkdir -p "$bin_folder"

# name of the binary
bin_name="codacy-cli-v2"

# Set binary path
bin_path="$bin_folder"/"$bin_name"

# download the tool
download_cli "$bin_folder" "$bin_path"
chmod +x "$bin_path"

run_command="$bin_path"
if [ -z "$run_command" ]; then
    fatal "Codacy cli v2 binary could not be found."
fi

if [ "$#" -eq 1 ] && [ "$1" = "download" ]; then
    echo "$g" "Codacy cli v2 download succeeded";
else
    eval "$run_command $*"
fi