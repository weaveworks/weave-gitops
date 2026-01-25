"""CI helper commands (is-tag, etc.)."""

import os


def run_ci(args) -> None:
    if getattr(args, "ci_cmd", None) == "is-tag":
        _run_is_tag()
        return
    raise SystemExit("weavetooling ci: use is-tag. Run: weavetooling ci --help")


def _run_is_tag() -> None:
    """Print 'true' if GITHUB_REF starts with refs/tags/v, else 'false'. For GHA job outputs."""
    ref = os.environ.get("GITHUB_REF", "")
    print("true" if ref.startswith("refs/tags/v") else "false")
