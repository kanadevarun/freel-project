# ==============================================================================
# Master Automated Setup and Launch Script for LogisticsHQ / Freel Project
# ==============================================================================
$ErrorActionPreference = "Continue"

Write-Host "=========================================================" -ForegroundColor Cyan
Write-Host "   LOGISTICSHQ: FULL-STACK SETUP & SERVICE ORCHESTRATION " -ForegroundColor Cyan
Write-Host "=========================================================" -ForegroundColor Cyan

# 1. Setup PATH
$env:PATH = "C:\Program Files\Go\bin;C:\Program Files\nodejs;C:\Users\Sai\AppData\Local\Programs\Python\Python311;C:\Users\Sai\AppData\Local\Programs\Python\Python311\Scripts;" + $env:PATH

# 2. Locate or Install MySQL / MariaDB
Write-Host "`n[1/6] Checking for MySQL / MariaDB installation..." -ForegroundColor Yellow

function Find-MySqlBin {
    $searchPaths = @(
        "C:\Program Files\MySQL\*\bin\mysql.exe",
        "C:\Program Files\MariaDB*\bin\mysql.exe",
        "C:\Program Files (x86)\MySQL\*\bin\mysql.exe",
        "C:\mysql*\bin\mysql.exe",
        "C:\Users\Sai\mysql*\bin\mysql.exe",
        "C:\Users\Sai\mariadb*\bin\mysql.exe"
    )
    foreach ($pattern in $searchPaths) {
        $found = Get-Item $pattern -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($found) { return $found.FullName }
    }
    $inPath = Get-Command "mysql.exe" -ErrorAction SilentlyContinue
    if ($inPath) { return $inPath.Source }
    return $null
}

$mysqlBin = Find-MySqlBin

if (-not $mysqlBin) {
    Write-Host "MySQL not found in common locations. Attempting install via winget..." -ForegroundColor Yellow
    winget install --id Oracle.MySQL --silent --accept-source-agreements --accept-package-agreements
    Start-Sleep -Seconds 3
    $mysqlBin = Find-MySqlBin
}

if (-not $mysqlBin) {
    Write-Host "Oracle.MySQL not found, trying MariaDB.Server via winget..." -ForegroundColor Yellow
    winget install --id MariaDB.Server --silent --accept-source-agreements --accept-package-agreements
    Start-Sleep -Seconds 3
    $mysqlBin = Find-MySqlBin
}

if (-not $mysqlBin) {
    Write-Host "Setting up portable MariaDB binary..." -ForegroundColor Yellow
    $zipUrl = "https://archive.mariadb.org/mariadb-11.4.3/winx64-packages/mariadb-11.4.3-winx64.zip"
    $destZip = "C:\Users\Sai\mariadb.zip"
    $destDir = "C:\Users\Sai\mariadb"
    
    if (-not (Test-Path $destDir)) {
        Write-Host "Downloading portable MariaDB (approx 90MB)..." -ForegroundColor Cyan
        Invoke-WebRequest -Uri $zipUrl -OutFile $destZip -UseBasicParsing
        Write-Host "Extracting MariaDB..." -ForegroundColor Cyan
        Expand-Archive -Path $destZip -DestinationPath "C:\Users\Sai" -Force
        $extracted = Get-ChildItem "C:\Users\Sai\mariadb-*" -Directory | Select-Object -First 1
        if ($extracted) {
            Rename-Item -Path $extracted.FullName -NewName "mariadb"
        }
    }
    $mysqlBin = "C:\Users\Sai\mariadb\bin\mysql.exe"
}

Write-Host "MySQL/MariaDB client located at: $mysqlBin" -ForegroundColor Green
$mysqlDir = Split-Path -Parent $mysqlBin
$mysqldBin = Join-Path $mysqlDir "mysqld.exe"
$env:PATH = $mysqlDir + ";" + $env:PATH

# 3. Ensure MySQL Service / Daemon is running
Write-Host "`n[2/6] Starting MySQL server..." -ForegroundColor Yellow

