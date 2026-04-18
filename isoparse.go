// Copyright 2026 Brad Solomon. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

// Package isoparse parses ISO 8601 date, time, and datetime strings into
// [time.Time] values without requiring a format string.
//
// The three entry points are [ParseDatetime], [ParseDate], and [ParseTime].
// Parse failures return a [*ParseError]; inspect via [errors.As].
//
// Inputs that omit a UTC offset are returned in [time.Local]; inputs with a
// recognizable offset use [time.FixedZone] (no IANA name is inferred, since
// an offset like -05:00 is DST-ambiguous). See [SetLoc] for re-attaching a
// different [*time.Location] without shifting wall-clock components.
//
// Most of the parsing logic is ported from the isoparser module (by Paul
// Ganssle) in Python's dateutil library; semantics are adapted to Go's time
// package. See the project README for supported formats and deviations from
// the ISO 8601:2004 standard.
package isoparse

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"
)

const (
	dateSep = '-'
	timeSep = ':'
	// Inclusive bounds for each date/time unit.
	minYear    = 1
	maxYear    = 9999
	minMonth   = 1
	maxMonth   = 12
	minHour    = 0
	maxHour    = 24 // 24:00 as midnight is valid
	minMin     = 0
	maxMin     = 59
	minSec     = 0
	maxSec     = 59
	minNsec    = 0
	maxNsec    = 999_999_999
	minISOWeek = 1
	maxISOWeek = 53
	minISODay  = 1
	maxISODay  = 7
)

// fractionRegex matches a period or comma followed by one or more digits —
// the optional fraction portion of an ISO time. It is the only regexp in use.
var fractionRegex = regexp.MustCompile(`[.,](?P<digits>[0-9]+)`)

