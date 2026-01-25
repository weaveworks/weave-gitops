"""weavetooling CLI entry point."""

import os
import sys
from pathlib import Path

from . import release as release_cli
from . import ci as ci_cli


def _get_project_root() -> Path:
    root = Path(os.environ.get("WEAVE_GITOPS_PROJECT_ROOT", ".")).resolve()
    if str(root) == ".":
        root = Path.cwd()
    return root


def main() -> None:
    parser = _build_parser()
    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    project_root = _get_project_root()

    if args.command == "release":
        if not getattr(args, "release_cmd", None):
            print("weavetooling release: missing subcommand")
            print("  bump, generate-notes")
            print("  Use: weavetooling release --help")
            sys.exit(1)
        release_cli.run_release(args, project_root)
        return

    if args.command == "ci":
        if not getattr(args, "ci_cmd", None):
            print("weavetooling ci: missing subcommand")
            print("  is-tag")
            print("  Use: weavetooling ci --help")
            sys.exit(1)
        ci_cli.run_ci(args)
        return

    print("weavetooling: use 'release' (bump, generate-notes) or 'ci' (is-tag).")
    sys.exit(1)


def _build_parser():
    import argparse

    p = argparse.ArgumentParser(
        prog="weavetooling",
        description="Weave GitOps tooling: release bump, generate-notes, ci helpers",
    )
    sub = p.add_subparsers(dest="command", help="Command")

    prl = sub.add_parser(
        "release",
        help="Release: bump (Chart/values/package.json), generate-notes (OpenAI/Anthropic)",
    )
    prl_sub = prl.add_subparsers(dest="release_cmd")
    prlb = prl_sub.add_parser(
        "bump",
        help="Bump version from Chart.yaml; update Chart, values, package.json",
    )
    prlb.add_argument(
        "bump",
        nargs="?",
        default="patch",
        choices=["patch", "minor", "major", "rc", "patch-rc", "minor-rc", "major-rc"],
        help="Bump: patch|minor|major|rc|patch-rc|minor-rc|major-rc. minor-rc: 0.39.x -> 0.40.0-rc.1",
    )
    prlgn = prl_sub.add_parser(
        "generate-notes",
        help="Generate release notes from commits since last tag via OpenAI or Anthropic",
    )
    prlgn.add_argument("--version", "-v", required=True, help="Release version (e.g. 1.2.3)")
    prlgn.add_argument("--output", "-o", help="Write notes to file (default: stdout)")
    prlgn.add_argument("--template", "-t", help="Custom template path (default: built-in)")
    prlgn.add_argument("--since-tag", help="Git ref for 'since' (default: previous tag)")
    _p = (os.environ.get("RELEASE_NOTES_PROVIDER") or "anthropic").strip().lower()
    _prov = _p if _p in ("openai", "anthropic") else "anthropic"
    prlgn.add_argument(
        "--provider",
        choices=["openai", "anthropic"],
        default=_prov,
        help="AI provider (default: RELEASE_NOTES_PROVIDER or anthropic)",
    )
    prlgn.add_argument(
        "--model",
        help="Model: OpenAI (gpt-4o-mini) or Anthropic (claude-sonnet-4-5-20250929)",
    )

    pci = sub.add_parser("ci", help="CI helpers: is-tag (for GHA job outputs)")
    pci_sub = pci.add_subparsers(dest="ci_cmd")
    pci_sub.add_parser("is-tag", help="Print true if GITHUB_REF is refs/tags/v*, else false")

    return p
