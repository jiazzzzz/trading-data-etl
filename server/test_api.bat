@echo off
echo Testing Stock Data API Server...
echo.
echo Make sure the server is running (go run main.go)
echo.
timeout /t 2 /nobreak > nul

echo 1. Testing API root...
curl -s http://localhost:8080/ | jq .
echo.

echo 2. Testing tables endpoint...
curl -s http://localhost:8080/api/tables | jq .
echo.

echo 3. Testing stock list...
curl -s "http://localhost:8080/api/stocks?limit=3" | jq .
echo.

echo 4. Testing search...
curl -s "http://localhost:8080/api/search?q=银行" | jq .
echo.

echo Tests complete!
pause
