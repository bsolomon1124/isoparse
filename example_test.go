package isoparse_test

import (
	"fmt"
	"time"

	"github.com/bsolomon1124/isoparse"
)

func ExampleParseDatetime() {
	t, err := isoparse.ParseDatetime("2024-03-15T14:30:45Z")
	if err != nil {
		panic(err)
	}
	fmt.Println(t.Format(time.RFC3339Nano))
	// Output: 2024-03-15T14:30:45Z
}

func ExampleParseDatetime_offset() {
	// Non-zero offset → time.FixedZone.
	t, _ := isoparse.ParseDatetime("2024-03-15T10:00:00-05:00")
	fmt.Println(t.Format(time.RFC3339))
	// Output: 2024-03-15T10:00:00-05:00
}

func ExampleParseDatetime_basic() {
	// Compact (no-separator) representation.
	t, _ := isoparse.ParseDatetime("20240315T143045Z")
	fmt.Println(t.Format(time.RFC3339))
	// Output: 2024-03-15T14:30:45Z
}

func ExampleParseDate() {
	t, _ := isoparse.ParseDate("2024-03-15")
	fmt.Println(t.Format("2006-01-02"))
	// Output: 2024-03-15
}

func ExampleParseDate_weekDate() {
	// ISO week date: week 11, day 5 of 2024.
	t, _ := isoparse.ParseDate("2024-W11-5")
	fmt.Println(t.Format("2006-01-02"))
	// Output: 2024-03-15
}

func ExampleParseDate_ordinal() {
	// Ordinal date: the 75th day of 2024.
	t, _ := isoparse.ParseDate("2024-075")
	fmt.Println(t.Format("2006-01-02"))
	// Output: 2024-03-15
}

func ExampleParseTime() {
	components, _, _ := isoparse.ParseTime("14:30:45.5Z")
	fmt.Printf("h=%d m=%d s=%d ns=%d\n", components[0], components[1], components[2], components[3])
	// Output: h=14 m=30 s=45 ns=500000000
}

func ExampleSetLoc() {
	// Parsing "2024-03-15T10:00:00" yields a Time in time.Local. If the input
	// was actually meant to be in a specific zone, SetLoc re-attaches that
	// zone without shifting the wall-clock components (unlike Time.In).
	t, _ := isoparse.ParseDatetime("2024-03-15T10:00:00")
	ny, _ := time.LoadLocation("America/New_York")
	t = isoparse.SetLoc(t, ny)
	fmt.Println(t.Format(time.RFC3339))
	// Output: 2024-03-15T10:00:00-04:00
}
