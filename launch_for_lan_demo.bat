@echo off
title LogisticsHQ LAN Demo Server
color 0B
echo ======================================================================
echo           LOGISTICSHQ - HOST FOR SECOND MACHINE / LAN DEMO
echo ======================================================================
echo.

set HOST_IP=192.168.1.33
for /f "tokens=4" %%a in ('route print 0.0.0.0 ^| findstr 0.0.0.0 ^| findstr /v "Default"') do (
    set HOST_IP=%%a
    goto :ip_found
)
:ip_found

echo [1/4] Starting MySQL Database on port 3306...
start "LogisticsHQ MariaDB" /min "C:\Program Files\MariaDB 12.3\bin\mysqld.exe" --console
timeout /t 3 /nobreak >nul

echo [2/4] Starting Go Backend on port 8080 (0.0.0.0:8080)...
cd /d "c:\Users\Sai\go\src\freel-project\backend"
start "LogisticsHQ Go Backend" "c:\Users\Sai\go\src\freel-project\backend\server.exe"
timeout /t 2 /nobreak >nul

echo [3/4] Starting Python AI Sidecar on port 8090...
cd /d "c:\Users\Sai\go\src\freel-project\ai_sidecar"
start "LogisticsHQ Python Sidecar" /min "C:\Users\Sai\.venvs\freel-ai\Scripts\python.exe" -m uvicorn main:app --host 127.0.0.1 --port 8090
timeout /t 2 /nobreak >nul

echo [4/4] Starting Frontend Vite Server on port 5173 (0.0.0.0:5173)...
cd /d "c:\Users\Sai\go\src\freel-project\frontend"
start "LogisticsHQ Frontend Vite" cmd /k "npm run dev -- --host 0.0.0.0 --port 5173"
timeout /t 3 /nobreak >nul

cls
echo ======================================================================
echo                 LOGISTICSHQ DEMO SERVER IS LIVE!
echo ======================================================================
echo.
echo   YOUR SECOND MACHINE CAN NOW OPEN:
echo   ------------------------------------------------------------------
echo   * Web App:         http://%HOST_IP%:5173/
echo   * Direct Sign-in:  http://%HOST_IP%:5173/login
echo.
echo   DEMO CREDENTIALS:
echo   * Email:     kanadevarun123@gmail.com
echo   * Password:  Varun@123
echo.
echo   LOCAL STATUS:
echo   * Go Backend:      http://127.0.0.1:8080/health
echo   * Python AI:       http://127.0.0.1:8090/health
echo   * MariaDB:         127.0.0.1:3306
echo ======================================================================
echo (Keep this window open during your demo session)
echo.
pause
