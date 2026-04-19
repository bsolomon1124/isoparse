# isoparse

[![CI](https://github.com/bsolomon1124/isoparse/actions/workflows/ci.yml/badge.svg)](https://github.com/bsolomon1124/isoparse/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/bsolomon1124/isoparse.svg)](https://pkg.go.dev/github.com/bsolomon1124/isoparse)
[![Go Report Card](https://goreportcard.com/badge/github.com/bsolomon1124/isoparse)](https://goreportcard.com/report/github.com/bsolomon1124/isoparse)
[![Coverage](https://img.shields.io/badge/coverage-99.5%25-brightgreen)](#testing)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

Package `isoparse` parses ISO 8601 date, time, and datetime strings into Go `time.Time` values — without requiring the caller to specify a layout in advance.

It is a Go port of the `isoparser` module (by Paul Ganssle) from Python's [`dateutil`](https://github.com/dateutil/dateutil) library, adapted to Go's `time` package semantics.

## Install

Requires Go 1.25 or newer.

```
go get github.com/bsolomon1124/isoparse
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/bsolomon1124/isoparse"
)

func main() {
	t, err := isoparse.ParseDatetime("2024-03-15T14:30:00Z")
	if err != nil {
		panic(err)
	}
	fmt.Println(t) // 2024-03-15 14:30:00 +0000 UTC

	d, _ := isoparse.ParseDate("2024-W11-5")
	fmt.Println(d) // 2024-03-15 00:00:00 ...

	hms, loc, _ := isoparse.ParseTime("14:30:00+02:00")
	fmt.Println(hms, loc) // [14 30 0 0] UTC
}
```

Full API documentation is on [pkg.go.dev](https://pkg.go.dev/github.com/bsolomon1124/isoparse).

## Supported formats

Dates:

| Format       | Example        |
|--------------|----------------|
| `YYYY`       | `2024`         |
| `YYYY-MM`    | `2024-03`      |
| `YYYY-MM-DD` | `2024-03-15`   |
| `YYYYMMDD`   | `20240315`     |
| `YYYY-DDD`   | `2024-075`     (ordinal) |
| `YYYYDDD`    | `2024075`      (ordinal) |
| `YYYY-Www-D` | `2024-W11-5`   (ISO week) |
| `YYYYWwwD`   | `2024W115`     (ISO week) |

Times (with optional `Z`, `±HH`, `±HHMM`, or `±HH:MM` offset):

| Format             | Example           |
|--------------------|-------------------|
| `HH`               | `14`              |
| `HH:MM` / `HHMM`   | `14:30`           |
| `HH:MM:SS`         | `14:30:00`        |
| `HH:MM:SS.ffffff`  | `14:30:00.123456` |

Datetimes are any date + any time joined by a non-numeric ASCII separator (`T` is standard; `space`, `_`, `-`, etc. are also accepted).

## API

```go
func ParseDatetime(s string) (time.Time, error)
func ParseDate(s string) (time.Time, error)
func ParseTime(s string) (components [4]int, tz *time.Location, err error)
func SetLoc(t time.Time, loc *time.Location) time.Time

type ParseError struct {
    Datetime string
    Message  string
}
```

`ParseDatetime` accepts date-only input as well, but `ParseDate` is faster when the caller knows no time component is present. Parse failures return a `*ParseError` — check via `errors.As`.

## Time zone handling

Go's `time.Time` has no concept of a naive datetime; a `Location` is mandatory. This package follows these rules:

- **No offset in the input** → result uses `time.Local`.
- **Offset present** (`Z`, `±HH`, `±HHMM`, `±HH:MM`) → result uses `time.FixedZone("UTC", secondsEast)`. No IANA name is inferred; e.g. `-05:00` is inherently ambiguous between EST, ECT, COT, etc.
- **Unicode minus sign** (U+2212) is accepted in place of ASCII `-`.

`SetLoc(t, loc)` returns a new `time.Time` with the same wall-clock components but a different `Location` — unlike `time.Time.In`, which preserves the instant and shifts the wall clock.

## Conformance to ISO 8601

Conforms mostly to [ISO 8601:2004](https://www.iso.org/standard/40874.html). Known deviations:

- Separator between date and time may be any non-numeric ASCII character, not strictly `T`.
- Years are restricted to `0001`–`9999` (no expanded representations, no negative years).
- Time intervals and recurring intervals (§4.4, §4.5) are not parsed.
- Reduced-accuracy century/hour representations (`"19"` for 1900, `"23"` for 23:00:00) are not accepted.
- Fractional minutes/hours (e.g. `14:30,5`) are not parsed; only fractional seconds are supported.
- Second-fraction precision beyond 9 digits is truncated (Go's `time` has nanosecond precision).

## Testing

```
go test -race -cover ./...
```

Current statement coverage: **99.5%**. The single uncovered line is a defensive branch in `ParseDatetime` that is unreachable under the current control flow.

CI runs on Go 1.25 and 1.26 across Linux, macOS, and Windows, plus a `golangci-lint` lint job on every push and PR.

## License

MIT — see [LICENSE](LICENSE).
