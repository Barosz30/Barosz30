# VS Code folderOpen backdoor cleanup

## Already opened
- https://github.com/Barosz30/Barosz30/pull/1

## Why only one PR from the cloud agent?
The Cursor agent token can push only to `Barosz30/Barosz30`. The other ~76 infected repos need **your** GitHub login.

## Run locally (creates draft PRs for all infected repos)

```bash
gh auth login   # account Barosz30 with repo scope
cd /path/to/Barosz30   # this profile repo, on the cleanup branch or after merge
python3 scripts/remove-vscode-backdoor.py
```

Options:

```bash
# selected repos
python3 scripts/remove-vscode-backdoor.py Portfolio MechaShopFront Chatbot

# also repos that only have dropper names in .gitignore (no payload file)
python3 scripts/remove-vscode-backdoor.py --include-gitignore-only
```

## What it removes
- `public/fonts/fa-solid-400.woff2` (JS malware)
- `.vscode/tasks.json` (folderOpen executor)
- planted `.vscode/settings.json` / `launch.json` / kit files
- hardens `.gitignore` (`.vscode/`, `temp_auto_push.bat`, …)

## Work / company repos
Copy this script and scan company orgs the same way — the IOCs are identical.
