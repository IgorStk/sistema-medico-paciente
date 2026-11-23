<#
.SYNOPSIS
  Roda os servicos Go (documentos e prontuario) direto no Windows, sem Docker,
  usando o Postgres e o Storage do Supabase configurados no arquivo .env.

.DESCRIPTION
  1. Le o .env da raiz do repositorio.
  2. Compila os dois servicos (os executaveis ficam fora do repositorio,
     em %LOCALAPPDATA%\sistema-medico-paciente\bin).
  3. Sobe os dois no mesmo terminal:
       documentos-service -> http://localhost:8081
       prontuario-service -> http://localhost:8082
  Ctrl+C encerra os dois.

.EXAMPLE
  .\scripts\rodar-supabase.ps1
#>
$ErrorActionPreference = 'Stop'
$raiz = Split-Path -Parent $PSScriptRoot

# ---------- 1. Le o .env ----------
$arquivoEnv = Join-Path $raiz '.env'
if (-not (Test-Path $arquivoEnv)) {
    throw "Arquivo .env nao encontrado. Copie .env.example para .env e preencha."
}
$vars = @{}
foreach ($linha in Get-Content $arquivoEnv -Encoding UTF8) {
    $l = $linha.Trim()
    if ($l -eq '' -or $l.StartsWith('#') -or -not $l.Contains('=')) { continue }
    $i = $l.IndexOf('=')
    $vars[$l.Substring(0, $i).Trim()] = $l.Substring($i + 1).Trim().Trim('"').Trim("'")
}
$obrigatorias = 'DOCUMENTOS_DATABASE_URL', 'DOCUMENTOS_S3_ENDPOINT', 'DOCUMENTOS_S3_REGION',
                'DOCUMENTOS_S3_ACCESS_KEY', 'DOCUMENTOS_S3_SECRET_KEY', 'PRONTUARIO_DATABASE_URL'
$faltando = $obrigatorias | Where-Object { -not $vars[$_] }
if ($faltando) { throw "Preencha no .env: $($faltando -join ', ')" }

foreach ($porta in 8081, 8082) {
    if (Get-NetTCPConnection -LocalPort $porta -State Listen -ErrorAction SilentlyContinue) {
        throw "A porta $porta ja esta em uso. Se forem os containers, rode: docker compose stop"
    }
}

# ---------- 2. Compila ----------
$bin = Join-Path $env:LOCALAPPDATA 'sistema-medico-paciente\bin'
New-Item -ItemType Directory -Force $bin | Out-Null
foreach ($servico in 'documentos-service', 'prontuario-service') {
    Write-Host "Compilando $servico..."
    Push-Location (Join-Path $raiz "services\$servico")
    try {
        go build -o (Join-Path $bin "$servico.exe") ./cmd/api
        if ($LASTEXITCODE -ne 0) { throw "Falha ao compilar $servico" }
    } finally { Pop-Location }
}

# ---------- 3. Sobe os servicos ----------
# Cada processo herda as variaveis definidas no momento em que e iniciado;
# depois elas sao restauradas para nao vazar de um servico para o outro.
function Start-ServicoGo([string]$nome, [hashtable]$ambiente) {
    $anteriores = @{}
    foreach ($k in $ambiente.Keys) {
        $anteriores[$k] = [Environment]::GetEnvironmentVariable($k)
        [Environment]::SetEnvironmentVariable($k, $ambiente[$k])
    }
    try {
        Start-Process -FilePath (Join-Path $bin "$nome.exe") -NoNewWindow -PassThru
    } finally {
        foreach ($k in $anteriores.Keys) { [Environment]::SetEnvironmentVariable($k, $anteriores[$k]) }
    }
}

$cors = if ($vars['CORS_ORIGINS']) { $vars['CORS_ORIGINS'] } else { 'http://localhost:5173,http://localhost:3000' }
$bucket = if ($vars['DOCUMENTOS_S3_BUCKET']) { $vars['DOCUMENTOS_S3_BUCKET'] } else { 'documentos' }

$documentos = Start-ServicoGo 'documentos-service' @{
    PORT          = '8081'
    DATABASE_URL  = $vars['DOCUMENTOS_DATABASE_URL']
    CORS_ORIGINS  = $cors
    S3_ENDPOINT   = $vars['DOCUMENTOS_S3_ENDPOINT']
    S3_REGION     = $vars['DOCUMENTOS_S3_REGION']
    S3_ACCESS_KEY = $vars['DOCUMENTOS_S3_ACCESS_KEY']
    S3_SECRET_KEY = $vars['DOCUMENTOS_S3_SECRET_KEY']
    S3_BUCKET     = $bucket
}
$prontuario = Start-ServicoGo 'prontuario-service' @{
    PORT                   = '8082'
    DATABASE_URL           = $vars['PRONTUARIO_DATABASE_URL']
    CORS_ORIGINS           = $cors
    DOCUMENTOS_SERVICE_URL = 'http://localhost:8081'
}

Write-Host ""
Write-Host "documentos-service -> http://localhost:8081/health"
Write-Host "prontuario-service -> http://localhost:8082/health"
Write-Host "Ctrl+C para parar os dois."
Write-Host ""

try {
    while (-not ($documentos.HasExited -or $prontuario.HasExited)) { Start-Sleep -Milliseconds 500 }
    $parou = if ($documentos.HasExited) { 'documentos-service' } else { 'prontuario-service' }
    Write-Warning "$parou parou (veja o erro acima). Encerrando o outro servico."
} finally {
    foreach ($p in $documentos, $prontuario) {
        if ($p -and -not $p.HasExited) { Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue }
    }
}
