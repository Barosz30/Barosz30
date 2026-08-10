#!/usr/bin/env python3
"""
Remove the VS Code folderOpen backdoor from Barosz30 repositories.

Requires:
  - GitHub CLI (`gh`) authenticated as a user with push access
  - git
  - python3

Windows:
  Install GitHub CLI from https://cli.github.com/
  Then in a NEW PowerShell:
    gh auth login
    python scripts/remove-vscode-backdoor.py

Usage:
  python scripts/remove-vscode-backdoor.py
  python scripts/remove-vscode-backdoor.py Portfolio MechaShopFront
  python scripts/remove-vscode-backdoor.py --include-gitignore-only
"""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

OWNER = "Barosz30"
BRANCH = "cursor/remove-vscode-backdoor-6880"
HERE = Path(__file__).resolve().parent
TARGETS_FILE = HERE / "infected-repos.json"

MALWARE_SETTINGS_MARKERS = (
    '"debug.openDebug": "neverOpen"',
    "compile-hero.disable-compile-files-on-did-save-code",
)
MALWARE_LAUNCH_MARKERS = ("Debug SST", "flo-ct-flo360", "sourceMapRenames")
MALWARE_TASK_MARKERS = ("fa-solid-400.woff2",)
GITIGNORE_IOC = {
    "branch_structure.json",
    "temp_auto_push.bat",
    "temp_interactive_push.bat",
}

PR_TITLE = "Remove VS Code folderOpen backdoor and malware payload"
PR_BODY = """## Summary
Removes a confirmed supply-chain backdoor:

1. **`.vscode/tasks.json`** — hidden task with `runOn: folderOpen` executing `node ./public/fonts/fa-solid-400.woff2`
2. **`public/fonts/fa-solid-400.woff2`** — JavaScript malware disguised as a font
3. Companion camouflage (`.vscode/settings.json` with `debug.openDebug: neverOpen`, planted `launch.json`, etc.)
4. Hardens `.gitignore` against `.vscode/` and known dropper filenames (`temp_auto_push.bat`, etc.)

## Test plan
- [ ] Payload file removed
- [ ] Malicious `.vscode/tasks.json` removed
- [ ] Opening the folder in VS Code/Cursor does not auto-run `node ...woff2`
"""

GH = None
GIT = None


def resolve_tool(name: str) -> str | None:
    """Find an executable on PATH (Windows-friendly)."""
    found = shutil.which(name)
    if found:
        return found
    if os.name == "nt":
        for candidate in (f"{name}.cmd", f"{name}.exe", f"{name}.bat"):
            found = shutil.which(candidate)
            if found:
                return found
    return None


def ensure_tools() -> None:
    global GH, GIT
    GIT = resolve_tool("git")
    GH = resolve_tool("gh")
    missing = []
    if not GIT:
        missing.append("git")
    if not GH:
        missing.append("gh (GitHub CLI)")
    if missing:
        print("ERROR: missing required tools:", ", ".join(missing), file=sys.stderr)
        print(file=sys.stderr)
        if "gh (GitHub CLI)" in missing:
            print("Install GitHub CLI:", file=sys.stderr)
            print("  https://cli.github.com/", file=sys.stderr)
            print("  winget install --id GitHub.cli", file=sys.stderr)
            print(file=sys.stderr)
            print("Then open a NEW PowerShell and run:", file=sys.stderr)
            print("  gh auth login", file=sys.stderr)
            print("  gh auth status", file=sys.stderr)
            print("  python scripts\\remove-vscode-backdoor.py", file=sys.stderr)
        sys.exit(2)

    # Verify gh can talk to GitHub
    try:
        status = run([GH, "auth", "status"], check=False)
    except FileNotFoundError:
        print("ERROR: found gh path but could not execute it:", GH, file=sys.stderr)
        sys.exit(2)
    if status.returncode != 0:
        print("ERROR: gh is installed but not logged in.", file=sys.stderr)
        print("Run:  gh auth login", file=sys.stderr)
        print((status.stderr or status.stdout or "").strip(), file=sys.stderr)
        sys.exit(2)


def run(cmd, cwd=None, check=True):
    try:
        return subprocess.run(
            cmd,
            cwd=cwd,
            check=check,
            text=True,
            capture_output=True,
            shell=False,
        )
    except FileNotFoundError as e:
        raise RuntimeError(
            f"Could not run {cmd[0]!r}. Is it installed and on PATH?\n{e}"
        ) from e


