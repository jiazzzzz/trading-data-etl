@echo off
echo Starting Frontend Server...
echo.
echo Frontend will be available at: http://localhost:3000
echo.
echo Make sure your API server is running at: http://localhost:8080
echo.
echo Starting Python HTTP server...
python -m http.server 3000
