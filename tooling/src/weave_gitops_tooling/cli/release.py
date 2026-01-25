"""CLI for weavegitops release: bump, generate-notes."""

from __future__ import annotations

import sys
from pathlib import Path

from weave_gitops_tooling.release.bump import run as bump_run
from weave_gitops_tooling.release.notes import run as notes_run


def run_release(args, project_root: Path) -> None:
    cmd = getattr(args, "release_cmd", None)
    if cmd == "bump":
        bump = (getattr(args, "bump", None) or "patch").lower()
        rc = bump_run(project_root, bump)
        sys.exit(rc)
    if cmd == "generate-notes":
        version = getattr(args, "version", None) or ""
        if not version:
            print("weavegitops release generate-notes: --version is required", file=sys.stderr)
            sys.exit(1)
        rc = notes_run(
            project_root,
            version,
            since_tag=getattr(args, "since_tag", None),
            template_path=Path(args.template) if getattr(args, "template", None) else None,
            output_path=Path(args.output) if getattr(args, "output", None) else None,
            model=getattr(args, "model", None),
            provider=getattr(args, "provider", None),
        )
        sys.exit(rc)
    print("weavegitops release: use subcommand 'bump' or 'generate-notes'")
    print("  weavegitops release bump [patch|minor|major|rc|patch-rc|minor-rc|major-rc]")
    print(
        "  weavegitops release generate-notes --version X.Y.Z [--output PATH] [--template PATH] [--since-tag TAG] [--provider openai|anthropic]"
    )
    sys.exit(1)
