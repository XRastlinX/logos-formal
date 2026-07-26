$ErrorActionPreference = "Stop"

$RepoDir = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$FixtureDir = Join-Path ([System.IO.Path]::GetTempPath()) ("cyonic-client-" + [guid]::NewGuid().ToString("N"))
$Pushed = $false

try {
    Push-Location $RepoDir
    $Pushed = $true

    go run ./runtime/cyonic-service fixture -dir $FixtureDir
    if ($LASTEXITCODE -ne 0) {
        throw "Fixture generation failed with exit code $LASTEXITCODE."
    }

    go run ./runtime/cyonic-service evaluate `
        -request (Join-Path $FixtureDir "request.json") `
        -trust-key (Join-Path $FixtureDir "trusted-public.key") `
        -issuer example-principal `
        -origin external `
        -evidence-log (Join-Path $FixtureDir "interactions.jsonl")
    if ($LASTEXITCODE -ne 0) {
        throw "Boundary evaluation failed with exit code $LASTEXITCODE."
    }

    Write-Host ""
    Write-Host "Interaction receipt:" -ForegroundColor Cyan
    Get-Content (Join-Path $FixtureDir "interactions.jsonl")
    Write-Host ""
    Write-Host "Externality remains CLAIMED_EXTERNAL until independently adjudicated." -ForegroundColor Yellow
}
finally {
    if ($Pushed) {
        Pop-Location
    }
    if (Test-Path -LiteralPath $FixtureDir) {
        Remove-Item -LiteralPath $FixtureDir -Recurse -Force
    }
}
