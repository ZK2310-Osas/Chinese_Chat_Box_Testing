@echo off
cd /d "%~dp0"
echo Starting Moonlit Dongpo - Gemini Edition...
where go >nul 2>&1
if errorlevel 1 (
  echo Go was not found. Install Go from https://go.dev/dl/
  pause
  exit /b 1
)
if not exist .env (
  echo Missing .env file. Copy .env.example to .env and add your Gemini key.
  pause
  exit /b 1
)
go run .
if errorlevel 1 (
  echo The application could not start. Review the error above.
  pause
)
