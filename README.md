# isoparse

Package isoparse parses strings representing ISO-8601-conformant datetimes,
dates, and times into instances of Go's `time.Time`.

Most of its parsing logic is ported directly from the isoparser module,
authored by Paul Ganssle, within Python's dateutil library.

Unlike the parser in Go's time package, isoparse's functions do not require
the precise format to be specified in advance.

It does not make every attempt to mirror Python's dateutil exactly,
partially because of differences in behavior between Python and Go.

The isoparse package exports three parsing functions:

- `ParseISODatetime`: parses a datetime (combined date and time string). Note that this function can also parse just a date in isolation, but if the user knows that input strings contain only dates with no time components, it will be faster to use ParseISODate.
- `ParseISODate`: parses a date string with no time component.
- `ParseISOTime`: parses a time string with no date component. This does not return a time.Time instance, but rather the hour/minute/second/nsec components and the location.

## A Note On Time Zone Handling

Python's datetime has a concept of a naive datetime:

> A naive object does not contain enough information to unambiguously locate itself relative to other date/time objects.

This is useful in situations where, given a datetime string such as
"2018-09-27T11:52:59", it is left up to the user to determine which time
zone should added as an attribute, if any at all. Contrarily, Go's
time.Date, which produces a time.Time instance, has a required location
parameter whose nil value is UTC, and is an unexported struct field that
cannot be changed independently of changing the timestamp itself.

For that reason:

- All datetimes and times that lack a visible offset will have `time.Local` attached to them. This represents a "best assumption" that the datetime string is from the package user's local time zone.
- This package also exports a simple function `SetLoc` that produces a new `time.Time` given a different time zone but the same timestamp components.  This is different from Go's `time.Time.In`, `time.Time.UTC`, or `time.Time.Local` in that these conversions may change attributes such as `t.Hour` in the resulting timestamp itself.

Note also that input strings that do contain a recognizable UTC offset will
be given a loc that is the result of time.FixedZone, with the generic name
of "UTC" and a specified seconds-east offset from UTC. There is no attempt
to determine an IANA time zone by name because, for instance, an offset of
-05:00 is still ambiguous based on whether daylight savings time is in
effect or not.

Because `time.Time.String` uses:

```go
func (t Time) String() string {
	s := t.Format("2006-01-02 15:04:05.999999999 -0700 MST")
	...
}
```

The `time.Time` resulting from isoparse's parsing functions will have a loc that
looks like `time.FixedZone("UTC", secondsEast)` and will be printed as:

    YYYY-MM-DD HH:MM:SS.sssssssss -HHMM UTC

If you want more control over the actual resulting format, use
`time.Time.Format` on the result.

## Conformance And Nonconformance To ISO-8601

isoparse conforms mostly to the [December 2004 ISO Standard 8601](https://www.iso.org/standard/40874.html), which
cancels and replaces the second edition (ISO 8601:2000) with minor
revisions.

For a (non-exhaustive) list of supported formats, execute the example
function ExampleParseISODatetime from example_test.go to see a range of
supported format examples.

The following is a list of ways in which this the exported functions in this
package deviates from the ISO-8601:2004 standard:

- The standard is strict about "T" being the separator between date and time. This package allows any ASCII character except 0 thru 9 as the separator between date and time, rather than just "T".
- The standard allows years less than 0 and greater than 9999. This package only permits years greater than 0 and less than 10,000.
- This package does not support parsing time intervals or recurring time intervals as defined in sections 4.4 and 4.5 of the standard, respectively.
- The standard technically allows "19" to represent the date 1900-01-01, or "23" to represent the time 23:00:00, as "representation[s] with reduced accuracy." This package does not allow these formats.  (Although YYYY-MM and YYYY are valid here.)
- Unless otherwise note, this package does not support "expanded representations" for dates (sections 4.1.2.4, 4.1.3.3, 4.1.4.4).
- Representations that "are only allowed by mutual agreement of the partners in information exchange" are generally not valid under this package.
- Support for fractional components other than seconds is part of the ISO-8601 standard, but is not currently implemented in this parser.  (This follows Python's dateutil.) For instance (from Wikipedia): "To denote '14 hours, 30 and one half minutes,' do not include a seconds figure. Represent it as '14:30,5', '1430,5', '14:30.5', or '1430.5'."  These 4 datetime strings will return a ParseError from ParseISODatetime.


## Other Notes

In addition to following closely with dateutil's isoparser module, this
package also ports code from Python's native datetime module and Go's time
package.

## Exported Objects

```
func ParseISODate(dateString string) (time.Time, error)
func ParseISODatetime(datetime string) (time.Time, error)
func ParseISOTime(timeString string) (components [4]int, tz *time.Location, err error)
func SetLoc(t time.Time, loc *time.Location) time.Time
type ParseError struct{ ... }
```
