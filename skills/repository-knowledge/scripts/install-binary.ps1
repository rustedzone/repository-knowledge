[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$')]
    [string]$Version,

    [ValidatePattern('^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$')]
    [string]$Repository = 'rustedzone/repository-knowledge',

    [string]$BinDir,
    [switch]$AddToPath,
    [switch]$Yes,
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'

if (-not [Environment]::Is64BitOperatingSystem) {
    throw 'Only 64-bit Windows release binaries are available.'
}
if ($env:PROCESSOR_ARCHITECTURE -ne 'AMD64') {
    throw "Unsupported Windows architecture: $($env:PROCESSOR_ARCHITECTURE)"
}

if ([string]::IsNullOrWhiteSpace($BinDir)) {
    $BinDir = Join-Path $env:LOCALAPPDATA 'Programs\repo-knowledge\bin'
}
$BinDir = [IO.Path]::GetFullPath($BinDir)
$Destination = Join-Path $BinDir 'repo-knowledge.exe'
$Artifact = 'repo-knowledge-windows-amd64.exe'
$ReleaseUrl = "https://github.com/$Repository/releases/download/$Version"
$PathEntries = @($env:Path -split ';' | ForEach-Object { $_.TrimEnd('\') })
$DestinationOnPath = $PathEntries -contains $BinDir.TrimEnd('\')

if (-not $DestinationOnPath -and -not $AddToPath) {
    throw "Destination is not on PATH: $BinDir. Rerun with -AddToPath after obtaining permission to change the user PATH, or pass -BinDir with a user-owned directory already on PATH."
}

Write-Host "Pinned release: $Version"
Write-Host "Source:         https://github.com/$Repository"
Write-Host "Artifact:       $Artifact"
Write-Host "Destination:    $Destination"
Write-Host 'Integrity:      SHA256SUMS; GitHub attestation too when a compatible gh command is available'
if ($AddToPath -and -not $DestinationOnPath) {
    Write-Host "PATH change:    add $BinDir to the current user's PATH"
}

if ($DryRun) {
    return
}
if (-not $Yes) {
    $Answer = Read-Host 'Install or replace this executable? [y/N]'
    if ($Answer -notmatch '^(?i:y|yes)$') {
        throw 'Installation cancelled.'
    }
}

$TemporaryDir = Join-Path ([IO.Path]::GetTempPath()) ("repo-knowledge-install-" + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $TemporaryDir | Out-Null
try {
    $ArtifactPath = Join-Path $TemporaryDir $Artifact
    $ChecksumsPath = Join-Path $TemporaryDir 'SHA256SUMS'
    Invoke-WebRequest -Uri "$ReleaseUrl/$Artifact" -OutFile $ArtifactPath -UseBasicParsing
    Invoke-WebRequest -Uri "$ReleaseUrl/SHA256SUMS" -OutFile $ChecksumsPath -UseBasicParsing

    $ChecksumLine = Get-Content $ChecksumsPath | Where-Object { $_ -match "^[0-9a-fA-F]{64}\s+\*?$([regex]::Escape($Artifact))$" } | Select-Object -First 1
    if (-not $ChecksumLine) {
        throw "SHA256SUMS has no entry for $Artifact"
    }
    $Expected = ($ChecksumLine -split '\s+')[0].ToLowerInvariant()
    $Actual = (Get-FileHash -Algorithm SHA256 -Path $ArtifactPath).Hash.ToLowerInvariant()
    if ($Actual -ne $Expected) {
        throw "Checksum mismatch for $Artifact"
    }

    $ReportedVersion = (& $ArtifactPath --version).Trim()
    if ($LASTEXITCODE -ne 0 -or $ReportedVersion -ne "repo-knowledge $($Version.TrimStart('v'))") {
        throw "Downloaded artifact reports unexpected version: $ReportedVersion"
    }

    $Gh = Get-Command gh -ErrorAction SilentlyContinue
    if ($Gh) {
        & $Gh.Source attestation verify --help *> $null
        if ($LASTEXITCODE -eq 0) {
            & $Gh.Source attestation verify $ArtifactPath --repo $Repository | Out-Null
            if ($LASTEXITCODE -ne 0) {
                throw 'GitHub artifact attestation verification failed.'
            }
            Write-Host 'Verified GitHub artifact attestation.'
        } else {
            Write-Host 'GitHub attestation verification skipped because the installed gh command is incompatible.'
        }
    } else {
        Write-Host 'GitHub attestation verification skipped because gh is unavailable.'
    }

    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    Copy-Item -Force -Path $ArtifactPath -Destination $Destination

    if ($AddToPath -and -not $DestinationOnPath) {
        $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        $NewUserPath = if ([string]::IsNullOrWhiteSpace($UserPath)) { $BinDir } else { "$UserPath;$BinDir" }
        [Environment]::SetEnvironmentVariable('Path', $NewUserPath, 'User')
        $env:Path = "$env:Path;$BinDir"
    }

    Write-Host "Installed repo-knowledge at $Destination (available as repo-knowledge on PATH; open terminals may need to be restarted)."
}
finally {
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $TemporaryDir
}