$port3306Open = Test-NetConnection -ComputerName 127.0.0.1 -Port 3306 -InformationLevel Quiet
if (-not $port3306Open) {
    # Check Windows services
    $svc = Get-Service -Name "*mysql*", "*mariadb*" -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($svc) {
        Write-Host "Starting service $($svc.Name)..." -ForegroundColor Cyan
        Start-Service $svc.Name -ErrorAction SilentlyContinue
    } else {
        Write-Host "Starting mysqld background process..." -ForegroundColor Cyan
        # Check if data directory initialized
        $dataDir = Join-Path (Split-Path -Parent $mysqlDir) "data"
        if (-not (Test-Path $dataDir)) {
            Write-Host "Initializing database data directory..." -ForegroundColor Cyan
            $initDbBin = Join-Path $mysqlDir "mysql_install_db.exe"
            if (Test-Path $initDbBin) {
                & $initDbBin --datadir="$dataDir"
            } else {
                & $mysqldBin --initialize-insecure --datadir="$dataDir"
            }
        }
        Start-Process -FilePath $mysqldBin -ArgumentList "--console" -WindowStyle Hidden
    }

    # Wait for port 3306
    for ($i = 0; $i -lt 30; $i++) {
        Start-Sleep -Seconds 1
        if (Test-NetConnection -ComputerName 127.0.0.1 -Port 3306 -InformationLevel Quiet) {
            Write-Host "MySQL server is active on port 3306!" -ForegroundColor Green
            break
        }
    }
} else {
    Write-Host "MySQL server is already active on port 3306!" -ForegroundColor Green
}

# 4. Provision Database and Load Existing Data
Write-Host "`n[3/6] Provisioning database 'freel_mysql' and seeding existing records..." -ForegroundColor Yellow
& $mysqlBin -h 127.0.0.1 -u root -e "CREATE DATABASE IF NOT EXISTS freel_mysql;"

Write-Host "Importing master schema up_mysql.sql..." -ForegroundColor Cyan
Get-Content "c:\Users\Sai\go\src\freel-project\backend\up_mysql.sql" -Raw | & $mysqlBin -h 127.0.0.1 -u root freel_mysql

Write-Host "Importing complete seed data (org, users, rfqs, quotes, shipments, rates)..." -ForegroundColor Cyan
Get-Content "c:\Users\Sai\go\src\freel-project\backend\seed_complete_mysql.sql" -Raw | & $mysqlBin -h 127.0.0.1 -u root freel_mysql

Write-Host "Importing contract records seed_contracts_polish.sql..." -ForegroundColor Cyan
Get-Content "c:\Users\Sai\go\src\freel-project\backend\scripts\seed_contracts_polish.sql" -Raw | & $mysqlBin -h 127.0.0.1 -u root freel_mysql

Write-Host "Database provisioned and seeded successfully!" -ForegroundColor Green

# 5. Setup Frontend Dependencies
Write-Host "`n[4/6] Installing Frontend dependencies (npm install)..." -ForegroundColor Yellow
Set-Location "c:\Users\Sai\go\src\freel-project\frontend"
& "C:\Program Files\nodejs\npm.cmd" install --no-audit --prefer-offline

# 6. Setup Python Sidecar Dependencies
# NOTE: venv is stored outside the repo at C:\Users\Sai\.venvs\freel-ai to reduce IDE watcher load.
Write-Host "`n[5/6] Setting up Python virtual environment and dependencies for AI Sidecar..." -ForegroundColor Yellow
Set-Location "c:\Users\Sai\go\src\freel-project\ai_sidecar"
$pyExe = "C:\Users\Sai\AppData\Local\Programs\Python\Python311\python.exe"
$venvRoot = "C:\Users\Sai\.venvs\freel-ai"

if (-not (Test-Path "$venvRoot\Scripts\python.exe")) {
    Write-Host "Creating Python virtualenv at $venvRoot..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path "C:\Users\Sai\.venvs" | Out-Null
    & $pyExe -m venv $venvRoot
}

