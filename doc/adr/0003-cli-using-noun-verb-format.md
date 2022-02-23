# 3. CLI using noun verb format

Date: 2021-07-09

Decision made before this date

## Status

Superceded by [11. CLI verb noun](0011-cli-verb-noun.md)

## Context

### Glossary
* **Command** - Action to perform, e.g., get, install, add, help
* **Object type** - Entity classification the CLI will operate on

We require consistency in the weave-gitops CLI `wego` and the ability to extend the object types and commands available.  For example, wego, wraps the flux UI and exposes it via the `wego flux` command.  We want a standardized way to enhance and augment the CLI without always requiring additional discussions.

## Decision

The CLI commands will be in the noun verb format.

We debated between verb noun, e.g., `wego add app foo` and noun verb, e.g., `wego app add foo`

### Pros for noun verb
* Format for new commands is consistent i.e.,`wego` noun command
* If we need to extend wego with a new object type, we aren't required to implement a specific set of commands or provide stub implementations
* Extensions can create commands independent of wego primitive commands
* Supports objects with diverse commands

### Pros for verb noun
* Consistent with many other tools in the space, e.g.,  `kubectl`, `git`.  Kubectl is a hybrid as it supports extensibility and is in the verb noun format
* Follows the defacto go standard CLI tooling [Cobra](https://github.com/spf13/cobra#concepts)
* Enables the ability to operate on multiple object types in the same call
* The ability to enforce a standard set of commands across object types

Ultimately, extensibility was the deciding factor for moving forward with a noun verb format.

## Consequences

Object type is required on CLIs.
