"""Tests for weave_gitops_tooling.release.bump."""

from pathlib import Path
from unittest.mock import patch

import pytest

from weave_gitops_tooling.release.bump import (
    _next_version,
    _read_current,
    _update_chart_yaml,
    _update_package_json,
    _update_values_yaml,
    run,
)


class TestReadCurrent:
    def test_reads_version_from_chart(self, tmp_path: Path) -> None:
        chart = tmp_path / "charts" / "gitops-server"
        chart.mkdir(parents=True)
        (chart / "Chart.yaml").write_text(
            'apiVersion: v2\nname: weave-gitops\nversion: 1.2.3\nappVersion: "v1.2.3"\n'
        )
        base, rc = _read_current(chart / "Chart.yaml")
        assert base == "1.2.3" and rc is None

    def test_parses_rc_suffix(self, tmp_path: Path) -> None:
        chart = tmp_path / "charts" / "gitops-server"
        chart.mkdir(parents=True)
        (chart / "Chart.yaml").write_text(
            'version: 0.39.0-rc.2 # x-release-please\nappVersion: "v0.39.0-rc.2"\n'
        )
        base, rc = _read_current(chart / "Chart.yaml")
        assert base == "0.39.0" and rc == 2

    def test_missing_raises(self, tmp_path: Path) -> None:
        (tmp_path / "Chart.yaml").write_text("name: x\n# no version\n")
        with pytest.raises(SystemExit):
            _read_current(tmp_path / "Chart.yaml")


class TestNextVersion:
    def test_patch(self) -> None:
        assert _next_version("0.1.0", "patch") == "0.1.1"
        assert _next_version("1.2.3", "patch") == "1.2.4"

    def test_minor(self) -> None:
        assert _next_version("0.1.0", "minor") == "0.2.0"
        assert _next_version("1.2.3", "minor") == "1.3.0"

    def test_major(self) -> None:
        assert _next_version("0.1.0", "major") == "1.0.0"
        assert _next_version("1.2.3", "major") == "2.0.0"

    def test_strips_v_prefix(self) -> None:
        assert _next_version("v0.1.0", "patch") == "0.1.1"

    def test_invalid_version_raises(self) -> None:
        for bad in ("1.2", "1.2.3.4", "x.y.z"):
            with pytest.raises(SystemExit):
                _next_version(bad, "patch")

    def test_invalid_bump_raises(self) -> None:
        with pytest.raises(SystemExit):
            _next_version("1.2.3", "foo")

    def test_rc_from_existing_rc(self) -> None:
        assert _next_version("0.39.0", "rc", rc_num=2) == "0.39.0-rc.3"
        assert _next_version("1.0.0", "rc", rc_num=1) == "1.0.0-rc.2"

    def test_rc_from_clean_version(self) -> None:
        assert _next_version("0.39.0", "rc", rc_num=None) == "0.39.0-rc.1"
        assert _next_version("1.2.3", "rc") == "1.2.3-rc.1"

    def test_minor_rc_from_039x(self) -> None:
        assert _next_version("0.39.0", "minor-rc") == "0.40.0-rc.1"
        assert _next_version("0.39.1", "minor-rc") == "0.40.0-rc.1"
        assert _next_version("0.39.0-rc.2", "minor-rc") == "0.40.0-rc.1"

    def test_patch_rc(self) -> None:
        assert _next_version("0.39.0", "patch-rc") == "0.39.1-rc.1"
        assert _next_version("0.39.0-rc.2", "patch-rc") == "0.39.1-rc.1"

    def test_major_rc(self) -> None:
        assert _next_version("0.39.0", "major-rc") == "1.0.0-rc.1"
        assert _next_version("1.2.3", "major-rc") == "2.0.0-rc.1"


