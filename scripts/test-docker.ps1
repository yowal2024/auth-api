$ErrorActionPreference = 'Stop'
$project = Split-Path -Parent $PSScriptRoot
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    $dockerBin = 'C:/Program Files/Docker/Docker/resources/bin'
    if (Test-Path (Join-Path $dockerBin 'docker.exe')) { $env:PATH = $dockerBin + ';' + $env:PATH }
}
Push-Location $project
try {
    docker compose -p auth-api-verification --profile test up -d --wait mysql redis-test
    if ($LASTEXITCODE -ne 0) { throw 'Test dependencies failed' }
    docker compose -p auth-api-verification --profile test run --build --rm test
    if ($LASTEXITCODE -ne 0) { throw 'Container tests failed' }
} finally {
    docker compose -p auth-api-verification --profile test down
    Pop-Location
}
