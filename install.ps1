# flickr-cli installer for Windows
# Usage: powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/thedavidweng/flickr-cli/main/install.ps1 | iex"

$ErrorActionPreference = "Stop"
$Repo = "thedavidweng/flickr-cli"
$Binary = "flickr"

function Step($msg)  { Write-Host "==> $msg" }
function Die($msg)   { Write-Error "ERROR: $msg"; exit 1 }

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Die "32-bit Windows is not supported."
}

$platformLabel = "windows/$arch"

# Resolve latest version
function Resolve-Version {
    $resp = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $resp.tag_name
}

# Main install
function Install-Flickr {
    Step "Installing flickr-cli ($platformLabel)"

    $version = Resolve-Version
    Step "Latest version: $version"

    $asset = "${Binary}_${version}_windows_${arch}.zip"
    $url = "https://github.com/$Repo/releases/download/$version/$asset"

    $installDir = if ($env:FLICKR_INSTALL_DIR) { $env:FLICKR_INSTALL_DIR } else {
        Join-Path $env:LOCALAPPDATA "flickr-cli\bin"
    }

    $tmpDir = Join-Path $env:TEMP "flickr-cli-install-$([guid]::NewGuid().ToString('N').Substring(0,8))"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null

    try {
        Step "Downloading $asset"
        $zipPath = Join-Path $tmpDir $asset
        Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

        Step "Installing to $installDir"
        Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force
        $exe = Join-Path $tmpDir "${Binary}.exe"
        if (-not (Test-Path $exe)) {
            Die "Could not find $Binary.exe in archive."
        }
        Copy-Item $exe (Join-Path $installDir "${Binary}.exe") -Force

        # Add to PATH if needed
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$installDir*") {
            [Environment]::SetEnvironmentVariable("Path", "$installDir;$userPath", "User")
            $env:Path = "$installDir;$env:Path"
            Step "Added $installDir to user PATH"
        }

        $versionOutput = & (Join-Path $installDir "${Binary}.exe") --version 2>$null
        Step "Installed $versionOutput"
    } finally {
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Uninstall-Flickr {
    $installDir = if ($env:FLICKR_INSTALL_DIR) { $env:FLICKR_INSTALL_DIR } else {
        Join-Path $env:LOCALAPPDATA "flickr-cli\bin"
    }
    $exe = Join-Path $installDir "${Binary}.exe"
    if (Test-Path $exe) {
        Step "Removing $exe"
        Remove-Item $exe -Force
    }
    Step "Uninstalled. You may also remove flickr-cli config from $env:APPDATA\flickr-cli\"
}

# Entry point
if ($args.Count -gt 0 -and $args[0] -eq "uninstall") {
    Uninstall-Flickr
} else {
    Install-Flickr
    Write-Host ""
    Step "Run 'flickr auth login --perms write' to get started."
    Step "Run 'flickr --help' to see available commands."
}
