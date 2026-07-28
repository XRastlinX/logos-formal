param(
    [ValidatePattern("^[A-Za-z0-9._-]{1,64}$")]
    [string]$Identity = "friend-node-01",

    [ValidateSet("codex", "antigravity", "manual")]
    [string]$Pathway = "codex"
)

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$Manifest = Get-Content -LiteralPath (Join-Path $Root "VERSION.json") -Raw | ConvertFrom-Json
$CanonDir = Join-Path $Root "canon\logos-formal"

foreach ($tool in @("git", "go")) {
    if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) {
        throw "Required tool '$tool' was not found on PATH."
    }
}

$goVersion = (& go env GOVERSION).Trim()
if ($goVersion -notmatch "^go(?<major>[0-9]+)\.(?<minor>[0-9]+)") {
    throw "Could not parse Go version: $goVersion"
}
$goMajor = [int]$Matches.major
$goMinor = [int]$Matches.minor
if ($goMajor -lt 1 -or ($goMajor -eq 1 -and $goMinor -lt 24)) {
    throw "Go 1.24 or newer is required. Found: $goVersion"
}

if (Test-Path -LiteralPath (Join-Path $CanonDir ".git")) {
    $dirty = & git -C $CanonDir status --porcelain
    if ($LASTEXITCODE -ne 0) {
        throw "Could not inspect the existing canon checkout."
    }
    if ($dirty) {
        throw "The existing canon checkout has local changes. Refusing to replace or discard them."
    }
}
else {
    if (Test-Path -LiteralPath $CanonDir) {
        throw "The canon path exists but is not a Git checkout: $CanonDir"
    }
    & git clone --filter=blob:none --no-checkout $Manifest.sourceRepository $CanonDir
    if ($LASTEXITCODE -ne 0) {
        throw "Git clone failed."
    }
}

& git -C $CanonDir fetch origin $Manifest.sourceBranch
if ($LASTEXITCODE -ne 0) {
    throw "Git fetch failed."
}
& git -C $CanonDir switch --detach $Manifest.sourceCommit
if ($LASTEXITCODE -ne 0) {
    throw "Could not check out the pinned source commit."
}

$actualCommit = (& git -C $CanonDir rev-parse HEAD).Trim()
if ($actualCommit -ne $Manifest.sourceCommit) {
    throw "Source mismatch: expected $($Manifest.sourceCommit), found $actualCommit."
}
$actualTree = (& git -C $CanonDir rev-parse "HEAD^{tree}").Trim()
if ($LASTEXITCODE -ne 0 -or $actualTree -ne $Manifest.sourceTree) {
    throw "Source tree mismatch: expected $($Manifest.sourceTree), found $actualTree."
}

$example = Get-Content -LiteralPath (Join-Path $Root "config\node.example.yaml") -Raw
$nodeConfig = $example.Replace("NODE_ID", $Identity).Replace("PATHWAY", $Pathway)
$nodeConfigPath = Join-Path $Root "config\node.yaml"
if (-not (Test-Path -LiteralPath $nodeConfigPath)) {
    [System.IO.File]::WriteAllText($nodeConfigPath, $nodeConfig, [System.Text.UTF8Encoding]::new($false))
}
else {
    $existingNodeConfig = Get-Content -LiteralPath $nodeConfigPath -Raw
    if ($existingNodeConfig -ne $nodeConfig) {
        throw "Existing node configuration does not match the requested identity/pathway. Refusing to report unapplied values."
    }
}

& (Join-Path $PSScriptRoot "validate.ps1")
if ($LASTEXITCODE -ne 0) {
    throw "Node validation failed."
}

Write-Host ""
Write-Host "NODE_SETUP_COMPLETE" -ForegroundColor Green
Write-Host "Node: $Identity"
Write-Host "Pathway: $Pathway"
Write-Host "Pinned commit: $actualCommit"
Write-Host "Pinned tree: $actualTree"
Write-Host "Group rules: $(Join-Path $Root 'GROUP_RULES.md')"
Write-Host "Local sessions remain 010 with authority_effect NONE."
