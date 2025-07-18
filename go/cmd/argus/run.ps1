# argus 자동 실행 스크립트 (Windows PowerShell)
# OS와 아키텍처를 자동으로 감지하여 적절한 바이너리 실행

# 스크립트 디렉토리로 이동
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptPath

# 색상 함수
function Write-ColorOutput($ForegroundColor, $Message) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    Write-Output $Message
    $host.UI.RawUI.ForegroundColor = $fc
}

# 플랫폼 감지
Write-ColorOutput Yellow "🔍 플랫폼 감지 중..."

# Windows 바이너리 경로
$binaryPath = "bin\windows\argus.exe"

# 바이너리 존재 확인
if (-not (Test-Path $binaryPath)) {
    Write-ColorOutput Red "❌ 바이너리를 찾을 수 없습니다: $binaryPath"
    Write-ColorOutput Yellow "💡 먼저 빌드를 실행해주세요:"
    Write-Output "   .\build.sh (Git Bash/WSL에서)"
    Write-Output "   또는"
    Write-Output "   make build (Git Bash/WSL에서)"
    exit 1
}

Write-ColorOutput Green "✓ Windows 플랫폼 감지됨"
Write-ColorOutput Green "🚀 argus 실행"
Write-Output ""

# 바이너리 실행 (모든 인자 전달)
& $binaryPath $args