// dim indexes days-in-month by 1-based month number; index 0 is a placeholder.
var dim = [13]int{-1, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// dbm indexes days-before-month (non-leap) by 1-based month number.
var dbm = [13]int{
	-1,
	0,
	31,
	59,
	90,
	120,
	151,
	181,
	212,
	243,
	273,
	304,
	334,
}

// strictDate returns a [time.Time] with no normalization: each unit must fall
// independently in its valid range. This differs from [time.Date], which
// overflows out-of-range values into the next larger unit.
func strictDate(year int, month time.Month, day, hour, minute, sec, nsec int, loc *time.Location) (time.Time, error) {
	mkErr := func(msg string) error {
		return &ParseError{
			Datetime: fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:%02d.%09d%v", year, month, day, hour, minute, sec, nsec, loc),
			Message:  msg,
		}
	}
	if year < minYear || year > maxYear {
		return time.Time{}, mkErr("year out of valid range")
	}
	if month < minMonth || month > maxMonth {
		return time.Time{}, mkErr("month out of valid range")
	}
	if day > daysInMonth(year, month) {
		return time.Time{}, mkErr("day out of valid range")
	}
	if hour < minHour || hour > maxHour {
		// Hour 24 is allowed; callers convert via time.Date rollover.
		return time.Time{}, mkErr("hour out of valid range")
	}
	if minute < minMin || minute > maxMin {
		return time.Time{}, mkErr("minute out of valid range")
	}
	if sec < minSec || sec > maxSec {
		return time.Time{}, mkErr("second out of valid range")
	}
	if nsec < minNsec || nsec > maxNsec {
		return time.Time{}, mkErr("nanosecond out of valid range")
	}
	// A nil *time.Location would make time.Date panic; default to Local.
	if loc == nil {
		loc = time.Local
	}
	return time.Date(year, month, day, hour, minute, sec, nsec, loc), nil
}

// btoi converts a bool to 0 or 1.
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isLeapYear reports whether year is a Gregorian leap year (ISO 8601:2004 §3.2.1).
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// daysInMonth returns the number of days in month for year (accounting for leap years).
func daysInMonth(year int, month time.Month) int {
	if isLeapYear(year) && month == time.February {
		return 29
	}
	return dim[month]
}

// daysBeforeYear returns the number of days preceding January 1 of year
// (year 1 is the epoch; returns 0).
func daysBeforeYear(year int) int {
	y := year - 1
	return y*365 + y/4 - y/100 + y/400
}

// daysBeforeMonth returns the number of days in year preceding the first of month.
func daysBeforeMonth(year int, month time.Month) int {
	if month > 2 && isLeapYear(year) {
		return dbm[month] + 1
	}
	return dbm[month]
}

// ymdToOrd converts a (year, month, day) tuple to its ordinal day count from year 1.
func ymdToOrd(year int, month time.Month, day int) int {
	return daysBeforeYear(year) + daysBeforeMonth(year, month) + day
}

// isoWeekday returns the day of the week in ISO numbering: Monday=1 .. Sunday=7.
func isoWeekday(date time.Time) int {
	year, month, day := date.Date()
	ordinal := ymdToOrd(year, month, day)
	wd := ordinal % 7
	if wd == 0 {
		wd = 7
	}
	return wd
}

// isoCalendar returns (ISO year, ISO week, ISO weekday). Built on [time.Time.ISOWeek].
func isoCalendar(date time.Time) [3]int {
	isoyear, isoweek := date.ISOWeek()
	return [3]int{isoyear, isoweek, isoWeekday(date)}
}

// calcWeekdate resolves an ISO year-week-day triple to the corresponding
// Gregorian date. Ported from Python's dateutil.
func calcWeekdate(year, week, day int) (time.Time, error) {
	if week < minISOWeek || week > maxISOWeek {
		return time.Time{}, &ParseError{
			Datetime: fmt.Sprintf("%04d-%02d-%02d", year, week, day),
			Message:  "invalid ISO week",
		}
	}
	if day < minISODay || day > maxISODay {
		return time.Time{}, &ParseError{
			Datetime: fmt.Sprintf("%04d-%02d-%02d", year, week, day),
			Message:  "invalid ISO day",
		}
	}
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	week1 := jan4.AddDate(0, 0, -1*(isoWeekday(jan4)-1))
	return week1.AddDate(0, 0, (week-1)*7+(day-1)), nil
}

// ParseError describes a failure to parse a date, time, or datetime string.
// It is the sole error type returned by the Parse* functions in this package;
// use [errors.As] to inspect.
type ParseError struct {
	Datetime string // the offending input
	Message  string // optional; a specific reason
}

// Error implements the [error] interface.
func (e *ParseError) Error() string {
	if e.Message == "" {
		return "cannot parse " + e.Datetime
	}
	return "cannot parse " + e.Datetime + ": " + e.Message
}

// parseDateCommon parses the common ISO 8601 date forms (YYYY, YYYY-MM,
// YYYY-MM-DD, YYYYMMDD) as a fastpath. On success, pos is the index just
// past the consumed date portion.
func parseDateCommon(dateString string) (components [3]int, pos int, err error) {
	length := len(dateString)
	if length < 4 {
		return components, pos, &ParseError{Datetime: dateString, Message: "date string too short"}
	}
	components = [3]int{1, 1, 1}
	components[0], _ = strconv.Atoi(dateString[:4])
	pos = 4
	if pos >= length {
		// Just YYYY, valid (becomes YYYY-01-01).
		return components, pos, nil
	}

	hasSep := dateString[pos] == dateSep
	pos += btoi(hasSep)

	// Remaining forms: MM-DD, MMDD, or MM.
	if length-pos < 2 {
		return components, pos, &ParseError{Datetime: dateString, Message: "invalid month"}
	}

	// May mis-parse a slice of YYYYDDD as the month; the Atoi error below and
	// the day parse further on are what force us to defer to parseDateUncommon.
	components[1], err = strconv.Atoi(dateString[pos : pos+2])
	pos += 2
	if err != nil {
		return components, pos, &ParseError{Datetime: dateString, Message: "invalid month"}
	}
	if pos >= length {
		if hasSep {
			// Day was omitted; defaults to 1.
			return components, pos, nil
		}
		// Something like "177607" — rejected to avoid confusion with YYMMDD.
		return components, pos, &ParseError{Datetime: dateString, Message: "invalid format"}
	}

	if hasSep {
		if dateString[pos] != dateSep {
			return components, pos, &ParseError{Datetime: dateString, Message: "invalid separator"}
		}
		pos++
	}

	if length-pos < 2 {
		return components, pos, &ParseError{Datetime: dateString, Message: "invalid common day"}
	}
	components[2], err = strconv.Atoi(dateString[pos : pos+2])
	if err != nil {
		// Failing here (e.g. "1985102") lets parseDateUncommon pick it up.
		return components, pos, &ParseError{Datetime: dateString, Message: "invalid day"}
	}
	return components, pos + 2, nil
}

// parseDateUncommon parses the less common ISO 8601 date forms — ISO week
// dates (YYYY-Www-D, YYYYWwwD, YYYY-Www, YYYYWww) and ordinal dates
// (YYYY-DDD, YYYYDDD). Called only after parseDateCommon fails.
func parseDateUncommon(dateString string) (components [3]int, pos int, err error) {
	length := len(dateString)
	if length < 4 {
		return components, pos, &ParseError{Datetime: dateString, Message: "date string too short"}
	}
	var t time.Time
	year, _ := strconv.Atoi(dateString[:4])
	pos = 4
	hasSep := length > pos && dateString[pos] == dateSep
	pos += btoi(hasSep)
	if pos >= length {
		return components, pos, &ParseError{Datetime: dateString, Message: "incomplete date"}
	}

	if dateString[pos] == 'W' {
		// Week date: Www, Www-D, or WwwD.
		pos++
		if pos+2 > length {
			return components, pos, &ParseError{Datetime: dateString, Message: "invalid ISO week"}
		}
		weekNum, _ := strconv.Atoi(dateString[pos : pos+2])
		pos += 2
		dayNum := 1
		if length > pos {
			if (dateString[pos] == dateSep) != hasSep {
				return components, pos, &ParseError{Datetime: dateString, Message: "inconsistent separator"}
			}
			if hasSep {
				pos++
			}
			if pos >= length {
				return components, pos, &ParseError{Datetime: dateString, Message: "invalid ISO day"}
			}
			dayNum, _ = strconv.Atoi(dateString[pos : pos+1])
			pos++
		}
		t, err = calcWeekdate(year, weekNum, dayNum)
		if err != nil {
			return components, pos, err
		}
	} else {
		// Ordinal date: YYYYDDD or YYYY-DDD.
		if length-pos < 3 {
			return components, pos, &ParseError{Datetime: dateString, Message: "invalid ordinal day"}
		}
		if length-pos == 4 {
			// Catch YYYY-MMDD / YYYYM-MDD style mismatches.
			if hasSep && dateString[length-3] != dateSep {
				return components, pos, &ParseError{Datetime: dateString, Message: "inconsistent separator"}
			}
			if !hasSep && dateString[length-3] == dateSep {
				return components, pos, &ParseError{Datetime: dateString, Message: "inconsistent separator"}
			}
		}
		ordinalDay, _ := strconv.Atoi(dateString[pos : pos+3])
		pos += 3
		if ordinalDay < 1 || ordinalDay > (365+btoi(isLeapYear(year))) {
			return components, pos, &ParseError{Datetime: dateString, Message: "invalid ordinal day for given year"}
		}
		t = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, ordinalDay-1)
	}
	components = [3]int{t.Year(), int(t.Month()), t.Day()}
	return components, pos, nil
}

