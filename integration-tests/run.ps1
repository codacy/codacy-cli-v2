# Stop on first error
$ErrorActionPreference = "Stop"

# Get the absolute path of the script's directory
$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$CLI_PATH = Join-Path (Get-Location) "cli-v2.exe"

Write-Host "Script directory: $SCRIPT_DIR"
Write-Host "Current working directory: $(Get-Location)"

# Check if API token is provided for token-based test
if (-not $env:CODACY_API_TOKEN) {
    Write-Host "Warning: CODACY_API_TOKEN environment variable is not set. Token-based test will be skipped."
}

# Function to normalize and sort configuration values
function Normalize-Config {
    param (
        [string]$file
    )
    
    $ext = [System.IO.Path]::GetExtension($file).TrimStart('.')
    
    switch ($ext) {
        { $_ -in @('yaml', 'yml') } {
            # For YAML files, use yq to sort
            # Note: Requires yq to be installed on Windows
            yq e '.' $file | Sort-Object
        }
        { $_ -in @('rc', 'conf', 'ini', 'xml') } {
            # For other config files, sort values after '=' and keep other lines
            Get-Content $file | ForEach-Object {
                if ($_ -match '^[^#].*=.*,') {
                    $parts = $_ -split '='
                    $values = $parts[1] -split ',' | Sort-Object
                    "$($parts[0])=$($values -join ',')"
                } else {
                    $_
                }
            } | Sort-Object
        }
        default {
            # For other files, just sort
            Get-Content $file | Sort-Object
        }
    }
}

function Compare-Files {
    param (
        [string]$expectedDir,
        [string]$actualDir,
        [string]$label
    )
    
    # Compare files in current directory
    Get-ChildItem -Path $expectedDir -File | ForEach-Object {
        $filename = $_.Name
        $actualFile = Join-Path $actualDir $filename
        
        if (-not (Test-Path $actualFile)) {
            Write-Host "❌ $label/$filename does not exist in actual output"
            Write-Host "Expected: $($_.FullName)"
            Write-Host "Actual should be: $actualFile"
            exit 1
        }
        
        $expectedContent = Normalize-Config $_.FullName
        $actualContent = Normalize-Config $actualFile
        
        if (Compare-Object $expectedContent $actualContent) {
            Write-Host "❌ $label/$filename does not match expected"
            Write-Host "=== Expected (normalized) ==="
            $expectedContent
            Write-Host "=== Actual (normalized) ==="
            $actualContent
            Write-Host "=== Diff ==="
            Compare-Object $expectedContent $actualContent
            Write-Host "==================="
            exit 1
        } else {
            Write-Host "✅ $label/$filename matches expected"
        }
    }
    
    # Compare subdirectories
    Get-ChildItem -Path $expectedDir -Directory | ForEach-Object {
        $dirname = $_.Name
        if ($dirname -eq "logs") { return }
        
        # Handle .codacy directory specially
        if ($dirname -eq ".codacy") {
            $actualSubDir = $actualDir
        } else {
            $actualSubDir = Join-Path $actualDir $dirname
        }
        
        if (-not (Test-Path $actualSubDir)) {
            Write-Host "❌ Directory $label/$dirname does not exist in actual output"
            Write-Host "Expected: $($_.FullName)"
            Write-Host "Actual should be: $actualSubDir"
            exit 1
        }
        Compare-Files $_.FullName $actualSubDir "$label/$dirname"
    }
}

function Run-InitTest {
    param (
        [string]$testDir,
        [string]$testName,
        [bool]$useToken
    )
    
    Write-Host "Running test: $testName"
    if (-not (Test-Path $testDir)) {
        Write-Host "❌ Test directory does not exist: $testDir"
        exit 1
    }
    
    # Store the original location
    $originalLocation = Get-Location
    
    try {
        # Change to the test directory
        Set-Location $testDir
        
        # Remove existing .codacy directory if it exists
        if (Test-Path ".codacy") {
            Remove-Item -Recurse -Force ".codacy"
        }
        
        if ($useToken) {
            if (-not $env:CODACY_API_TOKEN) {
                Write-Host "❌ Skipping token-based test: CODACY_API_TOKEN not set"
                return
            }
            & $CLI_PATH init --api-token $env:CODACY_API_TOKEN --organization troubleshoot-codacy-dev --provider gh --repository codacy-cli-test
        } else {
            & $CLI_PATH init
        }
        
        # Compare files using relative paths
        Compare-Files "expected" ".codacy" "Test $testName"
        Write-Host "✅ Test $testName completed successfully"
        Write-Host "----------------------------------------"
    }
    finally {
        # Always return to the original location
        Set-Location $originalLocation
    }
}

# Run both tests
Write-Host "Starting integration tests..."
Write-Host "----------------------------------------"

# Test 1: Init without token
Run-InitTest (Join-Path $SCRIPT_DIR "init-without-token") "init-without-token" $false

# Test 2: Init with token
Run-InitTest (Join-Path $SCRIPT_DIR "init-with-token") "init-with-token" $true

Write-Host "All tests completed successfully! 🎉" 