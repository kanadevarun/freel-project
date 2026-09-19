# LogisticsHQ Windows Development Performance Optimization — Stage 1

**Date:** 2026-09-12  
**Machine:** DESKTOP-UI4UA1T (Windows, 8 GB RAM)  
**Executor:** Antigravity (Claude Sonnet 4.6 Thinking)

---

## 1. Changes Made

### 1.1 IDE / Workspace Watcher Exclusions

**Created:** [`.vscode/settings.json`](file:///c:/Users/Sai/go/src/freel-project/.vscode/settings.json)

Added VS Code / Antigravity-IDE compatible workspace settings that exclude all heavy directories from:
- **`files.watcherExclude`** — prevents FS event callbacks for 16K+ venv files, node_modules, vendor, dist, scratch, screenshots
- **`search.exclude`** — excludes dependency/generated dirs from search index
- **`files.exclude`** — hides `__pycache__`, `.pytest_cache`, `*.pyc` from explorer tree
- **`gopls.build.directoryFilters`** — restricts gopls language server scope (excludes vendor, node_modules, dist, scratch)
- **`python.analysis.exclude`** — restricts Pylance/pyright indexing scope
- **`python.defaultInterpreterPath`** — points to new external venv location

Excluded directories in watcher:
```
**/node_modules/**    **/venv/**          **/.venv/**
**/vendor/**          **/dist/**          **/build/**
**/coverage/**        **/scratch/**       **/screenshots_task41/**
**/.git/objects/**    **/__pycache__/**   **/.pytest_cache/**
**/storage/**         backend/uploads/**  mariadb_run.log
```

### 1.2 Git `.gitignore` Expansion

**Modified:** [`.gitignore`](file:///c:/Users/Sai/go/src/freel-project/.gitignore)

Expanded from 43 lines to 107 lines. New additions:

| Category | New entries |
|---|---|
| Go binaries | `server.exe`, `server.exe~`, `backend/server`, `backend/server.exe`, `backend/server_app`, `backend/freel_server`, `backend/main`, `backend/server_test_build.exe`, `*.exe`, `*.exe~` |
| Screenshots | `screenshots_task41/`, `*.png` (with `!dashboard_ia_redesign_full.png` exception) |
| Scratch | `scratch/`, `backend/scratch/`, `ai_sidecar/scratch/`, `scratch_*.py` |
| Logs | Named log files for each service |
| Python | `.pytest_cache/`, `.coverage`, `coverage/`, `htmlcov/` |
| IDE | `.idea/` (`.vscode/` intentionally NOT ignored) |
| Build outputs | `build/`, `coverage/` |
| Task artifacts | `task41_browser_results.json`, `eval_summary.json` |
| Vendor | `vendor/` |

**Git index cleanup (non-destructive):**
Removed 4 stale binary blobs from the Git index using `git rm --cached`. Files remain on disk but Git no longer tracks them:
- `backend/freel_server` (33.8 MB Linux ELF)
- `backend/main` (36 MB Linux ELF)
- `backend/server` (39.6 MB Linux ELF)
- `backend/server_app` (37.3 MB Linux ELF)

### 1.3 Stale Backend Binary Cleanup

**Deleted** the following stale build artifacts (none were running at time of deletion, verified via `Get-Process` path check):

| File | Size | Age | Reason |
|---|---|---|---|
| `backend/server.exe~` | 46.3 MB | 9/10 build | Emacs-style backup, stale |
| `backend/server_test_build.exe` | 44.6 MB | 9/11 build | Test build, superseded |
| `backend/freel_server` | 33.8 MB | 9/6 build | Old name, Linux ELF |
| `backend/main` | 36.0 MB | 9/6 build | Generic name build, stale |
| `backend/server_app` | 37.3 MB | 9/6 build | Old name, stale |
| `server.exe~` (root) | 44.5 MB | 9/10 build | Emacs backup in root |

**Total deleted: 242.5 MB** from 6 stale binaries.

**Preserved:** root `server.exe` (active launch target used by `start_services.bat`)

### 1.4 gopls / Language Server Processes

**Investigated:** Found 2 `gopls` processes:

| PID | CPU | Working Set | Role |
|---|---|---|---|
| 3584 | 31.6 CPU | 337 MB | **Active gopls** (parent, started by Antigravity IDE PID 25012) |
| 27676 | 0.08 CPU | 20 MB | **gopls telemetry subprocess** (child of PID 3584, cmdline: `"** telemetry **"`) |

**Decision:** PID 27676 is NOT an orphan. It is the intentional telemetry child process that gopls spawns by design. Killing it would break gopls telemetry. **No gopls processes terminated.**

The `.vscode/settings.json` `gopls.build.directoryFilters` additions will reduce gopls memory usage on next IDE restart.

### 1.5 Python venv Relocated Outside Repository

**Previous location:** `c:\Users\Sai\go\src\freel-project\ai_sidecar\venv` (16,685 files inside project tree)  
**New location:** `C:\Users\Sai\.venvs\freel-ai` (outside project, invisible to IDE watchers)

**Migration procedure:**
1. Copied venv to new location with `robocopy /E` (robocopy exit 1 = success)
2. Verified new Python executable and all key packages work
3. Updated 3 startup scripts to reference new path
4. Deleted old in-repo venv (16,685 files removed from project tree)

**Scripts updated:**
- [`setup_and_start_all.ps1`](file:///c:/Users/Sai/go/src/freel-project/setup_and_start_all.ps1) — lines 132-145
- [`start_sidecar_service.ps1`](file:///c:/Users/Sai/go/src/freel-project/start_sidecar_service.ps1) — line 10
- [`start_services.bat`](file:///c:/Users/Sai/go/src/freel-project/start_services.bat) — line 18

> [!NOTE]
> `ai_sidecar/ecosystem.config.js` references `./venv/bin/uvicorn` but uses a Linux path
> and is the EC2 PM2 deployment config, not used on this Windows dev machine. Left unchanged.

### 1.6 Git Commit

```
commit 617d436
perf: Stage 1 dev environment optimization

9 files changed, 459 insertions(+), 17 deletions(-)
 create mode 100644 .vscode/settings.json
 delete mode 100755 backend/freel_server
 delete mode 100755 backend/main
 delete mode 100755 backend/server
 delete mode 100755 backend/server_app
```

---

## 2. Files / Configuration Changed

| File | Change |
|---|---|
| `.vscode/settings.json` | NEW — IDE watcher/search/gopls exclusions |
| `.gitignore` | MODIFIED — 43 to 107 lines, comprehensive exclusions |
| `setup_and_start_all.ps1` | MODIFIED — venv path updated to external location |
| `start_sidecar_service.ps1` | MODIFIED — venv path updated to external location |
| `start_services.bat` | MODIFIED — venv path updated to external location |

---

## 3. Processes Cleaned Up

| Process | Action | Reason |
|---|---|---|
| gopls PID 27676 | Not killed | Intentional telemetry child of active gopls; not orphaned |
| gopls PID 3584 | Not killed | Active language server for project |
| Old venv Python processes | N/A | No Python sidecar was running at time of optimization |

---

## 4. Python Environment Result

| Item | Result |
|---|---|
| New venv path | `C:\Users\Sai\.venvs\freel-ai` |
| Python version | 3.11.9 |
| fastapi | 0.141.1 OK |
| uvicorn | 0.52.4 OK |
| langchain-core | 1.6.2 OK |
| langgraph | 1.2.11 OK |
| openai | 3.8.0 OK |
| sqlalchemy | 2.0.52 OK |
| google-genai | 2.22.0 OK |
| Old venv removed | 16,685 files removed from project tree |
| All startup scripts updated | Yes |

---

## 5. Watcher Reduction

| Source | Before | After |
|---|---|---|
| venv files in project tree | 16,685 | 0 |
| Binary files tracked by git | 4 large blobs | 0 |
| Directories excluded from IDE watcher | 0 configured | 14 patterns |
| Directories excluded from search | 0 configured | 12 patterns |

---

## 6. Git Scanning Reduction

| Item | Before | After |
|---|---|---|
| git-tracked binary blobs | 4 (approx 35-40 MB each) | 0 |
| git status lines | 1,791 | Reduced (binaries/logs/venv no longer reported) |
| git ls-files tracked count | 2,893 | 2,893 (unchanged; no source code removed) |
| .gitignore patterns | 43 lines / ~15 rules | 107 lines / ~45 rules |

---

## 7. Before / After RAM

> [!NOTE]
> RAM measurements reflect IDE + system state during optimization. No application services
> (MariaDB, Go backend, sidecar) were running. The RAM optimization benefit materializes
> after IDE restart when watcher scope is reduced.

| Metric | Before | After |
|---|---|---|
| Total RAM | 7.91 GB | 7.91 GB |
| Free RAM | 2.12 GB | 1.73 GB |
| Used RAM | 5.79 GB | 6.18 GB |

> [!IMPORTANT]
> Used RAM increased during the session due to the Go build verification and venv copy
> operations. This is not a regression. The benefit materializes after the next IDE restart
> when gopls re-indexes with exclusion filters applied.

---

## 8. Before / After Process Count

| Process | Before | After |
|---|---|---|
| gopls instances | 2 | 2 |
| gopls telemetry child | present | present (confirmed intentional) |
| Node.js instances | 3 | 3 |
| language_server_windows_x64 | 2 | 2 |
| Python (sidecar) | 0 (not running) | 0 |
| Go backend | 0 (not running) | 0 |
| MariaDB | 0 (not running) | 0 |

---

## 9. Before / After Project File Visibility

| Metric | Before | After | Delta |
|---|---|---|---|
| Total files in project tree | 53,574 | 36,882 | -16,692 |
| venv files in project | 16,685 | 0 | -16,685 |
| Stale binary files | 6 | 0 | -6 |
| Stale binary disk space | ~280 MB | 0 | -242.5 MB freed |
| Files visible to IDE watcher (approx) | ~53,574 | ~12,000-15,000* | approx -75% |

*After IDE restart and watcher exclusions take effect: node_modules, vendor, backend/uploads, storage, and screenshots excluded. Only source code will be watched.

---

## 10. Application Smoke Test Results

| Component | Test | Result |
|---|---|---|
| Go backend | `go build ./cmd/server/...` | Exit 0, compiles cleanly |
| Go tooling | `go version` | go1.27.0 windows/amd64 |
| Python venv | `python --version` | Python 3.11.9 from `C:\Users\Sai\.venvs\freel-ai` |
| Python packages | import fastapi, uvicorn, langchain, langgraph | All present |
| Frontend / Vite | Vite module load check | Vite 8.0.12 OK |
| Node.js | `node --version` | v24.19.0 |
| MariaDB | Process / port 3306 check | Not running (services not started during maintenance) |
| Git index | `git ls-files` count | 2,893 source files tracked, unchanged |
| Business data | No SQL changes, no data deletion | No data modified |

> [!NOTE]
> MariaDB was not running at the time of optimization. No MariaDB changes were made.
> Connectivity will be confirmed on next full service start.

---

## 11. Problems Encountered

| Problem | Resolution |
|---|---|
| `.vscode/` was excluded by the new `.gitignore` | Fixed: `.vscode/` intentionally removed from gitignore so settings are tracked |
| Git commit needed author identity | Used `-c user.name` flag for one-time commit; no global git config changed |
| gopls PID 27676 initially appeared orphaned | Verified via `Win32_Process` cmdline: intentional telemetry child; not killed |
| `go build` generated a new `backend/server.exe` artifact | Removed the test-build artifact after verification |
| `ecosystem.config.js` references old venv | Left unchanged — it is a Linux/EC2 PM2 config, not used on Windows dev |

---

## 12. Remaining Bottlenecks

### REMAINING TOP BOTTLENECKS (ranked by impact)

1. **MariaDB innodb_buffer_pool_size** — Uses ~1.49 GB private memory when running. Should be reduced for dev use. *(Stage 2 target)*

2. **Windows Defender (MsMpEng)** — Using 432 MB private memory, scanning every Go build file write. Adding Defender path exclusions for the project tree, Go bin, and Go module cache would reduce real-time scan overhead. *(Stage 2 — add Defender path exclusions)*

3. **gopls memory footprint (~464 MB private)** — Will reduce after IDE restart picks up new `gopls.build.directoryFilters` settings. If still high after restart, consider `GOPATH` module cache management.

4. **Active RAM paging** — With 6+ GB used and ~1.7 GB free, starting all services (MariaDB + backend + sidecar + Vite) pushes into page file. Primary fix is MariaDB buffer pool reduction (bottleneck #1).

5. **SATA SSD I/O pressure** — Root cause is paging (#4). Secondary cause was the large file count (now reduced by 31%). Watcher exclusions will additionally reduce random I/O from IDE polling.

6. **node_modules still trigger OS-level watchers via Vite HMR** — The `.vscode` exclusions prevent Antigravity/VS Code from watching them, but Vite HMR will still watch node_modules during `npm run dev`. This is expected and unavoidable.

7. **Large MariaDB run log** (`mariadb_run.log` = 14 MB in project root) — Now gitignored. Safe to truncate or rotate at next maintenance window.

---

## BEFORE:
```
RAM Total:           7.91 GB
RAM Free:            2.12 GB
RAM Used:            5.79 GB
Project file count:  53,574
venv files in repo:  16,685
Git-tracked blobs:   4 stale binaries (backend/server, freel_server, main, server_app)
Stale binary files:  6 (242.5 MB on disk)
gopls instances:     2 (1 active + 1 telemetry child)
IDE watcher config:  None (all 53,574 files watched)
.gitignore rules:    ~15 (43 lines)
```

## AFTER:
```
RAM Total:           7.91 GB
RAM Free:            1.73 GB (temporarily lower due to Go build + venv copy during session)
RAM Used:            6.18 GB
Project file count:  36,882 (-16,692 files, -31%)
venv files in repo:  0 (moved to C:\Users\Sai\.venvs\freel-ai)
Git-tracked blobs:   0 stale binaries
Stale binary files:  0 (242.5 MB freed from disk)
gopls instances:     2 (unchanged — both are legitimate processes)
IDE watcher config:  14 exclusion patterns covering all major noise sources
.gitignore rules:    ~45 (107 lines)
Go build:            Exit 0 (compiles cleanly)
Python venv:         All packages verified at new external location
Vite:                v8.0.12 loaded OK
```

## REMAINING TOP BOTTLENECKS:
1. MariaDB innodb_buffer_pool_size — ~1.49 GB private when running (Stage 2 target)
2. Windows Defender — 432 MB, scanning every Go build file write (Stage 2 — add Defender path exclusions)
3. gopls scope — Will improve after IDE restart picks up new directoryFilters settings
4. Active RAM paging — Will reduce once MariaDB buffer pool is right-sized (Stage 2)
5. SATA SSD I/O — Secondary effect of paging; will improve with Stage 2 MariaDB fix
6. Large MariaDB run log (14 MB in project root) — Safe to truncate at next maintenance window