class TestUpdateChartYaml:
    def test_updates_version_and_appversion(self, tmp_path: Path) -> None:
        p = tmp_path / "Chart.yaml"
        p.write_text('version: 0.39.0-rc.2 # x-release-please\nappVersion: "v0.39.0-rc.2" # x\n')
        assert _update_chart_yaml(p, "0.39.0", "0.40.0") is True
        t = p.read_text()
        assert "version: 0.40.0" in t
        assert 'appVersion: "v0.40.0"' in t

    def test_returns_false_when_old_not_found(self, tmp_path: Path) -> None:
        p = tmp_path / "Chart.yaml"
        p.write_text('version: 0.50.0\nappVersion: "v0.50.0"\n')
        assert _update_chart_yaml(p, "0.39.0", "0.40.0") is False

    def test_rc_bump_updates_version_and_appversion(self, tmp_path: Path) -> None:
        p = tmp_path / "Chart.yaml"
        p.write_text('version: 0.39.0-rc.2 # x\nappVersion: "v0.39.0-rc.2" # x\n')
        assert _update_chart_yaml(p, "0.39.0", "0.39.0-rc.3") is True
        t = p.read_text()
        assert "version: 0.39.0-rc.3" in t
        assert 'appVersion: "v0.39.0-rc.3"' in t


class TestUpdateValuesYaml:
    def test_updates_image_tag(self, tmp_path: Path) -> None:
        p = tmp_path / "values.yaml"
        p.write_text('image:\n  repository: ghcr.io/foo\n  tag: "v0.39.0-rc.2" # x\n')
        assert _update_values_yaml(p, "0.39.0", "0.40.0") is True
        assert 'tag: "v0.40.0"' in p.read_text()

    def test_returns_false_when_old_not_found(self, tmp_path: Path) -> None:
        p = tmp_path / "values.yaml"
        p.write_text('image:\n  tag: "v0.50.0"\n')
        assert _update_values_yaml(p, "0.39.0", "0.40.0") is False


class TestUpdatePackageJson:
    def test_updates_version(self, tmp_path: Path) -> None:
        p = tmp_path / "package.json"
        p.write_text('{"name": "@weaveworks/weave-gitops", "version": "0.39.0-rc.2"}\n')
        assert _update_package_json(p, "0.39.0", "0.40.0") is True
        assert '"version": "0.40.0"' in p.read_text()

    def test_returns_false_when_old_not_found(self, tmp_path: Path) -> None:
        p = tmp_path / "package.json"
        p.write_text('{"version": "0.50.0"}\n')
        assert _update_package_json(p, "0.39.0", "0.40.0") is False


