# Stop on first error
$ErrorActionPreference = "Stop"

# Get the absolute path of the script's directory and CLI path
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
    param ([string]$file)
    
    $ext = [System.IO.Path]::GetExtension($file).TrimStart('.')
    
    switch ($ext) {
        { $_ -in @('yaml', 'yml') } {
            # For YAML files, preserve structure and sort within sections
            $content = Get-Content $file
            $output = @()
            $currentSection = ""
            $sectionContent = @()
            
            foreach ($line in $content) {
                $line = $line.Trim()
                if ($line -match '^(\w+):$') {
                    # If we have a previous section, sort and add its content
                    if ($currentSection -and $sectionContent.Count -gt 0) {
                        $output += $currentSection
                        $output += ($sectionContent | Sort-Object)
                        $sectionContent = @()
                    }
                    $currentSection = $line
                }
                elseif ($line -match '^\s*-\s*') {
                    $sectionContent += $line
                }
                elseif ($line -match '\S') {
                    $output += $line
                }
            }
            
            # Add the last section
            if ($currentSection -and $sectionContent.Count -gt 0) {
                $output += $currentSection
                $output += ($sectionContent | Sort-Object)
            }
            
            # Add empty line at the end if the original had one
            if ($content[-1] -match '^\s*$') {
                $output += ""
            }
            
            $output
        }
        { $_ -in @('rc', 'conf', 'ini', 'xml') } {
            Get-Content $file | ForEach-Object {
                if ($_ -match '^[^#].*=.*,') {
                    $parts = $_ -split '='
                    $values = $parts[1] -split ',' | Sort-Object
                    "$($parts[0])=$($values -join ',')"
                } else { $_ }
            } | Sort-Object
        }
        default { Get-Content $file | Sort-Object }
    }
}

function Compare-Files {
    param (
        [string]$expectedDir,
        [string]$actualDir,
        [string]$label
    )
    
    # Compare files
    Get-ChildItem -Path $expectedDir -File | ForEach-Object {
        $actualFile = Join-Path $actualDir $_.Name
        
        # Skip Semgrep-related files only for init-without-token test
        if ($label -match 'init-without-token' -and $_.Name -match 'semgrep') {
            Write-Host "⏭️  Skipping Semgrep file: $label/$($_.Name)"
            continue
        }
        
        if (-not (Test-Path $actualFile)) {
            Write-Host "❌ $label/$($_.Name) does not exist in actual output"
            Write-Host "Expected: $($_.FullName)"
            Write-Host "Actual should be: $actualFile"
            exit 1
        }
        
        $expectedContent = Normalize-Config $_.FullName
        $actualContent = Normalize-Config $actualFile
        
        # Compare line by line
        $diff = Compare-Object $expectedContent $actualContent -PassThru
        if ($diff) {
            Write-Host "❌ $label/$($_.Name) does not match expected"
            Write-Host "=== Expected (normalized) ==="
            $expectedContent
            Write-Host "=== Actual (normalized) ==="
            $actualContent
            Write-Host "=== Diff ==="
            $diff
            Write-Host "==================="
            exit 1
        }
        Write-Host "✅ $label/$($_.Name) matches expected"
    }
    
    # Compare subdirectories
    Get-ChildItem -Path $expectedDir -Directory | Where-Object { $_.Name -ne "logs" } | ForEach-Object {
        $actualSubDir = Join-Path $actualDir $_.Name
        
        if (-not (Test-Path $actualSubDir)) {
            Write-Host "❌ Directory $label/$($_.Name) does not exist in actual output"
            Write-Host "Expected: $($_.FullName)"
            Write-Host "Actual should be: $actualSubDir"
            exit 1
        }
        Compare-Files $_.FullName $actualSubDir "$label/$($_.Name)"
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
    
    $originalLocation = Get-Location
    try {
        Set-Location $testDir
        if (Test-Path ".codacy") { Remove-Item -Recurse -Force ".codacy" }
        
        if ($useToken) {
            if (-not $env:CODACY_API_TOKEN) {
                Write-Host "❌ Skipping token-based test: CODACY_API_TOKEN not set"
                return
            }
            & $CLI_PATH init --api-token $env:CODACY_API_TOKEN --organization troubleshoot-codacy-dev --provider gh --repository codacy-cli-test
        } else {
            & $CLI_PATH init
        }
        
        Compare-Files "expected" ".codacy" "Test $testName"
        Write-Host "✅ Test $testName completed successfully"
        Write-Host "----------------------------------------"
    }
    finally {
        Set-Location $originalLocation
    }
}

# Run tests
Write-Host "Starting integration tests..."
Write-Host "----------------------------------------"

Run-InitTest (Join-Path $SCRIPT_DIR "init-without-token") "init-without-token" $false
Run-InitTest (Join-Path $SCRIPT_DIR "init-with-token") "init-with-token" $true

Write-Host "All tests completed successfully! 🎉" 