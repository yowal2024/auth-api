$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $PSScriptRoot
$nodeModulesPath = Join-Path $projectRoot 'frontend/node_modules'
$releaseDirectory = Join-Path $projectRoot 'dist'
$temporaryOutputPath = Join-Path $releaseDirectory ".auth-api-linux-amd64.$([guid]::NewGuid().ToString('N')).tmp"

$goCommands = @(& { Get-Command go.exe -ErrorAction SilentlyContinue } | ForEach-Object { $_.Source })
$goCandidates = @($goCommands) + @(
    (Join-Path $env:ProgramFiles 'Go/bin/go.exe'),
    'D:/environment/Go/bin/go.exe'
)
$goExecutable = $goCandidates | Where-Object { $_ -and (Test-Path -LiteralPath $_) } | Select-Object -First 1
if (-not $goExecutable) {
    throw 'Go is required. Install Go 1.27.3 or ensure go.exe is available on PATH.'
}

$npmCommand = Get-Command npm.cmd -ErrorAction SilentlyContinue
if (-not $npmCommand) {
    throw 'Node.js and npm are required to build the embedded management frontend.'
}

if (-not $env:GOPROXY) {
    $env:GOPROXY = 'https://goproxy.cn,https://goproxy.io,direct'
}

$previousGoOS = $env:GOOS
$previousGoArch = $env:GOARCH
$previousGoAMD64 = $env:GOAMD64
$previousCgoEnabled = $env:CGO_ENABLED
$previousGoFlags = $env:GOFLAGS

try {
    Push-Location (Join-Path $projectRoot 'frontend')
    try {
        if (-not (Test-Path -LiteralPath $nodeModulesPath -PathType Container)) {
            & $npmCommand.Source ci --no-audit --no-fund
            if ($LASTEXITCODE -ne 0) {
                throw 'npm ci failed.'
            }
        }
        & $npmCommand.Source run check
        if ($LASTEXITCODE -ne 0) {
            throw 'The frontend source checks failed.'
        }
        & $npmCommand.Source run build -- --outDir ../backend/internal/webassets/dist --emptyOutDir
        if ($LASTEXITCODE -ne 0) {
            throw 'The embedded frontend build failed.'
        }
    }
    finally {
        Pop-Location
    }

    $frontendBuildInfo = Get-Content -Raw -Encoding UTF8 (Join-Path $projectRoot 'backend/internal/webassets/dist/build-info.json') | ConvertFrom-Json
    $buildStamp = [DateTimeOffset]::Parse($frontendBuildInfo.builtAt).UtcDateTime.ToString('yyyyMMddTHHmmssZ')
    $outputPath = Join-Path $releaseDirectory "auth-api-linux-amd64-v$($frontendBuildInfo.version)-$buildStamp"
    if (Test-Path -LiteralPath $outputPath) {
        throw "Refusing to replace the existing release: $outputPath"
    }
    [System.IO.Directory]::CreateDirectory($releaseDirectory) | Out-Null

    Push-Location (Join-Path $projectRoot 'backend')
    try {
        $goVersionOutput = & $goExecutable version
        if ($LASTEXITCODE -ne 0 -or $goVersionOutput -notmatch 'go version go1\.27\.3(?:\s|$)') {
            throw "Expected Go 1.27.3; found $goVersionOutput"
        }

        & $goExecutable test -mod=readonly ./...
        if ($LASTEXITCODE -ne 0) {
            throw 'The Go test suite failed.'
        }

        & $goExecutable test -mod=readonly -tags=embeddedweb ./cmd/auth-api ./internal/webassets
        if ($LASTEXITCODE -ne 0) {
            throw 'The embedded frontend tests failed.'
        }

        $env:GOOS = 'linux'
        $env:GOARCH = 'amd64'
        $env:GOAMD64 = 'v1'
        $env:CGO_ENABLED = '0'
        $env:GOFLAGS = ''

        & $goExecutable build -mod=readonly -trimpath -tags=embeddedweb '-ldflags=-s -w' -o $temporaryOutputPath ./cmd/auth-api
        if ($LASTEXITCODE -ne 0) {
            throw 'The Linux amd64 build failed.'
        }

        $buildInfo = & $goExecutable version -m $temporaryOutputPath
        if ($LASTEXITCODE -ne 0) {
            throw 'Could not verify the Linux executable build metadata.'
        }
        $buildInfoText = $buildInfo -join "`n"
        foreach ($requiredBuildMarker in @('go1.27.3', 'GOOS=linux', 'GOARCH=amd64', 'CGO_ENABLED=0', 'tags=embeddedweb')) {
            if ($buildInfoText -notlike "*$requiredBuildMarker*") {
                throw "The executable is missing its expected Go or platform metadata: $requiredBuildMarker"
            }
        }

        [System.IO.File]::Move($temporaryOutputPath, $outputPath)
    }
    finally {
        Pop-Location
    }
}
finally {
    $env:GOOS = $previousGoOS
    $env:GOARCH = $previousGoArch
    $env:GOAMD64 = $previousGoAMD64
    $env:CGO_ENABLED = $previousCgoEnabled
    $env:GOFLAGS = $previousGoFlags
}

$artifact = Get-Item -LiteralPath $outputPath
$sha256 = (Get-FileHash -LiteralPath $outputPath -Algorithm SHA256).Hash.ToLowerInvariant()
Write-Output "Build information: v$($frontendBuildInfo.version) / $($frontendBuildInfo.builtAt)"
Write-Output "Linux amd64 executable: $($artifact.FullName) ($($artifact.Length) bytes)"
Write-Output "SHA-256: $sha256"
