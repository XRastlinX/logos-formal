[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string]$Target = ".",

    [string]$Manifest,

    [ValidateSet("text", "json")]
    [string]$Format = "text"
)

$ErrorActionPreference = "Stop"
$scriptDirectory = Split-Path -Parent $MyInvocation.MyCommand.Path
$repositoryRoot = [System.IO.Path]::GetFullPath((Join-Path $scriptDirectory ".."))
$temporaryBase = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath())
$temporaryDirectory = Join-Path $temporaryBase ("cyonic-validate-" + [guid]::NewGuid().ToString("N"))
$binary = Join-Path $temporaryDirectory "cyonic-validate.exe"
$validatorExitCode = 14

try {
    New-Item -ItemType Directory -Path $temporaryDirectory | Out-Null
    Push-Location $repositoryRoot
    try {
        & go build -o $binary ./cmd/cyonic-validate
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }

        $arguments = @($Target, "--format", $Format)
        if ($Manifest) {
            $arguments += @("--manifest", $Manifest)
        }
        & $binary @arguments
        $validatorExitCode = $LASTEXITCODE
    }
    finally {
        Pop-Location
    }
}
finally {
    $resolvedTemporaryDirectory = [System.IO.Path]::GetFullPath($temporaryDirectory)
    if ($resolvedTemporaryDirectory.StartsWith($temporaryBase, [System.StringComparison]::OrdinalIgnoreCase) -and
        (Test-Path -LiteralPath $resolvedTemporaryDirectory)) {
        Remove-Item -LiteralPath $resolvedTemporaryDirectory -Recurse -Force
    }
}

exit $validatorExitCode
