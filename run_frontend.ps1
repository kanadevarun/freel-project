$env:PATH = "C:\Program Files\nodejs;" + $env:PATH
Set-Location "c:\Users\Sai\go\src\freel-project\frontend"
& "C:\Program Files\nodejs\npm.cmd" run dev -- --host 127.0.0.1
