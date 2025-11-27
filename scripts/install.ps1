# Coolify CLI Installer for Windows
# This script installs the coolify-cli from GitHub releases
# Supports Windows on amd64/arm64 architectures

#Requires -Version 5.1

[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string]$Version = "",
    
    [Parameter()]
    [switch]$User,
    
    [Parameter()]
    [switch]$Help,
    
    [Parameter()]
    [string]$InstallDir = ""
)

# Support environment variables for web-based installation
if (-not $Version -and $env:COOLIFY_VERSION) {
    $Version = $env:COOLIFY_VERSION
}
if (-not $User.IsPresent -and $env:COOLIFY_USER_INSTALL) {
    $User = [bool]($env:COOLIFY_USER_INSTALL -match '^(1|true|yes)$')
}
if (-not $InstallDir -and $env:COOLIFY_INSTALL_DIR) {
    $InstallDir = $env:COOLIFY_INSTALL_DIR
}

# Configuration
$Script:REPO = "coollabsio/coolify-cli"
$Script:BINARY_NAME = "coolify.exe"
$Script:GLOBAL_INSTALL_DIR = "$env:ProgramFiles\Coolify"
$Script:USER_INSTALL_DIR = "$env:LOCALAPPDATA\Coolify"
$Script:TEMP_FILE = ""

# Error action preference
$ErrorActionPreference = "Stop"

# Cleanup function
function Cleanup {
    if ($Script:TEMP_FILE -and (Test-Path $Script:TEMP_FILE)) {
        Remove-Item -Path $Script:TEMP_FILE -Force -ErrorAction SilentlyContinue
    }
}

# Register cleanup on exit
Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action { Cleanup } | Out-Null

# Show help
function Show-Help {
    Write-Host @"
Coolify CLI Installer for Windows

Usage: .\install.ps1 [OPTIONS] [VERSION]

OPTIONS:
  -User           Install to user directory (no admin rights required)
  -InstallDir     Custom installation directory
  -Help           Show this help message

ARGUMENTS:
  VERSION         Specific version to install (e.g., v1.0.0)
                  If not specified, installs the latest release

EXAMPLES:
  .\install.ps1                    Install latest version to Program Files
  .\install.ps1 -User              Install latest version to user directory
  .\install.ps1 v1.0.0             Install specific version to Program Files
  .\install.ps1 -User v1.0.0       Install specific version to user directory
  .\install.ps1 -InstallDir "C:\Tools"  Install to custom directory

NOTES:
  - Administrator privileges are required for global installation
  - User installation does not require admin rights
  - The installation directory will be added to PATH if not already present

"@
    exit 0
}

# Write colored output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$ForegroundColor = "White"
    )
    Write-Host $Message -ForegroundColor $ForegroundColor
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput $Message -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput $Message -ForegroundColor Yellow
}

function Write-ErrorMessage {
    param([string]$Message)
    Write-ColorOutput $Message -ForegroundColor Red
}

# Error handler
function Stop-WithError {
    param([string]$Message)
    Write-ErrorMessage "Error: $Message"
    Cleanup
    throw $Message
}

# Check if running as administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Detect platform and architecture
function Get-PlatformInfo {
    $os = "windows"
    
    # Detect architecture
    $arch = switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { "amd64" }
        "ARM64" { "arm64" }
        "x86" { Stop-WithError "32-bit Windows is not supported" }
        default { Stop-WithError "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
    }
    
    return @{
        OS = $os
        Arch = $arch
    }
}

# Fetch latest release version from GitHub
function Get-LatestVersion {
    Write-Host "Fetching latest release version..."
    
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Script:REPO/releases/latest" -Method Get
        $latestVersion = $response.tag_name
        
        if (-not $latestVersion) {
            Stop-WithError "Failed to fetch latest release version from GitHub"
        }
        
        return $latestVersion
    }
    catch {
        Stop-WithError "Failed to fetch latest release version: $($_.Exception.Message)"
    }
}

