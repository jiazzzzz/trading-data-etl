@echo off
echo Testing Go Server Setup...
echo.

REM Set proxy
set GOPROXY=https://goproxy.cn,direct

echo 1. Checking Go installation...
go version
if errorlevel 1 (
    echo ERROR: Go is not installed!
    pause
    exit /b 1
)
echo OK
echo.

echo 2. Checking database...
if not exist "..\jia-stk.db" (
    echo WARNING: Database not found at ..\jia-stk.db
    echo Please run: python dump.py first
    pause
    exit /b 1
)
echo OK
echo.

echo 3. Downloading Go modules...
go mod download
if errorlevel 1 (
    echo ERROR: Failed to download modules
    echo Try running: setup.bat
    pause
    exit /b 1
)
echo OK
echo.

echo 4. Building server...
go build -o stock-server.exe main.go
if errorlevel 1 (
    echo ERROR: Build failed
    pause
    exit /b 1
)
echo OK
echo.

echo ========================================
echo All tests passed!
echo ========================================
echo.
echo You can now start the server with:
echo   go run main.go
echo.
echo Or run the compiled binary:
echo   stock-server.exe
echo.
pause
