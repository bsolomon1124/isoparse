# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go test -race -count=1 ./...                              # full test run
go test -run TestParseDatetime ./...                      # single test
go test -cover -coverprofile=cov.out ./... && go tool cover -func=cov.out
go test -run '^$' -fuzz='^FuzzParseDatetime$' -fuzztime=30s ./...   # fuzz one target
go test -bench BenchmarkParseDatetime -benchmem -benchtime=3s ./... # bench one target
go vet ./...
gofmt -l .                                                # should output nothing
golangci-lint run ./...                                   # uses .golangci.yml (v2 schema)
```

Go 1.25 is the floor (see `go.mod`); no `toolchain` directive — users' local Go resolves. CI matrix covers Go 1.25 and 1.26 across Linux/macOS/Windows, plus separate `lint` and `govulncheck` jobs.

## Architecture

Single-package module; all source lives at repo root. Do not re-nest under a subdirectory — the import path is `github.com/bsolomon1124/isoparse`.

File layout:

- `isoparse.go` — all production code.
- `isoparse_test.go` — internal (`package isoparse`) tests; access to unexported helpers.
- `example_test.go`, `bench_test.go`, `fuzz_test.go` — external (`package isoparse_test`); examples with verified `// Output:`, benchmarks, and fuzz targets.
- `testdata/fuzz/<Target>/` — committed fuzz-corpus regressions; the fuzzer has already found two panics (truncated-time + truncated-week-date), both fixed and now guarded by corpus entries.

**Parsing pipeline.** `ParseDatetime` is the top-level entry. It calls `parseDate`, which tries two strategies in sequence:

1. `parseDateCommon` — fastpath for `YYYY`, `YYYY-MM`, `YYYY-MM-DD`, `YYYYMMDD`.
2. `parseDateUncommon` — fallback for ISO week dates (`YYYY-Www-D`, `YYYYWwwD`) and ordinal dates (`YYYY-DDD`, `YYYYDDD`).

Both return a cursor position `pos`. `ParseDatetime` uses `pos < len(datetime)` to decide whether a time portion follows; the separator between date and time must be a non-numeric ASCII byte (not strictly `T`). The time portion is handed to `ParseTime`, which in turn calls `parseTimezone` when it sees `Z`, `+`, `-`, or U+2212.

**strictDate validates units in declaration order** (year → month → day → hour → minute → second → nsec). When adding invalid-input fixtures, only *one* unit may be out of range per row or the earlier check will short-circuit and you won't exercise the intended branch. This is a real pitfall — the original test fixtures had this bug.

**Divergence from Go's `time.Date`**: Go normalizes overflow (month 13 → next January); `strictDate` does not. The only exception is hour 24 as midnight, which intentionally relies on `time.Date`'s rollover (see `ParseTime`).

**Timezone semantics.** No offset in input → `time.Local`. With offset → `time.FixedZone("UTC", secondsEast)` — no IANA name is inferred because offsets are DST-ambiguous. Unicode minus U+2212 is accepted in addition to ASCII `-`; `parseTimezone` uses `utf8.DecodeRuneInString` because U+2212 is multi-byte. `SetLoc` reattaches a `Location` without shifting the wall clock (unlike `time.Time.In`).

`*ParseError` is the sole exported error type. Callers use `errors.As`.

## Conventions

- Don't shadow Go 1.21+ builtins (`min`, `max`, `clear`) with local helpers.
- `.golangci.yml` uses the v2 config schema (required by golangci-lint 2.x). The enabled linters are intentionally conservative — `staticcheck` has already caught one real latent bug (SA4003 in separator validation), so treat its findings as substantive, not stylistic.
- `CHANGELOG`-style comments (e.g. `// removed X`, `// renamed from Y`) are not used; the git history is the source of truth.

## Pull request titles

Follow Conventional Commits format: `<type>: <description>` (e.g. `chore: update dependencies`, `fix: handle edge case`, `feat: add new parse function`).
