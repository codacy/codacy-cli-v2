#!/usr/bin/env bash

set -e +o pipefail

os_name=$(uname)
os_name_arch=$(uname -sm)

download_file() {
    local url="$1"

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
    local file_name="$2"
    local output_folder="$3"
    local output_filename="$4"
    local checksum_url="$5"
    local original_folder
    original_folder="$(pwd)"

    cd "$output_folder"

    download_file "$url"
    # checksum "$file_name" "$checksum_url"
    if [ "$os_name_arch" = "Linux x86_64" ] || [ "$os_name_arch" = "Darwin x86_64" ]; then
        mv "$file_name" "$output_filename"
    fi

    cd "$original_folder"
}

download_reporter() {
    # OS name lower case
    suffix=$(echo "$os_name" | tr '[:upper:]' '[:lower:]')

    local binary_name="codacy-cli-v2-$suffix"
    local reporter_path=$1
    local reporter_folder=$2
    local reporter_filename=$3

    if [ ! -f "$reporter_path" ]
    then
        echo "$i" "Downloading the codacy cli v2 $binary_name... ($CODACY_CLI_V2_VERSION)"

        binary_url="https://github.com/codacy/codacy-cli-v2/releases/download/${CODACY_CLI_V2_VERSION}/codacy-cli-v2_${CODACY_CLI_V2_VERSION}_darwin_amd64.tar.gz"
        # echo $binary_url
        # checksum_url="https://github.com/codacy/codacy-coverage-reporter/releases/download/$CODACY_CLI_V2_VERSION/$binary_name.SHA512SUM"

        download "$binary_url" "$binary_name" "$reporter_folder" "$reporter_filename" "$checksum_url"
    else
        echo "$i" "Codacy reporter $binary_name already in cache"
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

reporter_filename="codacy-cli-v2"

# Folder containing the binary
reporter_folder="$CODACY_CLI_V2_TMP_FOLDER"/"$CODACY_CLI_V2_VERSION"

# Create the reporter folder if not exists
mkdir -p "$reporter_folder"

# Set binary path
reporter_path="$reporter_folder"/"$reporter_filename"

download_reporter "$reporter_path" "$reporter_folder" "$reporter_filename"

chmod +x "$reporter_path"
run_command="$reporter_path"

if [ -z "$run_command" ]
then
    fatal "Codacy cli v2 binary could not be found."
fi

if [ "$#" -eq 1 ] && [ "$1" = "download" ];
then
    echo "$g" "Codacy reporter download succeeded";
else
    eval "$run_command $*"
fi