# Validate version format
function Test-VersionFormat {
    param([string]$VersionString)
    
    # Check if version matches semantic versioning (with or without 'v' prefix)
    if ($VersionString -notmatch '^v?\d+\.\d+\.\d+(-.*)?$') {
        Stop-WithError "Invalid version format: $VersionString`nExpected format: v1.0.0 or 1.0.0"
    }
    
    # Ensure version starts with 'v' for GitHub releases
    if ($VersionString -notmatch '^v') {
        $VersionString = "v$VersionString"
    }
    
    return $VersionString
}

# Check if coolify is already installed
function Test-ExistingInstallation {
    param([string]$NewVersion)
    
    try {
        $coolifyCmd = Get-Command coolify -ErrorAction SilentlyContinue
        if ($coolifyCmd) {
            $currentVersion = & coolify version 2>$null | Select-Object -First 1
            if (-not $currentVersion) {
                $currentVersion = "unknown"
            }
            
            Write-Warning "Coolify CLI is already installed: $currentVersion"
            Write-Host "This will upgrade/reinstall to version " -NoNewline
            Write-Success $NewVersion
            
            $response = Read-Host "Continue? [y/N]"
            if ($response -notmatch '^[Yy]$') {
                Write-Host "Installation cancelled."
                exit 0
            }
        }
    }
    catch {
        # Coolify not found, continue with installation
    }
}

# Download and install binary from GitHub release
function Install-FromGitHub {
    param(
        [string]$Repo,
        [string]$Release,
        [string]$Name,
        [string]$InstallDirectory,
        [hashtable]$Platform
    )
    
    # Clean version (remove 'v' prefix if present)
    $cleanVersion = $Release -replace '^v', ''
    
    $filename = "${Name}_${cleanVersion}_$($Platform.OS)_$($Platform.Arch).zip"
    $downloadUrl = "https://github.com/${Repo}/releases/download/${Release}/${filename}"
    
    Write-Success "Downloading $Name $Release"
    Write-Host "Platform: $($Platform.OS)/$($Platform.Arch)"
    Write-Host "URL: $downloadUrl"
    
    # Create temp file
    $Script:TEMP_FILE = [System.IO.Path]::GetTempFileName() + ".zip"
    
    # Download file
    try {
        $ProgressPreference = 'SilentlyContinue'  # Speed up download
        Invoke-WebRequest -Uri $downloadUrl -OutFile $Script:TEMP_FILE -ErrorAction Stop
        $ProgressPreference = 'Continue'
    }
    catch {
        Stop-WithError "Failed to download from ${downloadUrl}`nPlease check if the version exists or try again later.`nError: $($_.Exception.Message)"
    }
    
    # Verify downloaded file is not empty
    if ((Get-Item $Script:TEMP_FILE).Length -eq 0) {
        Stop-WithError "Downloaded file is empty"
    }
    
    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallDirectory)) {
        Write-Host "Creating directory: $InstallDirectory"
        try {
            New-Item -ItemType Directory -Path $InstallDirectory -Force | Out-Null
        }
        catch {
            Stop-WithError "Failed to create directory $InstallDirectory : $($_.Exception.Message)"
        }
    }
    
    # Extract binary
    Write-Host "Installing $Name to $InstallDirectory\$Script:BINARY_NAME"
    
    try {
        # Remove existing binary if present
        $binaryPath = Join-Path $InstallDirectory $Script:BINARY_NAME
        if (Test-Path $binaryPath) {
            Remove-Item -Path $binaryPath -Force
        }
        
        # Extract zip file
        Expand-Archive -Path $Script:TEMP_FILE -DestinationPath $InstallDirectory -Force
    }
    catch {
        Stop-WithError "Failed to extract binary: $($_.Exception.Message)"
    }
    
    # Verify installation
    $installedBinary = Join-Path $InstallDirectory $Script:BINARY_NAME
    if (-not (Test-Path $installedBinary)) {
        Stop-WithError "Binary was not installed to $installedBinary"
    }
    
    Write-Success "✓ $Name installed successfully to $installedBinary"
    
    # Add to PATH if not already present
    Add-ToPath -Directory $InstallDirectory
    
    # Show installed version
    try {
        $installedVersion = & $installedBinary version 2>$null | Select-Object -First 1
        if ($installedVersion) {
            Write-Host "Installed version: " -NoNewline
            Write-Success $installedVersion
        }
    }
    catch {
        # Version check failed, but installation was successful
    }
}