class TestRun:
    def test_success_updates_chart_values_package(self, tmp_path: Path) -> None:
        chart_dir = tmp_path / "charts" / "gitops-server"
        chart_dir.mkdir(parents=True)
        (chart_dir / "Chart.yaml").write_text('version: 0.39.0\nappVersion: "v0.39.0"\n')
        (chart_dir / "values.yaml").write_text('image:\n  tag: "v0.39.0"\n')
        (tmp_path / "package.json").write_text('{"version": "0.39.0"}\n')

        rc = run(tmp_path, "patch")

        assert rc == 0
        assert "version: 0.39.1" in (chart_dir / "Chart.yaml").read_text()
        assert 'appVersion: "v0.39.1"' in (chart_dir / "Chart.yaml").read_text()
        assert 'tag: "v0.39.1"' in (chart_dir / "values.yaml").read_text()
        assert '"version": "0.39.1"' in (tmp_path / "package.json").read_text()

    def test_fails_without_chart(self, tmp_path: Path) -> None:
        assert run(tmp_path, "patch") == 1

    def test_appends_to_github_output_when_set(self, tmp_path: Path) -> None:
        chart_dir = tmp_path / "charts" / "gitops-server"
        chart_dir.mkdir(parents=True)
        (chart_dir / "Chart.yaml").write_text('version: 0.39.0\nappVersion: "v0.39.0"\n')
        (chart_dir / "values.yaml").write_text('image:\n  tag: "v0.39.0"\n')
        (tmp_path / "package.json").write_text('{"version": "0.39.0"}\n')
        gh_out = tmp_path / "github_output.txt"
        with patch.dict("os.environ", {"GITHUB_OUTPUT": str(gh_out)}, clear=False):
            rc = run(tmp_path, "patch")
        assert rc == 0
        assert gh_out.is_file()
        assert "version=0.39.1\n" in gh_out.read_text()

    def test_rc_bump_from_rc_updates_all_files(self, tmp_path: Path) -> None:
        chart_dir = tmp_path / "charts" / "gitops-server"
        chart_dir.mkdir(parents=True)
        (chart_dir / "Chart.yaml").write_text(
            'version: 0.39.0-rc.2 # x\nappVersion: "v0.39.0-rc.2"\n'
        )
        (chart_dir / "values.yaml").write_text('image:\n  tag: "v0.39.0-rc.2"\n')
        (tmp_path / "package.json").write_text('{"version": "0.39.0-rc.2"}\n')

        rc = run(tmp_path, "rc")

        assert rc == 0
        assert "version: 0.39.0-rc.3" in (chart_dir / "Chart.yaml").read_text()
        assert 'appVersion: "v0.39.0-rc.3"' in (chart_dir / "Chart.yaml").read_text()
        assert 'tag: "v0.39.0-rc.3"' in (chart_dir / "values.yaml").read_text()
        assert '"version": "0.39.0-rc.3"' in (tmp_path / "package.json").read_text()

    def test_rc_bump_from_clean_produces_rc1(self, tmp_path: Path) -> None:
        chart_dir = tmp_path / "charts" / "gitops-server"
        chart_dir.mkdir(parents=True)
        (chart_dir / "Chart.yaml").write_text('version: 0.39.0\nappVersion: "v0.39.0"\n')
        (chart_dir / "values.yaml").write_text('image:\n  tag: "v0.39.0"\n')
        (tmp_path / "package.json").write_text('{"version": "0.39.0"}\n')

        rc = run(tmp_path, "rc")

        assert rc == 0
        assert "version: 0.39.0-rc.1" in (chart_dir / "Chart.yaml").read_text()
        assert 'tag: "v0.39.0-rc.1"' in (chart_dir / "values.yaml").read_text()

    def test_rc_bump_writes_github_output(self, tmp_path: Path) -> None:
        chart_dir = tmp_path / "charts" / "gitops-server"
        chart_dir.mkdir(parents=True)
        (chart_dir / "Chart.yaml").write_text('version: 0.39.0-rc.2\nappVersion: "v0.39.0-rc.2"\n')
        (chart_dir / "values.yaml").write_text('image:\n  tag: "v0.39.0-rc.2"\n')
        (tmp_path / "package.json").write_text('{"version": "0.39.0-rc.2"}\n')
        gh_out = tmp_path / "github_output.txt"
        with patch.dict("os.environ", {"GITHUB_OUTPUT": str(gh_out)}, clear=False):
            rc = run(tmp_path, "rc")
        assert rc == 0
        assert "version=0.39.0-rc.3\n" in gh_out.read_text()

    def test_minor_rc_from_039x_produces_0400_rc1(self, tmp_path: Path) -> None:
        chart_dir = tmp_path / "charts" / "gitops-server"
        chart_dir.mkdir(parents=True)
        (chart_dir / "Chart.yaml").write_text(
            'version: 0.39.0-rc.2 # x\nappVersion: "v0.39.0-rc.2"\n'
        )
        (chart_dir / "values.yaml").write_text('image:\n  tag: "v0.39.0-rc.2"\n')
        (tmp_path / "package.json").write_text('{"version": "0.39.0-rc.2"}\n')

        rc = run(tmp_path, "minor-rc")

        assert rc == 0
        assert "version: 0.40.0-rc.1" in (chart_dir / "Chart.yaml").read_text()
        assert 'appVersion: "v0.40.0-rc.1"' in (chart_dir / "Chart.yaml").read_text()
        assert 'tag: "v0.40.0-rc.1"' in (chart_dir / "values.yaml").read_text()
        assert '"version": "0.40.0-rc.1"' in (tmp_path / "package.json").read_text()
