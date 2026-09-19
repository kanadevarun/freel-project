# Ensure MySQL and Go Backend are up
$mysqlBin = "C:\Program Files\MariaDB 12.3\bin\mysql.exe"
$mysqldBin = "C:\Program Files\MariaDB 12.3\bin\mysqld.exe"

Write-Host "Checking MySQL on port 3306..." -ForegroundColor Yellow
$mysqlUp = Test-NetConnection -ComputerName 127.0.0.1 -Port 3306 -InformationLevel Quiet

if (-not $mysqlUp) {
    Write-Host "Starting MySQL server..." -ForegroundColor Cyan
    Start-Process -FilePath $mysqldBin -ArgumentList "--console" -WindowStyle Hidden
    for ($i = 0; $i -lt 15; $i++) {
        Start-Sleep -Seconds 1
        if (Test-NetConnection -ComputerName 127.0.0.1 -Port 3306 -InformationLevel Quiet) {
            $mysqlUp = $true
            break
        }
    }
}

if ($mysqlUp) {
    Write-Host "MySQL is active on port 3306." -ForegroundColor Green
    & $mysqlBin -h 127.0.0.1 -u root -e "CREATE DATABASE IF NOT EXISTS freel_mysql;"
    
    # Check if tables exist
    $tables = & $mysqlBin -h 127.0.0.1 -u root -D freel_mysql -e "SHOW TABLES;"
    if (-not $tables -or $tables.Count -lt 5) {
        Write-Host "Seeding database freel_mysql..." -ForegroundColor Cyan
        Get-Content "c:\Users\Sai\go\src\freel-project\backend\up_mysql.sql" -Raw | & $mysqlBin -h 127.0.0.1 -u root freel_mysql
        Get-Content "c:\Users\Sai\go\src\freel-project\backend\seed_complete_mysql.sql" -Raw | & $mysqlBin -h 127.0.0.1 -u root freel_mysql
    }
} else {
    Write-Host "Warning: Could not start MySQL automatically." -ForegroundColor Red
}

# Stop any existing process on port 8080
$conns = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue
if ($conns) {
    foreach ($c in $conns) {
        Stop-Process -Id $c.OwningProcess -Force -ErrorAction SilentlyContinue
    }
}

Write-Host "Starting Go Backend on port 8080..." -ForegroundColor Yellow
Set-Location "c:\Users\Sai\go\src\freel-project\backend"
$env:PATH = "C:\Program Files\Go\bin;" + $env:PATH

# If server.exe already built, run it, otherwise run go run -mod=mod
if (Test-Path "c:\Users\Sai\go\src\freel-project\backend\server.exe") {
    $backendProc = Start-Process -FilePath "c:\Users\Sai\go\src\freel-project\backend\server.exe" -WorkingDirectory "c:\Users\Sai\go\src\freel-project\backend" -RedirectStandardOutput "c:\Users\Sai\go\src\freel-project\backend\backend_stdout.log" -RedirectStandardError "c:\Users\Sai\go\src\freel-project\backend\backend_stderr.log" -PassThru
} else {
    $backendProc = Start-Process -FilePath "C:\Program Files\Go\bin\go.exe" -ArgumentList "run -mod=mod cmd/server/main.go" -WorkingDirectory "c:\Users\Sai\go\src\freel-project\backend" -RedirectStandardOutput "c:\Users\Sai\go\src\freel-project\backend\backend_stdout.log" -RedirectStandardError "c:\Users\Sai\go\src\freel-project\backend\backend_stderr.log" -PassThru
}

Write-Host "Verifying Go Backend port 8080..." -ForegroundColor Cyan
for ($i = 0; $i -lt 15; $i++) {
    Start-Sleep -Seconds 1
    if (Test-NetConnection -ComputerName 127.0.0.1 -Port 8080 -InformationLevel Quiet) {
        Write-Host "`nGo Backend is successfully running on http://127.0.0.1:8080 (PID: $($backendProc.Id))!" -ForegroundColor Green
        exit 0
    }
}

Write-Host "Go Backend startup logs:" -ForegroundColor Yellow
Get-Content "c:\Users\Sai\go\src\freel-project\backend\backend_stdout.log" -ErrorAction SilentlyContinue
Get-Content "c:\Users\Sai\go\src\freel-project\backend\backend_stderr.log" -ErrorAction SilentlyContinue
