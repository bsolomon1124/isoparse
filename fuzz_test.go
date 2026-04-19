package isoparse_test

import (
	"testing"

	"github.com/bsolomon1124/isoparse"
)

// FuzzParseDatetime asserts the no-panic invariant: ParseDatetime must return
// cleanly (either a Time or a *ParseError) for any input, regardless of how
// malformed. If a seed or generated input causes a panic, the fuzzer surfaces
// it as a failure.
func FuzzParseDatetime(f *testing.F) {
	seeds := []string{
		"",
		"2024",
		"2024-03-15",
		"2024-03-15T14:30:45Z",
		"2024-03-15T14:30:45.123456789+05:30",
		"2024-W11-5",
		"2024-075",
		"20240315T143045\u221205:00",
		"\x00\x00",
		"99999999",
		"\xff\xff\xff\xff\xff\xff\xff\xff",
		"2014-04-11T24:00:00",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = isoparse.ParseDatetime(s)
	})
}

// FuzzParseDate covers the no-time-component path.
func FuzzParseDate(f *testing.F) {
	seeds := []string{"", "2024", "2024-03-15", "20240315", "2024-W11-5", "2024-075", "\x00"}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _ = isoparse.ParseDate(s)
	})
}

// FuzzParseTime covers standalone time parsing, including the fractional
// second and timezone branches.
func FuzzParseTime(f *testing.F) {
	seeds := []string{"", "14", "14:30", "143045", "14:30:45.123456", "14:30:45Z", "14:30:45+05:30", "24:00"}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_, _, _ = isoparse.ParseTime(s)
	})
}
