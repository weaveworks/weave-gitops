"""Tests for weave_gitops_tooling.release.notes."""

from pathlib import Path
from unittest.mock import patch

import pytest

from weave_gitops_tooling.release.notes import (
    DEFAULT_TEMPLATE,
    _get_commits_since,
    _get_previous_tag,
    _load_template,
    run,
)


class TestGetPreviousTag:
    def test_returns_tag(self, tmp_path: Path) -> None:
        from types import SimpleNamespace

        with patch("subprocess.run") as m:
            m.return_value = SimpleNamespace(stdout="v1.2.3\n")
            assert _get_previous_tag(tmp_path) == "v1.2.3"

    def test_returns_none_on_error(self, tmp_path: Path) -> None:
        import subprocess

        with patch("subprocess.run", side_effect=subprocess.CalledProcessError(1, "git")):
            assert _get_previous_tag(tmp_path) is None


class TestGetCommitsSince:
    def test_returns_list(self, tmp_path: Path) -> None:
        from types import SimpleNamespace

        with patch("subprocess.run") as m:
            m.return_value = SimpleNamespace(stdout="feat: x\nfix: y\n")
            assert _get_commits_since(tmp_path, "v1.0.0") == ["feat: x", "fix: y"]

    def test_raises_system_exit_on_invalid_ref(self, tmp_path: Path) -> None:
        import subprocess

        err = subprocess.CalledProcessError(
            128, ["git", "log", "v99..HEAD", "--pretty=format:%s"], stderr="fatal: bad revision\n"
        )
        with patch("subprocess.run", side_effect=err), pytest.raises(SystemExit) as exc_info:
            _get_commits_since(tmp_path, "v99")
        assert exc_info.value.code == 1


class TestLoadTemplate:
    def test_returns_default_when_none(self) -> None:
        assert _load_template(None) == DEFAULT_TEMPLATE

    def test_returns_file_content(self, tmp_path: Path) -> None:
        p = tmp_path / "t.md"
        p.write_text("# Custom\n")
        assert _load_template(p) == "# Custom\n"


class TestRun:
    def test_fails_without_previous_tag(self, tmp_path: Path) -> None:
        with patch("weave_gitops_tooling.release.notes._get_previous_tag", return_value=None):
            assert run(tmp_path, "1.0.0", since_tag=None) == 1

    def test_succeeds_with_since_tag_and_mocked_openai(self, tmp_path: Path) -> None:
        out = tmp_path / "out.md"
        with (
            patch(
                "weave_gitops_tooling.release.notes._get_commits_since", return_value=["feat: x"]
            ),
            patch(
                "weave_gitops_tooling.release.notes._call_openai",
                return_value="# Release v1.0.0\n\nDone.",
            ),
        ):
            rc = run(tmp_path, "1.0.0", since_tag="v0.9.0", output_path=out, provider="openai")
        assert rc == 0
        assert "Release v1.0.0" in out.read_text()

    def test_run_provider_anthropic_uses_call_anthropic(
        self, tmp_path: Path, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with (
            patch(
                "weave_gitops_tooling.release.notes._get_commits_since", return_value=["feat: x"]
            ),
            patch(
                "weave_gitops_tooling.release.notes._call_anthropic",
                return_value="# Release v1.0.0\n\nClaude notes.",
            ),
        ):
            rc = run(tmp_path, "1.0.0", since_tag="v0.9.0", output_path=None, provider="anthropic")
        assert rc == 0
        assert "Claude notes" in capsys.readouterr().out

    def test_run_fails_on_invalid_provider(
        self, tmp_path: Path, capsys: pytest.CaptureFixture[str]
    ) -> None:
        with patch("weave_gitops_tooling.release.notes._get_commits_since", return_value=["a"]):
            rc = run(tmp_path, "1.0.0", since_tag="v0.9.0", output_path=None, provider="claude")
        assert rc == 1
        assert "Invalid provider" in capsys.readouterr().err
