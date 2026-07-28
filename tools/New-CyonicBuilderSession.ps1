[CmdletBinding()]
param(
    [Parameter()]
    [string] $SourceRepository = "",

    [Parameter(Mandatory)]
    [string] $SessionDirectory,

    [Parameter()]
    [string] $BaseRef = "main",

    [Parameter()]
    [string] $SessionName = ("builder-" + (Get-Date -Format "yyyyMMdd-HHmmss"))
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    throw "Required command 'git' is not available."
}

if ([string]::IsNullOrWhiteSpace($SourceRepository)) {
    $SourceRepository = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
}

$source = (Resolve-Path -LiteralPath $SourceRepository).Path
if (-not (Get-Item -LiteralPath $source).PSIsContainer) {
    throw "SourceRepository must be a directory."
}

git -C $source rev-parse --is-inside-work-tree | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "SourceRepository is not a Git working tree."
}

$target = [System.IO.Path]::GetFullPath($SessionDirectory)
if (Test-Path -LiteralPath $target) {
    throw "SessionDirectory already exists: $target"
}

$targetParent = Split-Path -Parent $target
if (-not (Test-Path -LiteralPath $targetParent)) {
    New-Item -ItemType Directory -Path $targetParent | Out-Null
}
$resolvedParent = (Resolve-Path -LiteralPath $targetParent).Path
if (-not $target.StartsWith($resolvedParent, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Resolved session target escaped its declared parent."
}

$baseCommit = (git -C $source rev-parse "$BaseRef^{commit}").Trim()
if ($LASTEXITCODE -ne 0 -or -not $baseCommit) {
    throw "Cannot resolve BaseRef '$BaseRef'."
}

$bundle = Join-Path $resolvedParent (".cyonic-source-" + [guid]::NewGuid().ToString("N") + ".bundle")
try {
    git -C $source bundle create $bundle $BaseRef
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create source bundle."
    }

    git clone $bundle $target
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to clone isolated session repository."
    }
}
finally {
    if (Test-Path -LiteralPath $bundle) {
        Remove-Item -LiteralPath $bundle -Force
    }
}

git -C $target remote remove origin
if ($LASTEXITCODE -ne 0) {
    throw "Failed to remove bundle origin."
}

$branch = "session/$SessionName"
git -C $target switch -c $branch $baseCommit
if ($LASTEXITCODE -ne 0) {
    throw "Failed to create session branch '$branch'."
}

git -C $target config credential.helper ""
git -C $target config core.hooksPath ".cyonic-no-hooks"
New-Item -ItemType Directory -Path (Join-Path $target ".cyonic-no-hooks") | Out-Null

$sessionRecord = [ordered]@{
    schema = "urn:cyonic:builder-session:v1"
    createdAt = (Get-Date).ToUniversalTime().ToString("o")
    sourceRepository = $source
    sourceBaseRef = $BaseRef
    sourceBaseCommit = $baseCommit
    sessionBranch = $branch
    cloudRemoteConfigured = $false
    credentialHelperConfigured = $false
    authorityEffect = "NONE"
    securityBoundary = "GUARDRAIL_ONLY"
    limits = @(
        "This clone has no configured remote and disables its local credential helper.",
        "It does not restrict filesystem access, network access, process execution, or global OS credentials.",
        "A process with sufficient access can reconfigure Git.",
        "Use Windows Sandbox, a container, or a restricted OS account for a real security boundary."
    )
}
$sessionRecord | ConvertTo-Json -Depth 5 |
    Set-Content -LiteralPath (Join-Path $target ".cyonic-session.json") -Encoding utf8

Write-Host ""
Write-Host "Created guarded builder clone:" -ForegroundColor Green
Write-Host "  $target"
Write-Host "Base commit:"
Write-Host "  $baseCommit"
Write-Host "Session branch:"
Write-Host "  $branch"
Write-Host ""
Write-Host "No cloud remote is configured. This is a guardrail, not an OS sandbox." -ForegroundColor Yellow
Write-Host ""
Write-Host "Enter the session after suppressing inherited Git credentials:" -ForegroundColor Cyan
Write-Host "  `$env:GIT_CONFIG_GLOBAL='NUL'"
Write-Host "  `$env:GIT_CONFIG_NOSYSTEM='1'"
Write-Host "  `$env:GIT_TERMINAL_PROMPT='0'"
Write-Host "  `$env:GCM_INTERACTIVE='never'"
Write-Host "  Remove-Item Env:GITHUB_TOKEN,Env:GH_TOKEN,Env:SSH_AUTH_SOCK -ErrorAction SilentlyContinue"
Write-Host "  Set-Location '$target'"