def file_contains(path: Path, markers) -> bool:
    if not path.is_file():
        return False
    text = path.read_text(encoding="utf-8", errors="replace")
    return any(m in text for m in markers)


def is_js_payload(path: Path) -> bool:
    if not path.is_file():
        return False
    data = path.read_bytes()
    return (b"require(" in data and (b"eth_" in data.lower() or b"spawn" in data)) or (
        b"global." in data and b"spawn" in data
    )


def clean_gitignore(path: Path) -> bool:
    if not path.is_file():
        path.write_text(
            ".vscode/\nbranch_structure.json\ntemp_auto_push.bat\ntemp_interactive_push.bat\n",
            encoding="utf-8",
        )
        return True

    original = path.read_text(encoding="utf-8", errors="replace")
    out = []
    changed = False
    for line in original.splitlines(True):
        s = line.strip()
        if s in GITIGNORE_IOC or s == ".gitignore":
            changed = True
            continue
        out.append(line)
    text = "".join(out)
    additions = []
    for entry in [".vscode/", "branch_structure.json", "temp_auto_push.bat", "temp_interactive_push.bat"]:
        if entry == ".vscode/" and re.search(r"^\.vscode/?$", text, flags=re.M):
            continue
        if not re.search(rf"^{re.escape(entry)}$", text, flags=re.M):
            additions.append(entry)
    if additions:
        if text and not text.endswith("\n"):
            text += "\n"
        text += "\n# Prevent reintroduction of VS Code folderOpen dropper artifacts\n"
        text += "\n".join(additions) + "\n"
        changed = True
    if changed:
        path.write_text(text, encoding="utf-8")
    return changed


def clean_repo_tree(root: Path) -> list[str]:
    changes: list[str] = []
    payload = root / "public" / "fonts" / "fa-solid-400.woff2"
    tasks = root / ".vscode" / "tasks.json"
    if payload.exists() and (is_js_payload(payload) or file_contains(tasks, MALWARE_TASK_MARKERS)):
        payload.unlink()
        changes.append("delete public/fonts/fa-solid-400.woff2")

    vscode = root / ".vscode"
    if vscode.is_dir():
        if file_contains(tasks, MALWARE_TASK_MARKERS):
            tasks.unlink()
            changes.append("delete .vscode/tasks.json")
        settings = vscode / "settings.json"
        if file_contains(settings, MALWARE_SETTINGS_MARKERS):
            settings.unlink()
            changes.append("delete .vscode/settings.json")
        launch = vscode / "launch.json"
        if file_contains(launch, MALWARE_LAUNCH_MARKERS):
            launch.unlink()
            changes.append("delete .vscode/launch.json")
        kit_deleted = any(c.startswith("delete .vscode/") for c in changes)
        if kit_deleted:
            for companion, needles in (
                ("extensions.json", ("plantuml", "prettier-vscode")),
                ("spellright.dict", ("Contentful", "Lerna")),
            ):
                p = vscode / companion
                if not p.exists():
                    continue
                txt = p.read_text(encoding="utf-8", errors="replace")
                if any(n.lower() in txt.lower() for n in needles):
                    p.unlink()
                    changes.append(f"delete .vscode/{companion}")
        try:
            if vscode.exists() and not any(vscode.iterdir()):
                vscode.rmdir()
                changes.append("delete empty .vscode/")
        except OSError:
            pass

    if clean_gitignore(root / ".gitignore"):
        changes.append("update .gitignore")
    return changes


def create_pr(name: str, default_branch: str) -> dict:
    compare = f"https://github.com/{OWNER}/{name}/compare/{default_branch}...{BRANCH}?expand=1"
    try:
        out = run(
            [
                GH,
                "pr",
                "create",
                "-R",
                f"{OWNER}/{name}",
                "--base",
                default_branch,
                "--head",
                BRANCH,
                "--title",
                PR_TITLE,
                "--body",
                PR_BODY,
                "--draft",
            ],
            check=False,
        )
    except RuntimeError as e:
        return {"ok": False, "error": str(e), "compare_url": compare}

    if out.returncode == 0:
        return {"ok": True, "url": out.stdout.strip()}

    err = (out.stderr or out.stdout or "").strip()
    low = err.lower()
    if "already exists" in low or "pull request already exists" in low:
        view = run(
            [GH, "pr", "list", "-R", f"{OWNER}/{name}", "--head", BRANCH, "--json", "url", "--limit", "1"],
            check=False,
        )
        if view.returncode == 0 and view.stdout.strip():
            rows = json.loads(view.stdout)
            if rows:
                return {"ok": True, "url": rows[0]["url"], "note": "already existed"}
    return {"ok": False, "error": err[-800:], "compare_url": compare}


