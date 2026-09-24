@echo off
setlocal
cd /d %~dp0

echo [1/3] Building frontend...
cd web
call npm run build
if errorlevel 1 (echo Frontend build FAILED & exit /b 1)
cd ..

echo [2/3] Copying frontend dist to server/dist...
if exist server\dist rmdir /s /q server\dist
mkdir server\dist
xcopy web\dist server\dist /e /i /q >nul
if errorlevel 1 (echo Copy FAILED & exit /b 1)

echo [3/3] Building backend (embedding frontend)...
cd server
set CGO_ENABLED=0
go build -trimpath -ldflags "-s -w" -o resume-builder.exe .
if errorlevel 1 (echo Backend build FAILED & exit /b 1)
cd ..

echo.
echo Build OK: server\resume-builder.exe
echo Double-click to run. Data is stored in the "data" folder next to the exe.
