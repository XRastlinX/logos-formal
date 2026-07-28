param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern("^[0-9a-fA-F]{40}$")]
    [string]$SourceCommit,

    [ValidatePattern("^[0-9]+\.[0-9]+\.[0-9]+$")]
    [string]$PackageVersion = "0.1.3",

    [ValidatePattern("^[A-Za-z0-9._/-]+$")]
    [string]$SourceBranch = "main",

    [ValidatePattern("^https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(?:\.git)?$")]
    [string]$SourceRepository = "https://github.com/XRastlinX/logos-formal.git",

    [string]$OutputDirectory
)

$ErrorActionPreference = "Stop"
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$TemplateRoot = Join-Path $RepoRoot "packaging\cyonic-node-setup-v0.1.3"
$BaseArchive = Join-Path $RepoRoot "dist\cyonic-node-setup-v0.1.2-portable-7e0f263.zip"
$ExpectedBaseHash = "8d1263b1dff0fcd84c42ac8911f132d270e8b172d50d06982d73d227499315a7"
$SourceCommit = $SourceCommit.ToLowerInvariant()

if (-not $OutputDirectory) {
    $OutputDirectory = Join-Path $RepoRoot "dist"
}
$OutputDirectory = [System.IO.Path]::GetFullPath($OutputDirectory)

if (-not (Test-Path -LiteralPath $BaseArchive)) {
    throw "Pinned base archive is missing: $BaseArchive"
}
$baseHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $BaseArchive).Hash.ToLowerInvariant()
if ($baseHash -ne $ExpectedBaseHash) {
    throw "Pinned base archive hash mismatch: $baseHash"
}

& git -C $RepoRoot cat-file -e "$SourceCommit`^{commit}"
if ($LASTEXITCODE -ne 0) {
    throw "Source commit is not present in the local repository: $SourceCommit"
}
& git -C $RepoRoot fetch --no-tags --depth=1 $SourceRepository $SourceCommit
if ($LASTEXITCODE -ne 0) {
    throw "Source commit could not be fetched from the declared source repository."
}
$fetchedCommit = (& git -C $RepoRoot rev-parse FETCH_HEAD).Trim().ToLowerInvariant()
if ($LASTEXITCODE -ne 0 -or $fetchedCommit -ne $SourceCommit) {
    throw "Declared source repository returned $fetchedCommit instead of $SourceCommit."
}
$currentHead = (& git -C $RepoRoot rev-parse HEAD).Trim().ToLowerInvariant()
if ($LASTEXITCODE -ne 0 -or $currentHead -ne $SourceCommit) {
    throw "Package overlays must be built from the checked-out source commit $SourceCommit; current HEAD is $currentHead."
}
$dirty = @(& git -C $RepoRoot status --porcelain --untracked-files=all)
if ($LASTEXITCODE -ne 0 -or $dirty.Count -ne 0) {
    throw "Package overlays must be built from a clean source worktree."
}
$sourceTree = (& git -C $RepoRoot rev-parse "$SourceCommit`^{tree}").Trim().ToLowerInvariant()
if ($LASTEXITCODE -ne 0 -or $sourceTree -notmatch "^[0-9a-f]{40}$") {
    throw "Could not resolve the exact source tree for $SourceCommit."
}
& git -C $RepoRoot grep -F --quiet "source-ref" $SourceCommit -- runtime/cyonic-service/http_probe.go
if ($LASTEXITCODE -ne 0) {
    throw "Source commit does not contain the explicit sourceRef probe contract."
}

$shortCommit = $SourceCommit.Substring(0, 7).ToLowerInvariant()
$archiveName = "cyonic-node-setup-v$PackageVersion-portable-$shortCommit.zip"
$archivePath = Join-Path $OutputDirectory $archiveName
if (Test-Path -LiteralPath $archivePath) {
    throw "Output already exists; refusing to overwrite: $archivePath"
}

Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem

$utf8 = [System.Text.UTF8Encoding]::new($false)
$files = [System.Collections.Generic.Dictionary[string, byte[]]]::new(
    [System.StringComparer]::Ordinal
)

