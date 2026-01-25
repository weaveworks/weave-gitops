"""Tests for weavetooling ci is-tag."""

import os

import pytest

from weave_gitops_tooling.cli.ci import _run_is_tag


def test_is_tag_true(monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.setenv("GITHUB_REF", "refs/tags/v1.2.3")
    _run_is_tag()
    assert capsys.readouterr().out.strip() == "true"


def test_is_tag_false_refs_heads(monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.setenv("GITHUB_REF", "refs/heads/main")
    _run_is_tag()
    assert capsys.readouterr().out.strip() == "false"


def test_is_tag_false_empty(monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.delenv("GITHUB_REF", raising=False)
    _run_is_tag()
    assert capsys.readouterr().out.strip() == "false"


def test_is_tag_false_refs_tags_no_v(monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture[str]) -> None:
    monkeypatch.setenv("GITHUB_REF", "refs/tags/1.0.0")  # no 'v' prefix
    _run_is_tag()
    assert capsys.readouterr().out.strip() == "false"
