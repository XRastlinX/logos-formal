param(
    [int]$Port = 18787
)

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$CanonDir = Join-Path $Root "canon\logos-formal"
$CacheDir = Join-Path $Root ".cache"
$EvidenceDir = Join-Path $Root "evidence"
$Executable = Join-Path $CacheDir "cyonic-service.exe"
$StdoutLog = Join-Path $CacheDir "server.stdout.log"
$StderrLog = Join-Path $CacheDir "server.stderr.log"
$BaseUrl = "http://127.0.0.1:$Port"
$InstanceId = "cyonic-validation-$([Guid]::NewGuid().ToString('N'))"
$Manifest = Get-Content -LiteralPath (Join-Path $Root "VERSION.json") -Raw | ConvertFrom-Json

if (-not (Test-Path -LiteralPath (Join-Path $CanonDir ".git"))) {
    throw "Canon checkout is missing. Run setup.cmd first."
}

$ActualCommit = (& git -C $CanonDir rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or $ActualCommit -ne $Manifest.sourceCommit) {
    throw "Canon checkout does not match the pinned source commit."
}
$ActualTree = (& git -C $CanonDir rev-parse "HEAD^{tree}").Trim()
if ($LASTEXITCODE -ne 0 -or $ActualTree -ne $Manifest.sourceTree) {
    throw "Canon checkout does not match the pinned source tree."
}
$Dirty = & git -C $CanonDir status --porcelain
if ($LASTEXITCODE -ne 0 -or $Dirty) {
    throw "Canon checkout has local changes; validation requires the pinned clean tree."
}

New-Item -ItemType Directory -Path $CacheDir -Force | Out-Null
New-Item -ItemType Directory -Path $EvidenceDir -Force | Out-Null

$RepositoryValidationOutput = & (Join-Path $CanonDir "scripts\validate.ps1") -Format json
if ($LASTEXITCODE -ne 0) {
    throw "Repository-bound validation rejected the pinned checkout."
}
$RepositoryValidation = ($RepositoryValidationOutput -join [Environment]::NewLine) | ConvertFrom-Json
if ($RepositoryValidation.validationStatus -ne "COMPLETE" -or
    $RepositoryValidation.status -ne "VALIDATED" -or
    $RepositoryValidation.cubedBit -ne "010" -or
    $RepositoryValidation.validatorOperator -ne "000" -or
    $RepositoryValidation.governanceAuthorityEffect -ne "NONE" -or
    $RepositoryValidation.executionContainment -ne "HOST" -or
    $RepositoryValidation.routerDecision -ne "OBSERVE_ONLY" -or
    $RepositoryValidation.checks.passed -ne 5 -or
    $RepositoryValidation.checks.total -ne 5) {
    throw "Repository validation output violated the node boundary."
}

Push-Location $CanonDir
try {
    & go build -o $Executable ./runtime/cyonic-service
    if ($LASTEXITCODE -ne 0) {
        throw "Could not build the Cyonic service."
    }
}
finally {
    Pop-Location
}

$Server = $null
try {
    $Server = Start-Process `
        -FilePath $Executable `
        -ArgumentList @(
            "serve-demo",
            "-listen", "127.0.0.1:$Port",
            "-instance-id", $InstanceId
        ) `
        -RedirectStandardOutput $StdoutLog `
        -RedirectStandardError $StderrLog `
        -PassThru `
        -WindowStyle Hidden

    Add-Type -AssemblyName System.Net.Http
    $HttpClient = [System.Net.Http.HttpClient]::new()
    $HttpClient.Timeout = [TimeSpan]::FromMilliseconds(500)
    $Ready = $false
    for ($attempt = 0; $attempt -lt 40; $attempt++) {
        try {
            $Health = $HttpClient.GetAsync("$BaseUrl/health").GetAwaiter().GetResult()
            if ([int]$Health.StatusCode -eq 200) {
                $HealthBody = $Health.Content.ReadAsStringAsync().GetAwaiter().GetResult() | ConvertFrom-Json
                if (-not $Server.HasExited -and $HealthBody.instanceId -eq $InstanceId) {
                    $Ready = $true
                    $Health.Dispose()
                    break
                }
            }
            $Health.Dispose()
        }
        catch {}
        if ($Server.HasExited) {
            throw "The service exited before it became ready. See $StderrLog"
        }
        Start-Sleep -Milliseconds 250
    }
    $HttpClient.Dispose()
    if (-not $Ready) {
        throw "The service did not become ready. See $StderrLog"
    }

    $ProbeOutput = & $Executable probe-http `
        -base-url $BaseUrl `
        -source-ref $Manifest.sourceCommit `
        -expected-instance $InstanceId
    if ($LASTEXITCODE -ne 0) {
        throw "The service did not pass its cold-call probe. See $StderrLog"
    }
    $Probe = ($ProbeOutput -join [Environment]::NewLine) | ConvertFrom-Json
    if ($Probe.status -ne "COLD_CALL_PASSED" -or
        $Probe.sourceRef -ne $Manifest.sourceCommit -or
        $Probe.serverInstance -ne $InstanceId -or
        $Probe.governanceState -ne "010" -or
        $Probe.authorityEffect -ne "NONE" -or
        $Probe.effect -ne "NOT_PERFORMED" -or
        $Probe.forwarded -ne $false -or
        $Probe.applyProbeStatus -ne 405 -or
        $Probe.applyProbeReason -ne "EFFECT_ROUTE_FORBIDDEN") {
        throw "The probe output did not satisfy the node boundary."
    }

    $Stamp = (Get-Date).ToUniversalTime().ToString("yyyyMMddTHHmmssZ")
    $RepositoryEvidencePath = Join-Path $EvidenceDir "repository-validation-$Stamp.json"
    $ServiceEvidencePath = Join-Path $EvidenceDir "service-probe-$Stamp.json"
    [System.IO.File]::WriteAllText(
        $RepositoryEvidencePath,
        (($RepositoryValidation | ConvertTo-Json -Depth 20) + [Environment]::NewLine),
        [System.Text.UTF8Encoding]::new($false)
    )
    [System.IO.File]::WriteAllText(
        $ServiceEvidencePath,
        (($Probe | ConvertTo-Json -Depth 20) + [Environment]::NewLine),
        [System.Text.UTF8Encoding]::new($false)
    )

    $RepositoryValidation | ConvertTo-Json -Depth 20
    $Probe | ConvertTo-Json -Depth 20
    Write-Host "Repository evidence: $RepositoryEvidencePath"
    Write-Host "Service evidence: $ServiceEvidencePath"
}
finally {
    if ($Server -and -not $Server.HasExited) {
        Stop-Process -Id $Server.Id -Force
        Wait-Process -Id $Server.Id -ErrorAction SilentlyContinue
    }
}
