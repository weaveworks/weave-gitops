"""Generate release notes from commits since last tag using OpenAI or Anthropic. For GitHub Release body."""

from __future__ import annotations

import json
import os
import re
import ssl
import subprocess
import sys
import urllib.error
import urllib.request
from pathlib import Path

import certifi

DEFAULT_TEMPLATE = """# Release v{{VERSION}}

## Summary
[2-3 sentence overview of this release]

## Changes

### Features
- [List enhancements and new features from commits]

### Fixes
- [List bug fixes from commits]

### Other
- [Docs, chore, refactor, build, ci, etc.]

---
*Generated from commits since the previous release.*"""


def _get_previous_tag(project_root: Path) -> str | None:
    """Return the most recent tag that is an ancestor of HEAD, or None if none."""
    try:
        out = subprocess.run(
            ["git", "describe", "--tags", "--abbrev=0"],
            cwd=project_root,
            capture_output=True,
            text=True,
            check=True,
        )
        return out.stdout.strip() or None
    except subprocess.CalledProcessError:
        return None


def _get_commits_since(project_root: Path, ref: str) -> list[str]:
    """Return list of commit subject lines from ref..HEAD (excluding ref, including HEAD)."""
    try:
        out = subprocess.run(
            ["git", "log", f"{ref}..HEAD", "--pretty=format:%s"],
            cwd=project_root,
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        hint = f" {e.stderr.strip()}" if (e.stderr and e.stderr.strip()) else ""
        print(
            f"Invalid git ref for --since-tag: {ref!r}. Check the tag or commit exists.{hint}",
            file=sys.stderr,
        )
        raise SystemExit(1) from e
    if not out.stdout.strip():
        return []
    return [s for s in out.stdout.strip().split("\n") if s]


def _load_template(path: Path | None) -> str:
    if path is not None and path.is_file():
        return path.read_text()
    return DEFAULT_TEMPLATE


def _call_openai(commits: list[str], format_instructions: str, version: str, model: str) -> str:
    """Call OpenAI chat completions. Returns the generated markdown."""
    key = os.environ.get("OPENAI_API_KEY", "").strip()
    if not key:
        print(
            "OPENAI_API_KEY is not set. Add it as a repository secret for release notes.",
            file=sys.stderr,
        )
        raise SystemExit(1)

    user_content = (
        f"Follow this format for the release note:\n\n{format_instructions}\n\n"
        f"Version to use in the title: {version}\n\n"
        "Commit messages (one per line):\n" + "\n".join(commits)
    )

    payload = {
        "model": model,
        "messages": [
            {
                "role": "system",
                "content": (
                    "You generate release notes in Markdown from a list of commit messages. "
                    "Follow the structure provided. Be concise. Use the exact section headers given. "
                    "Do not invent commits; only use the provided list. "
                    "Output only the release note body as Markdown, no extra text or code fences."
                ),
            },
            {"role": "user", "content": user_content},
        ],
        "temperature": 0.3,
    }

    req = urllib.request.Request(
        "https://api.openai.com/v1/chat/completions",
        data=json.dumps(payload).encode(),
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {key}",
        },
        method="POST",
    )

    ctx = ssl.create_default_context(cafile=certifi.where())
    try:
        with urllib.request.urlopen(req, timeout=60, context=ctx) as resp:
            data = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        body = e.read().decode() if e.fp else ""
        print(f"OpenAI API error {e.code}: {body}", file=sys.stderr)
        raise SystemExit(1) from e
    except urllib.error.URLError as e:
        print(f"OpenAI request failed: {e.reason}", file=sys.stderr)
        raise SystemExit(1) from e

    content = (data.get("choices") or [{}])[0].get("message", {}).get("content") or ""
    if content.startswith("```"):
        content = re.sub(r"^```(?:markdown)?\n?", "", content)
        content = re.sub(r"\n?```\s*$", "", content)
    return content.strip()


def _call_anthropic(commits: list[str], format_instructions: str, version: str, model: str) -> str:
    """Call Anthropic messages API. Returns the generated markdown."""
    key = os.environ.get("ANTHROPIC_API_KEY", "").strip()
    if not key:
        print(
            "ANTHROPIC_API_KEY is not set. Add it as a repository secret for release notes.",
            file=sys.stderr,
        )
        raise SystemExit(1)

    user_content = (
        f"Follow this format for the release note:\n\n{format_instructions}\n\n"
        f"Version to use in the title: {version}\n\n"
        "Commit messages (one per line):\n" + "\n".join(commits)
    )
    system = (
        "You generate release notes in Markdown from a list of commit messages. "
        "Follow the structure provided. Be concise. Use the exact section headers given. "
        "Do not invent commits; only use the provided list. "
        "Output only the release note body as Markdown, no extra text or code fences."
    )

    payload = {
        "model": model,
        "max_tokens": 2048,
        "system": system,
        "messages": [{"role": "user", "content": user_content}],
    }

    req = urllib.request.Request(
        "https://api.anthropic.com/v1/messages",
        data=json.dumps(payload).encode(),
        headers={
            "Content-Type": "application/json",
            "x-api-key": key,
            "anthropic-version": "2023-06-01",
        },
        method="POST",
    )

    ctx = ssl.create_default_context(cafile=certifi.where())
    try:
        with urllib.request.urlopen(req, timeout=60, context=ctx) as resp:
            data = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        body = e.read().decode() if e.fp else ""
        print(f"Anthropic API error {e.code}: {body}", file=sys.stderr)
        raise SystemExit(1) from e
    except urllib.error.URLError as e:
        print(f"Anthropic request failed: {e.reason}", file=sys.stderr)
        raise SystemExit(1) from e

    content = ""
    for block in data.get("content") or []:
        if block.get("type") == "text":
            content = block.get("text") or ""
            break
    if content.startswith("```"):
        content = re.sub(r"^```(?:markdown)?\n?", "", content)
        content = re.sub(r"\n?```\s*$", "", content)
    return content.strip()


def run(
    project_root: Path,
    version: str,
    *,
    since_tag: str | None = None,
    template_path: Path | None = None,
    output_path: Path | None = None,
    model: str | None = None,
    provider: str | None = None,
) -> int:
    """
    Generate release notes: commits since previous tag -> OpenAI or Anthropic -> write to file or stdout.
    Returns 0 on success, 1 on failure.
    """
    ref = since_tag or _get_previous_tag(project_root)
    if not ref:
        print(
            "Could not find a previous tag. Use --since-tag <ref> for the first release.",
            file=sys.stderr,
        )
        return 1

    commits = _get_commits_since(project_root, ref)
    template = _load_template(template_path)
    format_instructions = template.replace("{{VERSION}}", version)

    provider = (provider or os.environ.get("RELEASE_NOTES_PROVIDER") or "anthropic").strip().lower()
    if provider not in ("openai", "anthropic"):
        print(f"Invalid provider {provider!r}; must be 'openai' or 'anthropic'.", file=sys.stderr)
        return 1
    if provider == "anthropic":
        model = (model or os.environ.get("ANTHROPIC_MODEL") or "claude-sonnet-4-5-20250929").strip()
        body = _call_anthropic(commits, format_instructions, version, model)
    else:
        model = (model or os.environ.get("OPENAI_MODEL") or "gpt-4o-mini").strip()
        body = _call_openai(commits, format_instructions, version, model)

    if not (body or "").strip():
        print("Release notes generation produced empty output.", file=sys.stderr)
        return 1

    if output_path is not None:
        output_path = Path(output_path)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(body, encoding="utf-8")
        print(f"Wrote release notes to {output_path}")
    else:
        print(body)

    return 0
