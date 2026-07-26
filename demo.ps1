# Logos-Formal: Run all demonstrators
# Requires: Go 1.24+
# Usage: .\demo.ps1

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host ""
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  Logos-Formal: Fail-Closed AI Governance Demonstrators" -ForegroundColor White
Write-Host "  https://github.com/XRastlinX/logos-formal" -ForegroundColor Gray
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host ""

# Check Go is available
try {
    $goVersion = go version
    Write-Host "Go version: $goVersion" -ForegroundColor DarkGray
} catch {
    Write-Host "Error: Go is not installed. Get it at https://go.dev/dl/" -ForegroundColor Red
    exit 1
}
Write-Host ""

# -- Demo 1: Permit-Gated Cell --
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "  DEMO 1: Permit-Gated Actuation" -ForegroundColor White
Write-Host "  Shows: AI proposes -> Validator checks -> Attester seals ->" -ForegroundColor Gray
Write-Host "         Human permits -> Actuator executes. Then attacks it." -ForegroundColor Gray
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host ""
Push-Location "$ScriptDir\runtime\demonstrator"
go run .
Pop-Location
Write-Host ""

# -- Demo 2: Bleed-Over Membrane --
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "  DEMO 2: Bleed-Over (B) Epistemic Membrane" -ForegroundColor White
Write-Host "  Shows: Cross-field translation preserving provenance." -ForegroundColor Gray
Write-Host "         Rejects certainty inflation and provenance erasure." -ForegroundColor Gray
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host ""
Push-Location "$ScriptDir\runtime\dcif-membrane"
go run .
Pop-Location
Write-Host ""

# -- Demo 3: Autophagic Decay --
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "  DEMO 3: Autophagic Decay (a) -- PROPOSED RESEARCH" -ForegroundColor White
Write-Host "  Shows: Certainty collapse when AI recursively ingests" -ForegroundColor Gray
Write-Host "         its own outputs. Forces reconnection to D=0 roots." -ForegroundColor Gray
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host ""
Push-Location "$ScriptDir\runtime\autophagic-decay"
go run .
Pop-Location
Write-Host ""

# -- Tests --
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "  RUNNING TEST SUITE" -ForegroundColor White
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host ""
Push-Location $ScriptDir
go test ./...
Pop-Location

Write-Host ""
Write-Host "All demonstrators complete. All tests passed." -ForegroundColor Green
Write-Host ""
Write-Host "Read more:" -ForegroundColor Cyan
Write-Host "  docs/WHY_THIS_MATTERS.md   - Plain-language explainer"
Write-Host "  docs/ARCHITECTURE.md       - Visual system diagrams"
Write-Host "  docs/POSITIONING.md        - Standards alignment (RATS, ABAC, SLSA)"
Write-Host ""