def process(name: str, default_branch: str) -> dict:
    result = {"repo": name, "default_branch": default_branch}
    tmp = Path(tempfile.mkdtemp(prefix=f"clean-{name}-"))
    try:
        run(
            [
                GIT,
                "clone",
                "--depth",
                "1",
                "--branch",
                default_branch,
                f"https://github.com/{OWNER}/{name}.git",
                str(tmp / "repo"),
            ]
        )
        root = tmp / "repo"
        changes = clean_repo_tree(root)
        result["changes"] = changes
        if not changes:
            result.update({"ok": True, "skipped": True, "reason": "already clean"})
            return result
        run([GIT, "checkout", "-B", BRANCH], cwd=root)
        run([GIT, "add", "-A"], cwd=root)
        if not run([GIT, "status", "--porcelain"], cwd=root).stdout.strip():
            result.update({"ok": True, "skipped": True, "reason": "no staged changes"})
            return result
        run(
            [
                GIT,
                "-c",
                "user.email=security-cleanup@local",
                "-c",
                "user.name=Security Cleanup",
                "commit",
                "-m",
                "Remove VS Code folderOpen backdoor and malware payload\n\n"
                "Deletes malicious .vscode task kit, JS payload disguised as "
                "fa-solid-400.woff2, camouflage settings/launch files, and hardens .gitignore.",
            ],
            cwd=root,
        )
        run([GIT, "push", "-u", "origin", BRANCH, "--force"], cwd=root)
        pr = create_pr(name, default_branch)
        result.update(pr)
        result["branch"] = BRANCH
        return result
    except (subprocess.CalledProcessError, RuntimeError) as e:
        if isinstance(e, subprocess.CalledProcessError):
            msg = (e.stderr or e.stdout or str(e))[-1000:]
        else:
            msg = str(e)
        result.update({"ok": False, "error": msg})
        return result
    finally:
        shutil.rmtree(tmp, ignore_errors=True)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("repos", nargs="*", help="Optional repo names (default: all infected)")
    parser.add_argument("--gitignore-only", action="store_true", help="Only process gitignore-IOC repos")
    parser.add_argument(
        "--include-gitignore-only",
        action="store_true",
        help="Include gitignore-only targets too",
    )
    args = parser.parse_args()

    ensure_tools()
    print(f"Using git: {GIT}")
    print(f"Using gh:  {GH}")

    data = json.loads(TARGETS_FILE.read_text(encoding="utf-8"))
    if args.repos:
        wanted = set(args.repos)
        targets = [
            {
                "name": x["name"],
                "default_branch": x.get("default_branch") or x.get("defaultBranch") or "main",
            }
            for x in (data["infected"] + data["gitignore_only"])
            if x["name"] in wanted
        ]
        known = {t["name"] for t in targets}
        for name in args.repos:
            if name in known:
                continue
            meta = json.loads(run([GH, "api", f"repos/{OWNER}/{name}"]).stdout)
            targets.append({"name": name, "default_branch": meta.get("default_branch") or "main"})
    elif args.gitignore_only:
        targets = data["gitignore_only"]
    else:
        targets = list(data["infected"])
        if args.include_gitignore_only:
            targets.extend(data["gitignore_only"])

    results = []
    for i, t in enumerate(targets, 1):
        name = t["name"]
        default_branch = t.get("default_branch") or "main"
        print(f"[{i}/{len(targets)}] {name} @ {default_branch}", flush=True)
        res = process(name, default_branch)
        results.append(res)
        print(json.dumps(res, ensure_ascii=False)[:500], flush=True)

    out = HERE / "cleanup-results.json"
    out.write_text(json.dumps(results, indent=2), encoding="utf-8")
    ok = sum(1 for r in results if r.get("ok") and not r.get("skipped"))
    skipped = sum(1 for r in results if r.get("skipped"))
    failed = sum(1 for r in results if not r.get("ok"))
    print(f"\nDone. ok={ok} skipped={skipped} failed={failed}")
    print(f"Results: {out}")
    if failed:
        sys.exit(1)


if __name__ == "__main__":
    main()