# Add directory to PATH
function Add-ToPath {
    param([string]$Directory)
    
    # Determine which PATH to modify (user or system)
    $pathScope = if ($User -or -not (Test-Administrator)) { "User" } else { "Machine" }
    
    # Get current PATH
    $currentPath = [Environment]::GetEnvironmentVariable("Path", $pathScope)
    
    # Check if directory is already in PATH
    $pathDirs = $currentPath -split ';' | ForEach-Object { $_.Trim() }
    if ($pathDirs -contains $Directory) {
        Write-Host "Directory is already in PATH"
        return
    }
    
    Write-Host "Adding $Directory to PATH ($pathScope)..."
    
    try {
        # Add to PATH
        $newPath = "$currentPath;$Directory"
        [Environment]::SetEnvironmentVariable("Path", $newPath, $pathScope)
        
        # Update current session PATH
        $env:Path = "$env:Path;$Directory"
        
        Write-Success "✓ Added to PATH. You may need to restart your terminal for changes to take effect."
    }
    catch {
        Write-Warning "Failed to add to PATH automatically: $($_.Exception.Message)"
        Write-Host "Please add the following directory to your PATH manually:"
        Write-Host "  $Directory"
    }
}

# Main installation flow
function Install-CoolifyCLI {
    Write-Host "Coolify CLI Installer for Windows"
    Write-Host "=================================="
    Write-Host ""
    
    # Show help if requested
    if ($Help) {
        Show-Help
    }
    
    # Detect platform
    $platform = Get-PlatformInfo
    
    # Determine version to install
    $versionToInstall = if ($Version) {
        Test-VersionFormat -VersionString $Version
    } else {
        Get-LatestVersion
    }
    
    Write-Host "Version to install: $versionToInstall"
    Write-Host ""
    
    # Check existing installation
    Test-ExistingInstallation -NewVersion $versionToInstall
    
    # Determine install directory
    $installDirectory = if ($InstallDir) {
        $InstallDir
    } elseif ($User) {
        $Script:USER_INSTALL_DIR
    } else {
        $Script:GLOBAL_INSTALL_DIR
    }
    
    # Check for admin rights if installing globally
    if (-not $User -and -not $InstallDir -and -not (Test-Administrator)) {
        Write-Warning "Global installation requires administrator privileges."
        Write-Host "Please run this script as Administrator, or use -User flag for user installation."
        Write-Host ""
        $response = Read-Host "Switch to user installation? [Y/n]"
        if ($response -match '^[Nn]$') {
            Stop-WithError "Installation cancelled. Please run as Administrator for global installation."
        }
        $installDirectory = $Script:USER_INSTALL_DIR
    }
    
    $installMode = if ($User -or $installDirectory -eq $Script:USER_INSTALL_DIR) {
        "User (no admin rights required)"
    } else {
        "Global (administrator)"
    }
    
    Write-Host "Install mode: $installMode"
    Write-Host "Install directory: $installDirectory"
    Write-Host ""
    
    # Download and install
    Install-FromGitHub -Repo $Script:REPO -Release $versionToInstall -Name "coolify-cli" -InstallDirectory $installDirectory -Platform $platform
    
    Write-Host ""
    Write-Success "Installation complete!"
    Write-Host "Run 'coolify --help' to get started"
    Write-Host ""
    Write-Host "Note: If 'coolify' command is not found, please restart your terminal."
}

# Run main installation
try {
    Install-CoolifyCLI
}
catch {
    Write-ErrorMessage "Unexpected error: $($_.Exception.Message)"
    Write-ErrorMessage $_.ScriptStackTrace
    Cleanup
}
finally {
    Cleanup
}
