#Requires -Version 5.1
$ErrorActionPreference = 'Stop'

$Repo = 'dinacomputer/cli'
$Binary = 'dina.exe'

if (-not $env:INSTALL_DIR -or $env:INSTALL_DIR -eq '') {
    $InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\dina'
} else {
    $InstallDir = $env:INSTALL_DIR
}

switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { $Arch = 'amd64' }
    'ARM64' { $Arch = 'arm64' }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

if (-not $env:VERSION -or $env:VERSION -eq '') {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
    $Version = $release.tag_name -replace '^v',''
    if (-not $Version) { throw 'Could not determine latest version.' }
} else {
    $Version = $env:VERSION
}

$Filename = "dina_${Version}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/v${Version}/${Filename}"

Write-Host "Downloading dina v$Version for windows/$Arch..."

$tmp = Join-Path $env:TEMP ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    $zipPath = Join-Path $tmp $Filename
    Invoke-WebRequest -Uri $Url -OutFile $zipPath -UseBasicParsing
    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force

    Write-Host "Installing to $InstallDir\$Binary..."
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    Copy-Item -Path (Join-Path $tmp $Binary) -Destination (Join-Path $InstallDir $Binary) -Force
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$pathParts = if ($userPath) { $userPath -split ';' } else { @() }
if ($pathParts -notcontains $InstallDir) {
    $newPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Host "Added $InstallDir to your user PATH. Open a new terminal for it to take effect."
}

Write-Host "dina v$Version installed successfully!"
Write-Host "Run 'dina --help' to get started."
