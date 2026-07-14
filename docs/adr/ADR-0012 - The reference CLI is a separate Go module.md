---
adr-status: Accepted
superseded-by:
---

# The reference CLI is a separate Go module so the library's dependency graph stays clean

Amends a consequence of [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]] and overrides an Out-of-Scope deferral in [[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness|EPIC-0002]].

Today the whole repository is a **single Go module**. Its `go.mod` already carries `github.com/spf13/cobra`, which is used only by the CLI under `cmd/`, and the `view` command ([[0016 - CLI view read-only TUI dialogue browser|task 0016]]) adds a second CLI-only framework (Bubble Tea, plus its `charmbracelet` branch). With Go 1.17+ module-graph pruning a library consumer that imports only the `dcanvas` package never *compiles* cobra or Bubble Tea, but these dependencies still appear in the consumer's **module graph** and `go.sum`. [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]] sells the core as dependency-clean; letting a second CLI-only framework accrete in the shared `go.mod` erodes that.

## Decision

- Split the repository into **two Go modules**:
  - **Library module** at the repo root: `github.com/sbtlocalization/dcanvas`. It holds the core `dcanvas` package (stdlib only) **and** the `layout` package (with its `github.com/nulab/autog` dependency).
  - **CLI module** under `cmd/dcanvas/`: `github.com/sbtlocalization/dcanvas/cmd/dcanvas`, with its own `go.mod` that requires the library module plus cobra and Bubble Tea. No consumer of the library ever sees cobra or Bubble Tea.
- The boundary is drawn by one asymmetry — *is there a consumer who legitimately wants X without Y?* Someone importing `layout` genuinely wants `autog`, so `autog` is a real dependency, not pollution, and `layout` **stays in the library module** (layout ownership is unchanged — see [[ADR-0009 - Layout may read x-kind and the loop strategy is configurable|ADR-0009]]). No consumer ever wants cobra/Bubble Tea as a library, so those are the only things extracted. One split, not one-module-per-dependency.
- During co-development the two modules are bound by a **committed `go.work`** at the repo root, so in-repo `go build`/`go test` resolve the library locally without a `replace` directive. `go.work` affects only builds *inside this repo*; a library consumer never sees it.
- CI runs a **single job** that vets and tests each module explicitly (`go vet -C <module> ./...`, `go test -C <module> ./...`), because a root `./...` package pattern does not descend into a nested module.

## Consequences

- Makes good on [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]]'s zero-third-party intent at the module-graph level, not just at compile time.
- Overrides [[EPIC-0002 - dCanvas CLI TUI and round-trip library robustness|EPIC-0002]]'s Out-of-Scope note that deferred module splitting until release tagging: it is done now, because a second CLI-only dependency makes the cost concrete rather than speculative (the same "don't build in a vacuum" reasoning as [[ADR-0008 - Extract format and library to a separate repository|ADR-0008]]).
- When the library is eventually tagged for `go get`, the CLI module switches its `require` from the workspace-resolved local copy to a real version; `go.work` can remain for development or be dropped.
- The core decision of [[ADR-0011 - An in-repo reference CLI is a round-tripping consumer and the layout engine stays swappable|ADR-0011]] — the CLI ships in `cmd/` of this repo as the library's round-trip acceptance consumer — is unchanged; a nested module is still under `cmd/` in this repo.
