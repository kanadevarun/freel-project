# 1. Start MySQL Server
$mysqld = "C:\Program Files\MariaDB 12.3\bin\mysqld.exe"
$mysql = "C:\Program Files\MariaDB 12.3\bin\mysql.exe"

$mysqlListening = Test-NetConnection -ComputerName 127.0.0.1 -Port 3306 -InformationLevel Quiet
if (-not $mysqlListening) {
    Write-Host "Starting MySQL server..." -ForegroundColor Cyan
    Start-Process -FilePath $mysqld -ArgumentList "--console" -WindowStyle Hidden
    Start-Sleep -Seconds 4
}

# Verify MySQL port 3306
$mysqlListening = Test-NetConnection -ComputerName 127.0.0.1 -Port 3306 -InformationLevel Quiet
if ($mysqlListening) {
    Write-Host "MySQL on port 3306: ACTIVE" -ForegroundColor Green
    & $mysql -h 127.0.0.1 -u root -e "CREATE DATABASE IF NOT EXISTS freel_mysql;"
} else {
    Write-Host "MySQL on port 3306: FAILED" -ForegroundColor Red
}

# 2. Start Go Backend Server
Write-Host "Stopping existing Go Backend if running..." -ForegroundColor Cyan
$conns = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue
if ($conns) {
    foreach ($c in $conns) {
        Stop-Process -Id $c.OwningProcess -Force -ErrorAction SilentlyContinue
    }
    Start-Sleep -Seconds 1
}

Write-Host "Starting Go Backend..." -ForegroundColor Cyan
Set-Location "c:\Users\Sai\go\src\freel-project\backend"
$env:PATH = "C:\Program Files\Go\bin;" + $env:PATH

# Compile fresh server.exe
& "C:\Program Files\Go\bin\go.exe" build -o "c:\Users\Sai\go\src\freel-project\backend\server.exe" ./cmd/server

# Start server.exe in background
$backendProc = Start-Process -FilePath "c:\Users\Sai\go\src\freel-project\backend\server.exe" -WorkingDirectory "c:\Users\Sai\go\src\freel-project\backend" -RedirectStandardOutput "c:\Users\Sai\go\src\freel-project\backend\backend_stdout.log" -RedirectStandardError "c:\Users\Sai\go\src\freel-project\backend\backend_stderr.log" -PassThru

Start-Sleep -Seconds 3

# Verify Backend port 8080
$backendListening = Test-NetConnection -ComputerName 127.0.0.1 -Port 8080 -InformationLevel Quiet
if ($backendListening) {
    Write-Host "Go Backend on port 8080: ACTIVE (PID: $($backendProc.Id))" -ForegroundColor Green
    Write-Host "`nSUCCESS: Both MySQL (3306) and Go Backend (8080) are running!" -ForegroundColor Green
} else {
    Write-Host "Go Backend on port 8080: FAILED" -ForegroundColor Red
    Write-Host "`nBackend failed to start. Last stderr logs:" -ForegroundColor Red
    Get-Content "c:\Users\Sai\go\src\freel-project\backend\backend_stderr.log" -Tail 20
}