$baseZip = [System.IO.Compression.ZipFile]::OpenRead($BaseArchive)
try {
    foreach ($entry in $baseZip.Entries) {
        if ([string]::IsNullOrEmpty($entry.Name)) {
            continue
        }
        $separator = $entry.FullName.IndexOf("/")
        if ($separator -le 0 -or $separator -eq $entry.FullName.Length - 1) {
            throw "Unexpected base archive entry: $($entry.FullName)"
        }
        $relative = $entry.FullName.Substring($separator + 1)
        if ($relative -eq "MANIFEST.sha256") {
            continue
        }
        $stream = $entry.Open()
        try {
            $memory = [System.IO.MemoryStream]::new()
            try {
                $stream.CopyTo($memory)
                $files[$relative] = $memory.ToArray()
            }
            finally {
                $memory.Dispose()
            }
        }
        finally {
            $stream.Dispose()
        }
    }
}
finally {
    $baseZip.Dispose()
}

$overlays = [ordered]@{
    "README.md"            = "package-README.md.in"
    "VERSION.json"         = "VERSION.json.in"
    "scripts/setup.ps1"    = "scripts\setup.ps1"
    "scripts/setup.sh"     = "scripts\setup.sh"
    "scripts/status.sh"    = "scripts\status.sh"
    "scripts/trial.sh"     = "scripts\trial.sh"
    "scripts/validate.ps1" = "scripts\validate.ps1"
    "scripts/validate.sh"  = "scripts\validate.sh"
}

foreach ($relative in $overlays.Keys) {
    $templatePath = Join-Path $TemplateRoot $overlays[$relative]
    $text = Get-Content -Raw -Encoding UTF8 -LiteralPath $templatePath
    $text = $text.Replace("@@SOURCE_COMMIT@@", $SourceCommit)
    $text = $text.Replace("@@SOURCE_TREE@@", $sourceTree)
    $text = $text.Replace("@@PACKAGE_VERSION@@", $PackageVersion)
    $text = $text.Replace("@@SOURCE_BRANCH@@", $SourceBranch)
    $text = $text.Replace("@@SOURCE_REPOSITORY@@", $SourceRepository)
    if ($text.Contains("@@")) {
        throw "Unresolved packaging token in $templatePath"
    }
    $files[$relative] = $utf8.GetBytes($text)
}

function Get-SHA256Hex([byte[]]$Bytes) {
    $sha = [System.Security.Cryptography.SHA256]::Create()
    try {
        return ([System.BitConverter]::ToString($sha.ComputeHash($Bytes))).Replace("-", "").ToLowerInvariant()
    }
    finally {
        $sha.Dispose()
    }
}

$manifestLines = foreach ($path in ($files.Keys | Sort-Object)) {
    "$(Get-SHA256Hex $files[$path])  $path"
}
$manifestBytes = $utf8.GetBytes(($manifestLines -join "`n") + "`n")
$files["MANIFEST.sha256"] = $manifestBytes

New-Item -ItemType Directory -Path $OutputDirectory -Force | Out-Null
$temporaryPath = "$archivePath.partial"
if (Test-Path -LiteralPath $temporaryPath) {
    throw "Partial output already exists; inspect it before retrying: $temporaryPath"
}

$archive = [System.IO.Compression.ZipFile]::Open(
    $temporaryPath,
    [System.IO.Compression.ZipArchiveMode]::Create
)
try {
    $root = "cyonic-node-setup-v$PackageVersion"
    $timestamp = [System.DateTimeOffset]::new(1980, 1, 1, 0, 0, 0, [System.TimeSpan]::Zero)
    foreach ($path in ($files.Keys | Sort-Object)) {
        $entry = $archive.CreateEntry(
            "$root/$($path.Replace('\', '/'))",
            [System.IO.Compression.CompressionLevel]::NoCompression
        )
        $entry.LastWriteTime = $timestamp
        $stream = $entry.Open()
        try {
            $bytes = $files[$path]
            $stream.Write($bytes, 0, $bytes.Length)
        }
        finally {
            $stream.Dispose()
        }
    }
}
finally {
    $archive.Dispose()
}

Move-Item -LiteralPath $temporaryPath -Destination $archivePath
$archiveHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $archivePath).Hash.ToLowerInvariant()

[pscustomobject]@{
    status          = "LOCAL_CANDIDATE_BUILT"
    package         = $archiveName
    packageVersion  = $PackageVersion
    sourceRepository = $SourceRepository
    sourceBranch    = $SourceBranch
    sourceCommit    = $SourceCommit
    sourceTree      = $sourceTree
    archiveRoot     = "sha256:$archiveHash"
    manifestRoot    = "sha256:$(Get-SHA256Hex $manifestBytes)"
    entryCount      = $files.Count
    authorityEffect = "NONE"
    publication     = "NOT_PERFORMED"
} | ConvertTo-Json
