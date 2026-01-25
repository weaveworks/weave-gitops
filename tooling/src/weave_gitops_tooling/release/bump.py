"""Bump version from charts/gitops-server/Chart.yaml; update Chart, values, package.json.

Source of truth: charts/gitops-server/Chart.yaml `version`.

Bump types:
  - patch|minor|major: strip -rc.N, bump X.Y.Z (e.g. 0.39.0-rc.2 --minor--> 0.40.0).
  - rc: keep base, X.Y.Z-rc.(N+1) or X.Y.Z-rc.1 (e.g. 0.39.0-rc.2 -> 0.39.0-rc.3).
  - patch-rc|minor-rc|major-rc: bump base then add -rc.1 (e.g. 0.39.x --minor-rc--> 0.40.0-rc.1).

Writes: Chart.yaml (version, appVersion), values.yaml (image.tag), package.json (version).
Tags use v prefix (workflow adds 'v'). Appends version=... to GITHUB_OUTPUT when set.
"""

from __future__ import annotations

import os
import re
import sys
from pathlib import Path


def _read_current(chart_yaml: Path) -> tuple[str, int | None]:
    """Read base X.Y.Z and optional -rc.N from Chart.yaml. Returns (base, rc_num or None). Raises on parse error."""
    text = chart_yaml.read_text()
    # version: 0.39.0-rc.2 or version: 0.39.0
    m = re.search(r"^\s*version:\s*v?(\d+\.\d+\.\d+)(?:-rc\.(\d+))?", text, re.MULTILINE)
    if not m:
        msg = "Could not find version in charts/gitops-server/Chart.yaml"
        raise SystemExit(msg)
    base = m.group(1)
    rc = int(m.group(2)) if m.group(2) else None
    return (base, rc)


def _next_version(old: str, bump: str, *, rc_num: int | None = None) -> str:
    """Compute next version. old is X.Y.Z or vX.Y.Z. bump: patch|minor|major|rc|patch-rc|minor-rc|major-rc.
    For rc: keeps base, X.Y.Z-rc.(N+1) or X.Y.Z-rc.1. For *-rc: bump base then -rc.1."""
    b = (bump or "patch").lower()
    if b == "rc":
        old = old.lstrip("v")
        parts = old.split(".")
        if len(parts) != 3 or not all(p.isdigit() for p in parts):
            msg = f"Invalid version in Chart.yaml: {old}"
            raise SystemExit(msg)
        n = (rc_num + 1) if rc_num is not None else 1
        return f"{old}-rc.{n}"
    # patch-rc, minor-rc, major-rc: bump base then append -rc.1 (ignore rc_num)
    if b in ("patch-rc", "minor-rc", "major-rc"):
        base = old.lstrip("v").split("-")[0]  # 0.39.0 or 0.39.0 from 0.39.0-rc.2
        base = _next_version(base, b.replace("-rc", ""), rc_num=None)
        return f"{base}-rc.1"
    # patch, minor, major: ignore rc_num
    old = old.lstrip("v")
    parts = old.split(".")
    if len(parts) != 3 or not all(p.isdigit() for p in parts):
        msg = f"Invalid version in Chart.yaml: {old}"
        raise SystemExit(msg)
    x, y, z = int(parts[0]), int(parts[1]), int(parts[2])
    if b == "patch":
        z += 1
    elif b == "minor":
        y += 1
        z = 0
    elif b == "major":
        x += 1
        y = z = 0
    else:
        msg = f"Unknown bump: {bump}. Use patch, minor, major, rc, patch-rc, minor-rc, or major-rc."
        raise SystemExit(msg)
    return f"{x}.{y}.{z}"


def _update_chart_yaml(path: Path, old_base: str, new: str) -> bool:
    """Update version and appVersion in Chart.yaml. old_base is X.Y.Z we're replacing (may have had -rc)."""
    text = path.read_text()
    changed = False
    # version: 0.39.0-rc.2 ... -> version: X.Y.Z
    pat_ver = re.compile(
        r"^(\s*version:\s*)v?" + re.escape(old_base) + r"(?:-[^\s#]*)?(\s*(?:#.*)?)$",
        re.MULTILINE,
    )
    if pat_ver.search(text):
        text = pat_ver.sub(rf"\g<1>{new}\2", text, count=1)
        changed = True
    # appVersion: "v0.39.0-rc.2" ... -> appVersion: "vX.Y.Z"
    pat_app = re.compile(
        r'^(\s*appVersion:\s*")v?' + re.escape(old_base) + r'(?:-[^\"]*)?("[\s#]*(?:#.*)?)$',
        re.MULTILINE,
    )
    if pat_app.search(text):
        text = pat_app.sub(rf"\g<1>v{new}\2", text, count=1)
        changed = True
    if changed:
        path.write_text(text)
    return changed


def _update_values_yaml(path: Path, old_base: str, new: str) -> bool:
    """Update image.tag in values.yaml."""
    text = path.read_text()
    # tag: "v0.39.0-rc.2" ... -> tag: "vX.Y.Z"
    pat = re.compile(
        r'^(\s*tag:\s*")v?' + re.escape(old_base) + r'(?:-[^\"]*)?("[\s#]*(?:#.*)?)$',
        re.MULTILINE,
    )
    if not pat.search(text):
        return False
    text = pat.sub(rf"\g<1>v{new}\2", text, count=1)
    path.write_text(text)
    return True


def _update_package_json(path: Path, old_base: str, new: str) -> bool:
    """Update "version" in package.json. old_base is X.Y.Z; we match X.Y.Z or X.Y.Z-rc.N."""
    text = path.read_text()
    # "version": "0.39.0-rc.2" or "version": "0.39.0"
    pat = re.compile(r'("version"\s*:\s*")v?' + re.escape(old_base) + r'(?:-[^"]*)?(")')
    if not pat.search(text):
        return False
    text = pat.sub(rf"\g<1>{new}\2", text, count=1)
    path.write_text(text)
    return True


def run(project_root: Path, bump: str) -> int:
    """Bump: read from Chart.yaml, update Chart, values, package.json. Return 0 or 1."""
    chart = project_root / "charts" / "gitops-server" / "Chart.yaml"
    values = project_root / "charts" / "gitops-server" / "values.yaml"
    pkg = project_root / "package.json"

    if not chart.is_file():
        print("charts/gitops-server/Chart.yaml not found", file=sys.stderr)
        return 1

    old_base, rc_num = _read_current(chart)
    new = _next_version(old_base, bump, rc_num=rc_num)

    updated: list[Path] = []
    if _update_chart_yaml(chart, old_base, new):
        updated.append(chart.relative_to(project_root))
    if values.is_file() and _update_values_yaml(values, old_base, new):
        updated.append(values.relative_to(project_root))
    if pkg.is_file() and _update_package_json(pkg, old_base, new):
        updated.append(pkg.relative_to(project_root))

    if not updated:
        print(
            f"No files had version {old_base!r} to update (Chart, values, package.json).",
            file=sys.stderr,
        )
        return 1

    print(f"Bumped {old_base} -> {new} ({bump}); updated {len(updated)} file(s)")
    for u in updated:
        print(f"  {u}")

    go = os.environ.get("GITHUB_OUTPUT")
    if go:
        with Path(go).open("a") as f:
            f.write(f"version={new}\n")

    return 0