// parseDate tries the common fastpath then falls back to the uncommon parser.
// Returns (Y, M, D) components and the cursor position, so the caller can
// tell whether the input continues with a time portion.
func parseDate(dateString string) (components [3]int, pos int, err error) {
	components, pos, err = parseDateCommon(dateString)
	if err != nil {
		components, pos, err = parseDateUncommon(dateString)
		if err != nil {
			return components, pos, err
		}
	}
	return components, pos, nil
}

// ParseDate parses an ISO 8601 date string (no time component) into a
// [time.Time] with [time.Local] as its location.
func ParseDate(dateString string) (time.Time, error) {
	components, pos, err := parseDate(dateString)
	if err != nil {
		return time.Time{}, err
	}
	if pos < len(dateString) {
		return time.Time{}, &ParseError{Datetime: dateString, Message: "string contains unknown iso components"}
	}
	return strictDate(components[0], time.Month(components[1]), components[2], 0, 0, 0, 0, time.Local)
}

// parseTimezone parses an ISO 8601 timezone string: Z, ±HH, ±HHMM, or ±HH:MM.
// Accepts ASCII + / - and also Unicode minus-sign (U+2212).
func parseTimezone(tzString string) (tz *time.Location, err error) {
	if tzString[0] == 'Z' {
		return time.UTC, nil
	}

	// Except for Z, a leading sign is required. The sign may be ASCII + / -
	// or Unicode U+2212 (multi-byte in UTF-8).
	var (
		mult    int
		signLen int
	)
	if r, size := utf8.DecodeRuneInString(tzString); r == '+' {
		mult, signLen = 1, size
	} else if r == '-' || r == '\u2212' {
		mult, signLen = -1, size
	} else {
		return tz, &ParseError{Datetime: tzString, Message: "unrecognized timezone sign"}
	}

	body := tzString[signLen:]
	length := len(body)
	switch length {
	case 2, 4, 5:
		// OK: HH, HHMM, or HH:MM
	default:
		return time.Local, &ParseError{Datetime: tzString, Message: "time zone offset must be Z, ±HH, ±HHMM or ±HH:MM"}
	}

	hours, _ := strconv.Atoi(body[:2])
	var minutes int
	if length != 2 {
		if body[2] == ':' {
			minutes, _ = strconv.Atoi(body[3:])
		} else {
			minutes, _ = strconv.Atoi(body[2:])
		}
	}

	if hours == 0 && minutes == 0 {
		return time.UTC, nil
	}
	if hours < minHour || hours > maxHour || minutes < minMin || minutes > maxMin {
		return time.Local, &ParseError{Datetime: tzString, Message: "offset component out of valid range"}
	}

	// Name is generic because an offset alone cannot identify an IANA zone
	// (DST ambiguity); callers can attach a specific zone via SetLoc.
	return time.FixedZone("UTC", mult*60*(hours*60+minutes)), nil
}

