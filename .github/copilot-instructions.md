# Copilot Coding Agent Instructions

## Setup

This is a Go project. Before making any changes, ensure the Go toolchain is available:

```sh
go version
```

The linter used in CI is **golangci-lint v2** (`golangci/golangci-lint-action@v9` with `version: v2.10.1`).
Install it locally to run before finalizing changes:

```sh
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.10.1
```

## Workflow — Always Follow This Order

1. **Build** — confirm the code compiles:
   ```sh
   go build ./...
   ```

2. **Test** — run all tests and check for failures:
   ```sh
   go test ./...
   ```

3. **Lint** — run the linter and fix every issue before committing:
   ```sh
   golangci-lint run
   ```
   The linter enforces strict rules (see `.golangci.yml`). Common pitfalls:
   - `varnamelen`: variable names must be descriptive (avoid `id`, `ok`, `tc` — use `hashID`, `found`, `testCase`)
   - `wsl_v5`: blank lines required before `for`/`if`/`return` statements that follow a different statement type
   - `gci`: imports must be grouped — stdlib | third-party | local module — with a blank line between each group
   - `goconst`: string literals used 2+ times must be extracted to a constant
   - `dupl`: functions with similar structure (≥100 lines) must be refactored (use table-driven tests)
   - `prealloc`: slices appended to in a loop must be preallocated
   - `gocritic` rangeValCopy: avoid range-copying large structs; use index-based access

4. **Report progress** only after all three checks above pass cleanly.

## Import Grouping (gci)

Always format imports in three separate groups with a blank line between each:

```go
import (
    "os"         // stdlib
    "testing"

    "github.com/lainio/err2/try"  // third-party

    "github.com/lauravuo/vegaanibotti/blog/base"  // local (same module)
)
```

The module path is `github.com/lauravuo/vegaanibotti` (see `go.mod`).

## Test Coverage

CI uploads coverage to Codecov. Patch coverage should be ≥80%.
After adding new code, add matching tests that cover all new branches.
Use table-driven tests for similar scenarios to avoid the `dupl` linter.