$venvPy = "$venvRoot\Scripts\python.exe"
Write-Host "Installing requirements in ai_sidecar..." -ForegroundColor Cyan
& $venvPy -m pip install --upgrade pip --quiet
& $venvPy -m pip install -r requirements.txt --quiet

# 7. Start Go Backend, Python AI Sidecar, and Frontend Servers
Write-Host "`n[6/6] Starting all servers..." -ForegroundColor Yellow

# Kill existing servers on ports 8080, 8090, 5173 if running
foreach ($p in 8080, 8090, 5173) {
    $conns = Get-NetTCPConnection -LocalPort $p -State Listen -ErrorAction SilentlyContinue
    if ($conns) {
        foreach ($c in $conns) {
            Stop-Process -Id $c.OwningProcess -Force -ErrorAction SilentlyContinue
        }
    }
}

# Start Go Backend
Write-Host "Starting Go Backend on port 8080..." -ForegroundColor Cyan
Set-Location "c:\Users\Sai\go\src\freel-project\backend"
$backendProc = Start-Process -FilePath "C:\Program Files\Go\bin\go.exe" -ArgumentList "run cmd/server/main.go" -WorkingDirectory "c:\Users\Sai\go\src\freel-project\backend" -PassThru -WindowStyle Hidden

# Start Python AI Sidecar
Write-Host "Starting Python AI Sidecar on port 8090..." -ForegroundColor Cyan
Set-Location "c:\Users\Sai\go\src\freel-project\ai_sidecar"
$sidecarProc = Start-Process -FilePath $venvPy -ArgumentList "-m uvicorn main:app --host 127.0.0.1 --port 8090" -WorkingDirectory "c:\Users\Sai\go\src\freel-project\ai_sidecar" -PassThru -WindowStyle Hidden

# Start Frontend
Write-Host "Starting Frontend Vite Dev Server on port 5173..." -ForegroundColor Cyan
Set-Location "c:\Users\Sai\go\src\freel-project\frontend"
$frontendProc = Start-Process -FilePath "C:\Program Files\nodejs\npm.cmd" -ArgumentList "run dev -- --host 127.0.0.1" -WorkingDirectory "c:\Users\Sai\go\src\freel-project\frontend" -PassThru -WindowStyle Hidden

# 8. Verification & Polling
Write-Host "`nVerifying services..." -ForegroundColor Yellow
$maxWait = 30
$allUp = $false

for ($i = 0; $i -lt $maxWait; $i++) {
    Start-Sleep -Seconds 2
    $bUp = Test-NetConnection -ComputerName 127.0.0.1 -Port 8080 -InformationLevel Quiet
    $sUp = Test-NetConnection -ComputerName 127.0.0.1 -Port 8090 -InformationLevel Quiet
    $fUp = Test-NetConnection -ComputerName 127.0.0.1 -Port 5173 -InformationLevel Quiet
    
    if ($bUp -and $sUp -and $fUp) {
        $allUp = $true
        break
    }
    Write-Host "Waiting for servers to bind... [Go Backend: $bUp | AI Sidecar: $sUp | Frontend: $fUp]"
}

Write-Host "`n=========================================================" -ForegroundColor Green
Write-Host "   ALL LOGISTICSHQ SERVICES ARE RUNNING! " -ForegroundColor Green
Write-Host "=========================================================" -ForegroundColor Green
Write-Host "  Database:      MySQL 127.0.0.1:3306 (DB: freel_mysql)" -ForegroundColor White
Write-Host "  Go Backend:    http://127.0.0.1:8080 (PID: $($backendProc.Id))" -ForegroundColor White
Write-Host "  AI Sidecar:    http://127.0.0.1:8090 (PID: $($sidecarProc.Id))" -ForegroundColor White
Write-Host "  Frontend App:  http://127.0.0.1:5173 (PID: $($frontendProc.Id))" -ForegroundColor White
Write-Host "=========================================================" -ForegroundColor Green
