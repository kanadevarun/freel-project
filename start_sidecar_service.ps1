# Stop existing process on port 8090 if any
$conns = Get-NetTCPConnection -LocalPort 8090 -State Listen -ErrorAction SilentlyContinue
if ($conns) {
    foreach ($c in $conns) {
        Stop-Process -Id $c.OwningProcess -Force -ErrorAction SilentlyContinue
    }
}

Write-Host "Starting Python AI Sidecar on port 8090..." -ForegroundColor Yellow
# NOTE: venv is at C:\Users\Sai\.venvs\freel-ai (outside repo — reduces IDE watcher load)
$venvPy = "C:\Users\Sai\.venvs\freel-ai\Scripts\python.exe"
$sidecarProc = Start-Process -FilePath $venvPy -ArgumentList "-m uvicorn main:app --host 127.0.0.1 --port 8090" -WorkingDirectory "c:\Users\Sai\go\src\freel-project\ai_sidecar" -RedirectStandardOutput "c:\Users\Sai\go\src\freel-project\ai_sidecar\sidecar_stdout.log" -RedirectStandardError "c:\Users\Sai\go\src\freel-project\ai_sidecar\sidecar_stderr.log" -PassThru

Write-Host "Verifying Python AI Sidecar port 8090..." -ForegroundColor Cyan
for ($i = 0; $i -lt 15; $i++) {
    Start-Sleep -Seconds 1
    if (Test-NetConnection -ComputerName 127.0.0.1 -Port 8090 -InformationLevel Quiet) {
        Write-Host "`nPython AI Sidecar is successfully running on http://127.0.0.1:8090 (PID: $($sidecarProc.Id))!" -ForegroundColor Green
        exit 0
    }
}

Write-Host "Sidecar startup logs:" -ForegroundColor Yellow
Get-Content "c:\Users\Sai\go\src\freel-project\ai_sidecar\sidecar_stdout.log" -ErrorAction SilentlyContinue
Get-Content "c:\Users\Sai\go\src\freel-project\ai_sidecar\sidecar_stderr.log" -ErrorAction SilentlyContinue
