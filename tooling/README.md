# weave-gitops-tooling

Release and build tooling for Weave GitOps.

## Commands

- **weavegitops release bump** [patch|minor|major|rc|patch-rc|minor-rc|major-rc]  
  Bump from `charts/gitops-server/Chart.yaml`; updates Chart, values, package.json. **rc**: `0.39.0-rc.2` → `0.39.0-rc.3` or `0.39.0` → `0.39.0-rc.1`. **minor-rc**: `0.39.x` → `0.40.0-rc.1` (start RC for next minor). Also **patch-rc**, **major-rc**. Writes `version` to `GITHUB_OUTPUT` when set.

- **weavegitops release generate-notes** --version X.Y.Z [--output PATH] [--template PATH] [--since-tag TAG] [--provider openai|anthropic]  
  Generate release notes from commits since the previous tag via OpenAI or Anthropic.

## Install

```bash
pip install -e ./tooling
# or with dev (ruff, pytest, pre-commit): pip install -e ./tooling[dev]
# In CI: pip install -e ./tooling
```

## Lint and format

Uses [ruff](https://docs.astral.sh/ruff/) with the same rules as RERP tooling (pycodestyle, Pyflakes, bugbear, isort, etc.). See `[tool.ruff]` in `pyproject.toml`.

```bash
cd tooling
.venv/bin/pip install -e ".[dev]"
.venv/bin/ruff check src/ tests/
.venv/bin/ruff format src/ tests/
```

## Version files

- **Source of truth:** `charts/gitops-server/Chart.yaml` → `version` (X.Y.Z or X.Y.Z-rc.N).
- **Updated by bump:** Chart.yaml (version, appVersion), values.yaml (image.tag), package.json (version).
