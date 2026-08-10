# VS Code folderOpen backdoor cleanup

## Already merged
- https://github.com/Barosz30/Barosz30/pull/1 (this profile repo is cleaned)

## Run on Windows (PowerShell)

1. Install **GitHub CLI** (this was the cause of `WinError 2`):
   ```powershell
   winget install --id GitHub.cli
   ```
   Or download: https://cli.github.com/

2. **Close and reopen** PowerShell (so `gh` is on PATH).

3. Log in and verify:
   ```powershell
   gh auth login
   gh auth status
   ```

4. Run the cleaner from the repo root (or from `scripts\`):
   ```powershell
   cd "$env:USERPROFILE\Desktop\Projekty\Nowy folder\Barosz30"
   python scripts\remove-vscode-backdoor.py
   ```

   Or only a few repos first:
   ```powershell
   python scripts\remove-vscode-backdoor.py Portfolio MechaShopFront Chatbot
   ```

## What it removes
- `public/fonts/fa-solid-400.woff2` (JS malware)
- `.vscode/tasks.json` (folderOpen executor)
- planted `.vscode/settings.json` / `launch.json` / kit files
- hardens `.gitignore` (`.vscode/`, `temp_auto_push.bat`, …)

## If you see `WinError 2` / „Nie można odnaleźć określonego pliku”
`gh` is not installed or not on PATH. Install GitHub CLI, open a **new** terminal, run `gh auth status`, then retry.

## Work / company repos
Copy this script and scan company orgs the same way — the IOCs are identical.
