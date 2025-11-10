@echo off
echo Setting up Go environment...
echo.

REM Set Go proxy to China mirror (if needed)
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

echo Go proxy configured!
echo.
echo Downloading dependencies...
echo This may take a few minutes on first run...
go mod download

if errorlevel 1 (
    echo.
    echo ERROR: Failed to download dependencies
    echo Please check your internet connection
    pause
    exit /b 1
)

echo.
echo ========================================
echo Setup complete!
echo ========================================
echo.
echo You can now start the server with:
echo   go run main.go
echo.
echo Or build a binary:
echo   go build -o stock-server.exe main.go
echo.
pause