// ParseTime parses an ISO 8601 time string (no date component) and returns
// the component ints [hour, minute, second, nanosecond] along with the
// location. Accepted forms include HH, HH:MM, HHMM, HH:MM:SS, HHMMSS, and
// any of the above with a fractional seconds suffix and/or timezone suffix.
func ParseTime(timeString string) (components [4]int, tz *time.Location, err error) {
	tz = time.Local
	length := len(timeString)
	pos, comp := 0, -1

	if length < 2 {
		return components, tz, &ParseError{Datetime: timeString, Message: "length of time string must be >= 2"}
	}
	hasSep := length >= 3 && timeString[2] == timeSep

	// Fractional components of hour or minute are not supported (see README).

	for pos < length && comp < 4 {
		comp++

		if start := timeString[pos]; start == 'Z' || start == '+' || start == '-' {
			tz, err = parseTimezone(timeString[pos:])
			if err != nil {
				return components, tz, err
			}
			pos = length
			break
		}

		if comp < 3 {
			if pos+2 > length {
				return components, tz, &ParseError{Datetime: timeString, Message: "incomplete time component"}
			}
			components[comp], _ = strconv.Atoi(timeString[pos : pos+2])
			pos += 2
			if hasSep && pos < length && timeString[pos] == timeSep {
				pos++
			}
		}

		if comp == 3 {
			frac := fractionRegex.FindStringSubmatch(timeString[pos:])
			if frac == nil {
				continue
			}
			// Truncate (not round) to 9 digits — Go's nanosecond precision.
			secondsFrac, _ := strconv.ParseFloat("0."+frac[1][:min(9, len(frac[1]))], 64)
			components[comp] = int(secondsFrac * 1e9)
			pos += len(frac[0])
		}
	}

	if pos < length {
		return components, tz, &ParseError{Datetime: timeString, Message: "unused components"}
	}

	if components[0] == 24 {
		// 24:00 is a valid midnight representation, but no sub-hour units may be nonzero.
		for _, i := range components[1:] {
			if i != 0 {
				return components, tz, &ParseError{Datetime: timeString, Message: "hour == 24 implies 0 for other time units"}
			}
		}
		// The caller (via time.Date rollover) converts hour 24 → next-day 00:00.
	}
	return components, tz, nil
}

// ParseDatetime parses an ISO 8601 datetime (combined date and time). The
// date/time separator may be any non-numeric ASCII character, not strictly "T".
//
// ParseDatetime also accepts date-only input, but [ParseDate] is faster when
// the caller knows no time portion is present.
//
// Behaves like [time.Parse] but without a layout argument: on error, the
// returned [time.Time] is the zero value.
func ParseDatetime(datetime string) (time.Time, error) {
	dateParts, pos, err := parseDate(datetime)
	if err != nil {
		return time.Time{}, err
	}

	var (
		hour, minute, second, nsec int
		tz                         *time.Location // nil; populated below
	)

	switch {
	case len(datetime) > pos:
		// Separator must be a non-numeric ASCII byte (bytes [0,127] minus '0'..'9').
		sep := datetime[pos]
		if sep > 127 || (sep >= '0' && sep <= '9') {
			return time.Date(1, 1, 1, 0, 0, 0, 0, time.Local),
				&ParseError{Datetime: datetime, Message: "date/time separator must be a non-numeric ASCII character"}
		}
		var timeParts [4]int
		timeParts, tz, err = ParseTime(datetime[pos+1:])
		if err != nil {
			// Return a non-nil-loc sentinel so callers inspecting the Time don't panic.
			return time.Date(1, 1, 1, 0, 0, 0, 0, time.Local), err
		}
		hour, minute, second, nsec = timeParts[0], timeParts[1], timeParts[2], timeParts[3]
	case len(datetime) < pos:
		// Defensive: unreachable under current parseDate behavior.
		return time.Time{}, &ParseError{Datetime: datetime}
	}
	return strictDate(dateParts[0], time.Month(dateParts[1]), dateParts[2], hour, minute, second, nsec, tz)
}

// SetLoc returns a new [time.Time] with the same wall-clock components
// (year, month, day, hour, minute, second, nanosecond) as t, but with loc
// as its [*time.Location].
//
// Unlike [time.Time.In], SetLoc does not preserve the instant: it reinterprets
// the same wall-clock reading in loc. Use this when an input like
// "2024-03-15T10:00:00" was meant to be in, say, America/New_York rather
// than the machine's [time.Local].
func SetLoc(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}
