# Stop on first error
$ErrorActionPreference = "Stop"

# Get the absolute path of the script's directory and CLI path
$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$CLI_PATH = Join-Path (Get-Location) "cli-v2.exe"

Write-Host "=== Environment Information ==="
Write-Host "Script directory: $SCRIPT_DIR"
Write-Host "Current working directory: $(Get-Location)"
Write-Host "CLI Path: $CLI_PATH"
Write-Host "OS: $([System.Environment]::OSVersion.Platform)"
Write-Host "PowerShell Version: $($PSVersionTable.PSVersion)"
Write-Host "==============================`n"

# Check if API token is provided for token-based test
if (-not $env:CODACY_API_TOKEN) {
    Write-Host "Warning: CODACY_API_TOKEN environment variable is not set. Token-based test will be skipped."
}

# Function to normalize and sort configuration values
function Normalize-Config {
    param ([string]$file)
    
    Write-Host "Normalizing config file: $file"
    $ext = [System.IO.Path]::GetExtension($file).TrimStart('.')
    
    switch ($ext) {
        { $_ -in @('yaml', 'yml') } {
            Normalize-YamlConfig $file
        }
        { $_ -in @('mjs', 'js') } {
            Normalize-EslintConfig $file
        }
        'toml' {
            Normalize-TomlConfig $file
        }
        { $_ -in @('rc', 'conf', 'ini') } {
            Normalize-RcConfig $file
        }
        'xml' {
            Normalize-XmlConfig $file
        }
        default { 
            Get-Content $file | Sort-Object 
        }
    }
}

# Normalize YAML configuration files
function Normalize-YamlConfig {
    param([string]$file)
    
    # For YAML files, try to preserve structure - just return as-is for now
    # Complex YAML sorting can break structure, so we keep original order
    Get-Content $file
}

# Normalize ESLint configuration files (.mjs/.js)
function Normalize-EslintConfig {
    param([string]$file)
    
    $content = Get-Content $file
    $output = @()
    $inRules = $false
    $ruleLines = @()
    
    foreach ($line in $content) {
        if ($line -match 'rules: \{') {
            $output += $line
            $inRules = $true
        } elseif ($inRules -and $line -match '^\s*\}') {
            # Sort collected rule lines (since expected file already has properties sorted)
            $output += ($ruleLines | Sort-Object)
            $ruleLines = @()
            $inRules = $false
            $output += $line
        } elseif ($inRules) {
            # Collect rule lines for sorting
            $ruleLines += $line
        } else {
            $output += $line
        }
    }
    $output
}

# Normalize TOML configuration files
function Normalize-TomlConfig {
    param([string]$file)
    
    Get-Content $file | ForEach-Object {
        if ($_ -match '^[^#].*=.*\[.*\]') {
            # Handle TOML arrays like: rules = ["a", "b", "c"]
            $parts = $_ -split '='
            if ($parts[1] -match '\[(.*)\]') {
                $arrayContent = $matches[1]
                $values = $arrayContent -split ',\s*' | Sort-Object
                "$($parts[0])=[$($values -join ', ')]"
            } else { $_ }
        } elseif ($_ -match '^[^#].*=.*,') {
            # Handle simple comma-separated values
            $parts = $_ -split '='
            $values = $parts[1] -split ',' | Sort-Object
            "$($parts[0])=$($values -join ',')"
        } else { $_ }
    } | Sort-Object
}

# Normalize RC/INI configuration files
function Normalize-RcConfig {
    param([string]$file)
    
    Get-Content $file | ForEach-Object {
        if ($_ -match '^[^#].*=.*,') {
            # Handle simple comma-separated values
            $parts = $_ -split '='
            $values = $parts[1] -split ',' | Sort-Object
            "$($parts[0])=$($values -join ',')"
        } else { $_ }
    } | Sort-Object
}

# Normalize XML configuration files
function Normalize-XmlConfig {
    param([string]$file)
    
    $lines = Get-Content $file
    $rules = @()
    $output = @()
    $endTag = $null
    $inProps = $false
    $properties = @()
    $propsStart = $null

    foreach ($line in $lines) {
        $trimmed = $line.TrimStart()
        
        if ($trimmed -match '^<properties>') {
            $inProps = $true
            $propsStart = $trimmed
            $properties = @()
        } elseif ($trimmed -match '^</properties>') {
            $inProps = $false
            # Add sorted properties block
            $output += $propsStart
            $output += ($properties | Sort-Object)
            $output += $trimmed
        } elseif ($inProps -and $trimmed -match '^<property') {
            $properties += $trimmed
        } elseif ($trimmed -match '^<rule ref=') {
            $rules += $trimmed
        } elseif ($trimmed -match '^</ruleset>') {
            $endTag = $trimmed
        } else {
            $output += $trimmed
        }
    }
    
    $output + ($rules | Sort-Object) + $endTag
}



