package isoparse_test

import (
	"testing"

	"github.com/bsolomon1124/isoparse"
)

var benchDatetimes = []string{
	"2024-03-15T14:30:45Z",
	"2024-03-15T14:30:45.123456789+05:30",
	"20240315T143045",
	"2024-W11-5T14:30:45Z",
	"2024-075T14:30:45",
}

func BenchmarkParseDatetime(b *testing.B) {
	for _, s := range benchDatetimes {
		b.Run(s, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_, _ = isoparse.ParseDatetime(s)
			}
		})
	}
}

func BenchmarkParseDate(b *testing.B) {
	dates := []string{"2024-03-15", "20240315", "2024-W11-5", "2024-075"}
	for _, s := range dates {
		b.Run(s, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_, _ = isoparse.ParseDate(s)
			}
		})
	}
}

func BenchmarkParseTime(b *testing.B) {
	times := []string{"14:30:45", "143045", "14:30:45.123456", "14:30:45+05:30"}
	for _, s := range times {
		b.Run(s, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_, _, _ = isoparse.ParseTime(s)
			}
		})
	}
}
