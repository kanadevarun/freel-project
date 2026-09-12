@echo off
title LogisticsHQ Background Services
echo ========================================================
echo   LOGISTICSHQ: STARTING ALL SERVICES
echo ========================================================

echo [1/3] Starting MySQL on port 3306...
start "LogisticsHQ MySQL" /min "C:\Program Files\MariaDB 12.3\bin\mysqld.exe" --console
timeout /t 3 /nobreak >nul

echo [2/3] Starting Go Backend on port 8080...
cd /d "c:\Users\Sai\go\src\freel-project\backend"
start "LogisticsHQ Go Backend" "c:\Users\Sai\go\src\freel-project\backend\server.exe"
timeout /t 2 /nobreak >nul

echo [3/3] Starting Python AI Sidecar on port 8090...
cd /d "c:\Users\Sai\go\src\freel-project\ai_sidecar"
start "LogisticsHQ Python Sidecar" /min "C:\Users\Sai\.venvs\freel-ai\Scripts\python.exe" -m uvicorn main:app --host 127.0.0.1 --port 8090

echo.
echo ========================================================
echo   ALL SERVICES ARE ACTIVE!
echo ========================================================
echo   - MySQL:     127.0.0.1:3306
echo   - Backend:   http://127.0.0.1:8080
echo   - Sidecar:   http://127.0.0.1:8090
echo   - Frontend:  http://localhost:5173
echo.
echo You can now open http://localhost:5173/signup in your browser!
echo (Keep this window open or minimize it to keep services running)
echo ========================================================
pause
