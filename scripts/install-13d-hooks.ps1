$ErrorActionPreference = "Stop"

$repositoryRoot = (git rev-parse --show-toplevel).Trim()
if (-not $repositoryRoot) {
    throw "Not inside a Git repository."
}

git -C $repositoryRoot config core.hooksPath .githooks
Write-Host "Installed repository-local 13D hooks via core.hooksPath=.githooks"
Write-Host "authorityEffect: NONE"
