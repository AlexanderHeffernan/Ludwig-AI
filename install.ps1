# Ludwig Windows Installer
# Run this script in PowerShell to install Ludwig globally

$ErrorActionPreference = "Stop"

Write-Host "Fetching latest Ludwig release..." -ForegroundColor Cyan

# Get the latest release from GitHub API
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/AlexanderHeffernan/Ludwig-AI/releases/latest"

# Find the Windows binary
$asset = $release.assets | Where-Object { $_.name -match "Windows_x86_64\.zip" } | Select-Object -First 1

if (-not $asset) {
  Write-Host "Error: Could not find Windows binary in latest release" -ForegroundColor Red
  exit 1
}

$downloadUrl = $asset.browser_download_url
Write-Host "Downloading from: $downloadUrl" -ForegroundColor Cyan

# Create temp directory
$tempDir = [System.IO.Path]::GetTempPath() + "ludwig_install_" + (Get-Random)
New-Item -ItemType Directory -Path $tempDir | Out-Null

try {
  # Download the zip file
  $zipPath = Join-Path $tempDir "ludwig.zip"
  Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath
  
  # Extract
  Write-Host "Extracting..." -ForegroundColor Cyan
  Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force
  
  # Find ludwig executable
  $ludwigPath = Get-ChildItem -Path $tempDir -Recurse -Name "ludwig.exe" | Select-Object -First 1
  if (-not $ludwigPath) {
    $ludwigPath = "ludwig.exe"
  }
  
  $ludwigExe = Join-Path $tempDir $ludwigPath
  
  # Check if we need to create install directory
  $installDir = "C:\Program Files\Ludwig"
  if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
  }
  
  # Copy to install directory
  $targetPath = Join-Path $installDir "ludwig.exe"
  Write-Host "Installing to $targetPath..." -ForegroundColor Cyan
  Copy-Item -Path $ludwigExe -Destination $targetPath -Force
  
  # Add to PATH if not already there
  $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
  if ($currentPath -notlike "*$installDir*") {
    Write-Host "Adding to PATH..." -ForegroundColor Cyan
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:Path += ";$installDir"
  }
  
  Write-Host ""
  Write-Host "✓ Ludwig installed successfully!" -ForegroundColor Green
  Write-Host ""
  
  # Verify installation
  & $targetPath --version
  
} finally {
  # Cleanup
  Remove-Item -Path $tempDir -Recurse -Force
}
