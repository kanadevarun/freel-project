@echo off
title LogisticsHQ SPortal - Dedicated Host Launcher
echo ========================================================
echo   LogisticsHQ SPortal - Live Hosting Engine
echo ========================================================
echo.

:: 1. Check and Start MariaDB
echo [1/4] Checking MariaDB Database on port 3306...
netstat -ano | findstr ":3306" | findstr "LISTENING" >nul
if %errorlevel% equ 0 (
    echo       MariaDB is already RUNNING.
) else (
    echo       Starting MariaDB...
    start "MariaDB Daemon" /min "C:\Program Files\MariaDB 12.3\bin\mysqld.exe" --console
    timeout /t 3 /nobreak >nul
)

:: 2. Check and Start Go Backend
echo [2/4] Checking Go Backend API on port 8080...
netstat -ano | findstr ":8080" | findstr "LISTENING" >nul
if %errorlevel% equ 0 (
    echo       Go Backend is already RUNNING.
) else (
    echo       Starting Go Backend server.exe...
    start "LogisticsHQ Backend" /min "%~dp0server.exe"
    timeout /t 3 /nobreak >nul
)

:: 3. Check and Start SPortal Frontend
echo [3/4] Checking SPortal Vite Server on port 5174...
netstat -ano | findstr ":5174" | findstr "LISTENING" >nul
if %errorlevel% equ 0 (
    echo       SPortal Frontend is already RUNNING.
) else (
    echo       Starting SPortal Frontend...
    cd /d "%~dp0sportal"
    start "SPortal Vite Server" /min cmd /c npm run dev
    cd /d "%~dp0"
    timeout /t 5 /nobreak >nul
)

:: 4. Start / Ensure Cloudflare Tunnel with HTTP/2 (Zero Disconnects)
echo [4/4] Starting Cloudflare Global Tunnel (HTTP/2 Protocol)...
echo.
echo ========================================================
echo   DO NOT CLOSE THIS WINDOW WHILE SHARING WITH FRIENDS
echo ========================================================
echo.
"%~dp0scratch\cloudflared.exe" tunnel --protocol http2 --edge-ip-version 4 --url http://localhost:5174
pause
