@echo off
REM argus 자동 실행 스크립트 (Windows CMD)
REM Windows용 바이너리를 자동으로 실행

cd /d "%~dp0"

echo [33m🔍 플랫폼 감지 중...[0m

REM Windows 바이너리 경로
set BINARY_PATH=bin\windows\argus.exe

REM 바이너리 존재 확인
if not exist "%BINARY_PATH%" (
    echo [31m❌ 바이너리를 찾을 수 없습니다: %BINARY_PATH%[0m
    echo [33m💡 먼저 빌드를 실행해주세요:[0m
    echo    build.sh ^(Git Bash/WSL에서^)
    echo    또는
    echo    make build ^(Git Bash/WSL에서^)
    exit /b 1
)

echo [32m✓ Windows 플랫폼 감지됨[0m
echo [32m🚀 argus 실행[0m
echo.

REM 바이너리 실행 (모든 인자 전달)
"%BINARY_PATH%" %*
