@echo off
echo Starting Stock Data API Server...
echo.

REM Set Go proxy for China (if needed)
set GOPROXY=https://goproxy.cn,direct

echo Server will be available at: http://localhost:8080
echo Web UI available at: server/index.html
echo.
echo Starting server...
go run main.go