function Compare-Files {
    param (
        [string]$expectedDir,
        [string]$actualDir,
        [string]$label
    )
    
    Write-Host "`n=== Starting Directory Comparison ==="
    Write-Host "Comparing directories:"
    Write-Host "Expected dir: $expectedDir"
    Write-Host "Actual dir: $actualDir"
    Write-Host "Label: $label"
    
    # Convert to absolute paths and normalize separators
    $expectedDir = (Resolve-Path $expectedDir).Path
    $actualDir = (Resolve-Path $actualDir).Path
    
    Write-Host "Resolved paths:"
    Write-Host "Expected dir (resolved): $expectedDir"
    Write-Host "Actual dir (resolved): $actualDir"
    
    # List directory contents before comparison
    Write-Host "`nExpected directory contents:"
    Get-ChildItem -Path $expectedDir -Recurse | ForEach-Object {
        Write-Host "  $($_.FullName.Replace($expectedDir, ''))"
    }
    
    Write-Host "`nActual directory contents:"
    Get-ChildItem -Path $actualDir -Recurse | ForEach-Object {
        Write-Host "  $($_.FullName.Replace($actualDir, ''))"
    }
    
    # Compare files
    Get-ChildItem -Path $expectedDir -File | ForEach-Object {
        $relativePath = $_.FullName.Replace($expectedDir, '').TrimStart('\')
        $actualFile = Join-Path $actualDir $relativePath
        Write-Host "`nChecking file: $relativePath"
        Write-Host "Expected file: $($_.FullName)"
        Write-Host "Actual file: $actualFile"
        
        if (-not (Test-Path $actualFile)) {
            Write-Host "❌ $label/$relativePath does not exist in actual output"
            Write-Host "Expected: $($_.FullName)"
            Write-Host "Actual should be: $actualFile"
            Write-Host "Current directory structure:"
            Get-ChildItem -Path $actualDir -Recurse | ForEach-Object {
                Write-Host "  $($_.FullName)"
            }
            exit 1
        }
        
        Write-Host "Comparing file contents..."
        $expectedContent = Normalize-Config $_.FullName
        $actualContent = Normalize-Config $actualFile
        
        # Compare line by line
        $diff = Compare-Object $expectedContent $actualContent -PassThru
        if ($diff) {
            Write-Host "❌ $label/$relativePath does not match expected"
            Write-Host "=== Expected (normalized) ==="
            $expectedContent
            Write-Host "=== Actual (normalized) ==="
            $actualContent
            Write-Host "=== Diff ==="
            $diff
            Write-Host "==================="
            exit 1
        }
        Write-Host "✅ $label/$relativePath matches expected"
    }
    
    # Compare subdirectories
    Get-ChildItem -Path $expectedDir -Directory | Where-Object { $_.Name -ne "logs" } | ForEach-Object {
        $relativePath = $_.FullName.Replace($expectedDir, '').TrimStart('\')
        $actualSubDir = Join-Path $actualDir $relativePath
        Write-Host "`nChecking subdirectory: $relativePath"
        Write-Host "Expected dir: $($_.FullName)"
        Write-Host "Actual dir: $actualSubDir"
        
        if (-not (Test-Path $actualSubDir)) {
            Write-Host "❌ Directory $label/$relativePath does not exist in actual output"
            Write-Host "Expected: $($_.FullName)"
            Write-Host "Actual should be: $actualSubDir"
            Write-Host "Current directory structure:"
            Get-ChildItem -Path $actualDir -Recurse | ForEach-Object {
                Write-Host "  $($_.FullName)"
            }
            exit 1
        }
        Compare-Files $_.FullName $actualSubDir "$label/$relativePath"
    }
    
    Write-Host "`n=== Directory Comparison Complete ==="
}

function Run-InitTest {
    param (
        [string]$testDir,
        [string]$testName,
        [bool]$useToken
    )
    
    Write-Host "`n=== Running Test: $testName ==="
    Write-Host "Test directory: $testDir"
    Write-Host "Using token: $useToken"
    
    if (-not (Test-Path $testDir)) {
        Write-Host "❌ Test directory does not exist: $testDir"
        exit 1
    }
    
    $originalLocation = Get-Location
    try {
        Write-Host "Changing to test directory: $testDir"
        Set-Location $testDir
        Write-Host "Current location: $(Get-Location)"
        
        if (Test-Path ".codacy") {
            Write-Host "Removing existing .codacy directory"
            Remove-Item -Recurse -Force ".codacy"
        }
        
        if ($useToken) {
            if (-not $env:CODACY_API_TOKEN) {
                Write-Host "❌ Skipping token-based test: CODACY_API_TOKEN not set"
                return
            }
            Write-Host "Running CLI with token..."
            & $CLI_PATH init --api-token $env:CODACY_API_TOKEN --organization troubleshoot-codacy-dev --provider gh --repository codacy-cli-test
        } else {
            Write-Host "Running CLI without token..."
            & $CLI_PATH init
        }
        
        Write-Host "`nVerifying directory structure after CLI execution:"
        Get-ChildItem -Recurse | ForEach-Object {
            Write-Host "  $($_.FullName)"
        }
        
        Compare-Files "expected" ".codacy" "Test $testName"
        Write-Host "✅ Test $testName completed successfully"
        Write-Host "----------------------------------------"
    }
    catch {
        Write-Host "❌ Error during test execution:"
        Write-Host $_.Exception.Message
        Write-Host $_.ScriptStackTrace
        exit 1
    }
    finally {
        Write-Host "Returning to original location: $originalLocation"
        Set-Location $originalLocation
    }
}

# Run tests
Write-Host "`n=== Starting Integration Tests ==="
Write-Host "----------------------------------------"

Run-InitTest (Join-Path $SCRIPT_DIR "init-without-token") "init-without-token" $false
Run-InitTest (Join-Path $SCRIPT_DIR "init-with-token") "init-with-token" $true

Write-Host "`nAll tests completed successfully! 🎉" 