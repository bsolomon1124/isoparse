// Use of this source code is governed by Apache License, Version 2.0, that can be found
// in the LICENSE file.

// Package isoparse parses strings representing ISO-8601-conformant datetimes, dates, and times
// into instances of Go's time.Time.
//
// Most of its parsing logic is ported directly from the isoparser module, authored by Paul Ganssle,
// within Python's dateutil library.
//
// Unlike the parser in Go's time package, isoparser's functions do not require the precise format
// to be specified in advance.
//
// It does not make every attempt to mirror Python's dateutil exactly, partially
// because of differences in behavior between Python and Go.
//
// The isoparse package exports three parsing functions:
//
// -	ParseISODatetime: parses a datetime (combined date and time string).  Note that this
// 		function can also parse just a date in isolation, but if the user knows that input strings
// 		contain only dates with no time components, it will be faster to use ParseISODate.
// -	ParseISODate: parses a date string with no time component.
// -	ParseISOTime: parses a time string with no date component.  This does not return a
// 		time.Time instance, but rather the hour/minute/second/nsec components and the location.
//
// A Note On Time Zone Handling
//
// Python's datetime has a concept of a naive datetime:
//
//		A naive object does not contain enough information to unambiguously locate itself relative
//		to other date/time objects.
//
// This is useful in situations where, given a datetime string such as "2018-09-27T11:52:59",
// it is left up to the user to determine which time zone should added as an attribute, if any at
// all.  Contrarily, Go's time.Date, which produces a time.Time instance, has a required location
// parameter whose nil value is UTC, and is an unexported struct field that cannot be changed
// independently of changing the timestamp itself.
//
// For that reason:
//
// -	All datetimes and times that lack a visible offset will have time.Local attached to them.
// 		This represents a "best assumption" that the datetime string is from the package
// 		user's local time zone.
// -	This package also exports a simple function SetLoc that produces a new time.Time given
// 		a different time zone but the same timestamp components.  This is different from
// 		Go's time.Time.In, time.Time.UTC, or time.Time.Local in that these conversions may
// 		change attributes such as t.Hour() in the resulting timestamp itself.
//
// Note also that input strings that do contain a recognizable UTC offset will be given a
// loc that is the result of time.FixedZone, with the generic name of "UTC" and a specified
// seconds-east offset from UTC.  There is no attempt to determine an IANA time zone by name because,
// for instance, an offset of -05:00 is still ambiguous based on whether daylight savings time
// is in effect or not.
//
// Because time.Time.String() uses:
//
//		func (t Time) String() string {
//			s := t.Format("2006-01-02 15:04:05.999999999 -0700 MST")
//			...
//		}
//
// The time.Time results from isoparse's parsing functions will have a loc that looks like
// time.FixedZone("UTC", secondsEast) and will be printed as:
//
//		YYYY-MM-DD HH:MM:SS.sssssssss -HHMM UTC
//
// If you want more control over the actual resulting format, use time.Time.Format.
//
// Conformance And Nonconformance To ISO-8601
//
// isoparse conforms mostly to the December 2004 ISO Standard 8601 which cancels and replaces
// the second edition (ISO 8601:2000) with minor revisions. https://www.iso.org/standard/40874.html
//
// For a (non-exhaustive) list of supported formats, execute the example function
// ExampleParseISODatetime from example_test.go to see a range of supported format examples.
//
// The following is a list of ways in which this the exported functions in this package
// deviates from the ISO-8601:2004 standard:
//
// -	The standard is strict about "T" being the separator between date and time.
// 		This package allows any ASCII character except 0 thru 9 as the separator
// 		between date and time, rather than just "T".
// -	The standard allows years less than 0 and greater than 9999.
// 		This package only permits years greater than 0 and less than 10,000.
// -	This package does not support parsing time intervals or recurring time intervals
// 		as defined in sections 4.4 and 4.5 of the standard, respectively.
// -	The standard technically allows "19" to represent the date 1900-01-01, or "23" to
// 		represent the time 23:00:00, as "representation[s] with reduced accuracy."
// 		This package does not allow these formats.  (Although YYYY-MM and YYYY are valid here.)
// -	Unless otherwise note, this package does not support "expanded representations" for
// 		dates (sections 4.1.2.4, 4.1.3.3, 4.1.4.4).
// -	Representations that "are only allowed by mutual agreement of the partners in
// 		information exchange" are generally not valid under this package.
// -	Support for fractional components other than seconds is part of the ISO-8601 standard,
// 		but is not currently implemented in this parser.  (This follows Python's dateutil.)
// 		For instance (from Wikipedia): "To denote '14 hours, 30 and one half minutes,'
// 		do not include a seconds figure. Represent it as '14:30,5', '1430,5', '14:30.5', or
// 		'1430.5'."  These 4 datetime strings will return a ParseError from ParseISODatetime.
//
// Other Notes
//
// In addition to following closely with dateutil's isoparser module, this package also ports code
// from Python's native datetime module and Go's time package.
package isoparse

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	dateSep = '-'
	timeSep = ':'
	// These mins & maxs are inclusive on both bounds.
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
	maxNsec    = int(1e9 - 1)
	minISOWeek = 1
	maxISOWeek = 53
	minISODay  = 1
	maxISODay  = 7
)

// Period, or comma, followed by 1 or more digits.
// This is used to grab the optional fraction portion of the time.
// It is the only regexp used in this package.
var fractionRegex = regexp.MustCompile(`[.,](?P<digits>[0-9]+)`)

// Days in month.  -1 is a placeholder because calendars are more intuitively 1-indexed.
var dim = [13]int{-1, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// Days before month.  There is a similar var in Go's time,
// but this more closely follows Python's datetime design and the helper functions here.
var dbm = [13]int{
	-1,
	0,
	31,  // 31,
	59,  // 31 + 28,
	90,  // 31 + 28 + 31,
	120, // 31 + 28 + 31 + 30,
	151, // 31 + 28 + 31 + 30 + 31,
	181, // 31 + 28 + 31 + 30 + 31 + 30,
	212, // 31 + 28 + 31 + 30 + 31 + 30 + 31,
	243, // 31 + 28 + 31 + 30 + 31 + 30 + 31 + 31,
	273, // 31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30,
	304, // 31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31,
	334, // 31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30,
}

// Helper functions
// These follow closely with both Python's datetime and Go's time modules.
// We use the time.Month type (just an int) wherever possible for consistency.

// Go's time.Date "normalizes" values, overflowing into the next smallest unit.
// (Providing month=1, day=32 normalizes to month=2, day=1.)
// This package is more strict: if the input string doesn't itself form a valid date, don't attempt to reconform it.
// Each unit must be strictly in its independently defined range.
func strictDate(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) (time.Time, error) {
	if year < minYear || year > maxYear {
		datetime := fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:%02d.%09d%v", year, month, day, hour, min, sec, nsec, loc)
		return time.Time{}, &ParseError{datetime, "year out of valid range"}
	}
	if month < minMonth || month > maxMonth {
		datetime := fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:%02d.%09d%v", year, month, day, hour, min, sec, nsec, loc)
		return time.Time{}, &ParseError{datetime, "month out of valid range"}
	}
	if day > daysInMonth(year, month) {
		datetime := fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:%02d.%09d%v", year, month, day, hour, min, sec, nsec, loc)
		return time.Time{}, &ParseError{datetime, "day out of valid range"}
	}
	if hour < minHour || hour > maxHour {
		// We do *not* handle the 24:00 -> midnight aspect here.  Hour may be 24.
		datetime := fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:%02d.%09d%v", year, month, day, hour, min, sec, nsec, loc)
		return time.Time{}, &ParseError{datetime, "hour out of valid range"}
	}
	if min < minMin || min > maxMin {
		datetime := fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:%02d.%09d%v", year, month, day, hour, min, sec, nsec, loc)
		return time.Time{}, &ParseError{datetime, "minute out of valid range"}
	}
	if sec < minSec || sec > maxSec {
		datetime := fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:%02d.%09d%v", year, month, day, hour, min, sec, nsec, loc)
		return time.Time{}, &ParseError{datetime, "second out of valid range"}
	}
	if nsec < minNsec || nsec > maxNsec {
		datetime := fmt.Sprintf("%02d-%02d-%02dT%02d:%02d:%02d.%09d%v", year, month, day, hour, min, sec, nsec, loc)
		return time.Time{}, &ParseError{datetime, "nanosecond out of valid range"}
	}

	// We need to be careful with the fact that time.UTC != nil, but the zero value for
	// *time.Location will be represented as UTC
	if loc == nil {
		loc = time.Local
	}

	// We can't validate the hours/minutes on loc here because there are unexported
	// fields of Location.  That checking is performed in parseTimezone
	return time.Date(year, month, day, hour, min, sec, nsec, loc), nil
}

// Bool to int
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

// isLeapYear tests whether a given year is a leap year.
// A leap year is a year whose year number is divisible by four an integral number of times.
// However, a centennial year is not a leap year unless its year number is divisible
// by four hundred an integral number of times. (ISO 8601:2004 3.2.1)
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// daysInMonth returns the number of days in a given month for a particular year.
// `year` is required because the result varies based on whether the year is a leap year.
func daysInMonth(year int, month time.Month) int {
	if isLeapYear(year) && month == time.February {
		return 29
	}
	return dim[month]
}

// daysBeforeYear calculates the number of days before January 1st of `year`.
//
// Taken from Python's datetime module.
// For example, there are 736694 before the year 2018.  There are, by definition, 0 days before
// year 1.  (Minimum date is 1/1/1 for our purposes.)
func daysBeforeYear(year int) int {
	// See also: http://www.staff.science.uu.nl/~gent0113/calendar/isocalendar.htm
	y := year - 1
	return y*365 + y/4 - y/100 + y/400
}

// daysBeforeMonth calculates the number of days in the year preceding the first day of the given month.
//
// Taken from Python's datetime module.
func daysBeforeMonth(year int, month time.Month) int {
	addon := month > 2 && isLeapYear(year)
	if addon {
		return dbm[month] + 1
	}
	return dbm[month] + 0
}

// ymdToOrder converts a (year, month, day) tuple to its ordinal equivalent.
//
// Taken from Python's datetime module.
func ymdToOrd(year int, month time.Month, day int) int {
	return daysBeforeYear(year) + daysBeforeMonth(year, month) + day
}

// isoWeekday returns the day of the week, where Monday == 1 ... Sunday == 7.
func isoWeekday(date time.Time) int {
	year, month, day := date.Date()
	ordinal := ymdToOrd(year, month, day)
	isoweekday := ordinal % 7
	if isoweekday == 0 {
		isoweekday = 7
	}
	return isoweekday
}

// isoCalendar returns a 3-tuple of (ISO year, ISO week number, and ISO weekday).
// This relies on Go's own `func (time.Time) ISOWeek`.
// This deviates from Python's method because ISOWeek is already available via Go time.
func isoCalendar(date time.Time) [3]int {
	isoyear, isoweek := date.ISOWeek()

	// We have the ISO 8601 year and week
	// Now we just need the day of the week, where Monday=1 ... Sunday=7
	isoweekday := isoWeekday(date)
	return [3]int{isoyear, isoweek, isoweekday}
}

// calcWeekdate calculates the Proleptic Gregorian calendar day corresponding to the
// given ISO year-week-day tuple.
//
// Ported directly from the Python dateutil package.
func calcWeekdate(year, week, day int) (time.Time, error) {
	if week < minISOWeek || week > maxISOWeek {
		dateString := fmt.Sprintf("%04d-%02d-%02d", year, week, day)
		return time.Time{}, &ParseError{dateString, "invalid ISO week"}
	} else if day < minISODay || day > maxISODay {
		dateString := fmt.Sprintf("%04d-%02d-%02d", year, week, day)
		return time.Time{}, &ParseError{dateString, "invalid ISO day"}
	}
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.Local)
	week1 := jan4.AddDate(0, 0, -1*(isoWeekday(jan4)-1))
	weekOffset := (week-1)*7 + (day - 1)
	return week1.AddDate(0, 0, weekOffset), nil
}

// ParseError describes any problem parsing a datetime, date, or time string.
// It is the sole error exported by this package.
// (It also exists with similar structure in Go's time package.)
type ParseError struct {
	Datetime string // This should always be passed
	Message  string // Treat as optional unless the reason is specific
}

func (e *ParseError) Error() string {
	if e.Message == "" {
		return "cannot parse " + e.Datetime
	}
	return "cannot parse " + e.Datetime + ": " + e.Message
}

// parseIsoDateCommon parses common-format ISO-8601 date strings (no time portion).
// Examples: YYYY-MM-DD, YYYYMMDD, YYYY, YYYY-MM.
//
// It attempts to take a fast route for more common ISO-8601 date strings.
//
// `components` is a [3]int of (year, month, day).
// `pos` is the position of the "cursor" that has parsed through the string.
// It is used in the exported function ParseISODatetime to determine if a time portion is present.
//
// Note: this returns simple ints, *not* time.Month instances.  Careful with comparison.
func parseISODateCommon(dateString string) (components [3]int, pos int, err error) {
	length := len(dateString)
	if length < 4 {
		// The shortest string we should possibly have is YYYY.
		return components, pos, &ParseError{dateString, "date string too short"}
	}
	components = [3]int{1, 1, 1}
	components[0], _ = strconv.Atoi(dateString[:4])
	pos = 4
	if pos >= length {
		// We received just YYYY, which is valid and becomes YYYY-01-01.
		return components, pos, nil
	}

	// Advance forward 1 position and ignore separator if it exists.
	hasSep := dateString[pos] == dateSep
	pos += btoi(hasSep)

	// At this point we are left with one of the following: MM-DD, MMDD, MM
	if length-pos < 2 {
		return components, pos, &ParseError{dateString, "invalid month"}
	}

	// Note that this *may* incorrectly pick up on a portion of YYYYDDD as the month.
	// But will then raise later on.
	components[1], err = strconv.Atoi(dateString[pos : pos+2])
	// This is one place where we definitely need to check the error.
	// It is what allows us to catch "2004W537" and defer it to parseISODateUncommon.
	pos += 2
	if err != nil {
		return components, pos, &ParseError{dateString, "invalid month"}
	}
	if pos >= length {
		if hasSep {
			// Day was not given; it will default to 1
			return components, pos, nil
		} else {
			// We have something like 177607, which is invalid
			// (Designed to avoid confusion with truncated representation YYMMDD still often used)
			return components, pos, &ParseError{dateString, "invalid format"}
		}
	}

	if hasSep {
		if dateString[pos] != dateSep {
			// Separator must be consistent.
			return components, pos, &ParseError{dateString, "invalid separator"}
		}
		pos += 1
	}

	// Day
	if length-pos < 2 {
		return components, pos, &ParseError{dateString, "invalid common day"}
	}
	components[2], err = strconv.Atoi(dateString[pos : pos+2])
	if err != nil {
		// Again, check the success of the conversion to make sure things like YYYYDDD fail here.
		// (And get picked up by parseISODateUncommon.)  We have may otherwise parsed the
		// month as the first two DD characters, and without this check 1985102 gets detected
		// as 1985-10-0.
		return components, pos, &ParseError{dateString, "invalid day"}
	}
	return components, pos + 2, nil
}

// parseIsoDateUncommon parses common-format ISO-8601 date strings (no time portion).
// Examples: YYYY-Www, YYYYWww, YYYY-Www-D, YYYYWwwD, YYYYDDD,  YYYY-DDD
//
// This function is called after the fast-route parseISODateCommon is exhausted.
//
// `components` is a [3]int of (year, month, day).
// `pos` is the position of the "cursor" that has parsed through the string.
// It is used in the exported function ParseISODatetime to determine if a time portion is present.
//
// Note: this returns simple ints, *not* time.Month instances.  Careful with comparison.
func parseISODateUncommon(dateString string) (components [3]int, pos int, err error) {
	// We are running through the entire string again here, starting at beginning.
	// The tradeoff is that parseISODateCommon is a fastpath that should handle most cases.
	length := len(dateString)
	if length < 4 {
		return components, pos, &ParseError{dateString, "date string too short"}
	}
	var t time.Time
	year, _ := strconv.Atoi(dateString[:4])
	pos = 4
	hasSep := dateString[pos] == dateSep
	pos += btoi(hasSep)

	// We have now moved past YYYY or YYYY-
	if dateString[pos] == 'W' {
		// Choose from Www, Www-D, or WwwD
		pos += 1
		weekNum, _ := strconv.Atoi(dateString[pos : pos+2])
		pos += 2
		dayNum := 1
		if length > pos {
			if (dateString[pos] == dateSep) != hasSep {
				// Prevent things like YYYY-MMDD (either use sep, or don't)
				return components, pos, &ParseError{dateString, "inconsistent separator"}
			}
			if hasSep {
				pos += 1
			}
			dayNum, _ = strconv.Atoi(dateString[pos : pos+1])
			pos += 1
		}
		t, err = calcWeekdate(year, weekNum, dayNum)
		if err != nil {
			return components, pos, err
		}
	} else {
		// Ordinal dates, YYYYDDD or YYYY-DDD (already at DDD)
		if length-pos < 3 {
			return components, pos, &ParseError{dateString, "invalid ordinal day"}
		}
		if length-pos == 4 {
			// First prevent things like YYYY-MMDD (either use sep, or don't)
			if hasSep && dateString[length-3] != dateSep {
				return components, pos, &ParseError{dateString, "inconsistent separator"}
			} else if !hasSep && dateString[length-3] == dateSep {
				// Vice-versa
				return components, pos, &ParseError{dateString, "inconsistent separator"}
			}
		}
		ordinalDay, _ := strconv.Atoi(dateString[pos : pos+3])
		pos += 3
		if ordinalDay < 1 || ordinalDay > (365+btoi(isLeapYear(year))) {
			return components, pos, &ParseError{dateString, "invalid ordinal day for given year"}
		}
		t = time.Date(year, 1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, ordinalDay-1)
	}
	components = [3]int{t.Year(), int(t.Month()), t.Day()}
	return components, pos, nil
}

// parseISODate parses an ISO-8601 date string with no time component and returns components.
//
// It tries both parseISODateCommon and parseISODateUncommon in succession.
//
// `components` is a [3]int of (year, month, day).
// We pass through `pos` to let the caller (ParseISODatetime) know if the full
// string contains a time, or just a date.
//
// Note: this returns simple ints, *not* time.Month instances.  Careful with comparison.
func parseISODate(dateString string) (components [3]int, pos int, err error) {
	components, pos, err = parseISODateCommon(dateString)
	if err != nil {
		components, pos, err = parseISODateUncommon(dateString)
		if err != nil {
			return components, pos, err
		}
	}
	return components, pos, nil
}

// ParseISODate parses an ISO-8601 date string with no time component and returns components.
func ParseISODate(dateString string) (time.Time, error) {
	components, pos, err := parseISODate(dateString)
	if err != nil {
		return time.Time{}, err
	}
	if pos < len(dateString) {
		// This final check needs to remain separate.
		// I.e. this logic is not followed in ParseISODatetime
		return time.Time{}, &ParseError{dateString, "string contains unknown iso components"}
	}
	return strictDate(components[0], time.Month(components[1]), components[2], 0, 0, 0, 0, time.Local)
}

// parseTimezone parses an ISO-8601 timezone string, from Z, ±HH:MM, ±HHMM, or ±HH.
// It allows Unicode minus-sign or minus-hyphen as the leading sign, in addition to plus-sign.
func parseTimezone(tzString string) (tz *time.Location, err error) {
	if tzString[0] == 'Z' {
		// var UTC *Location = &utcLoc
		return time.UTC, nil
	}

	length := len(tzString)
	if _, ok := map[int]bool{3: true, 5: true, 6: true}[length]; !ok {
		return time.Local, &ParseError{tzString, "time zone offset string must be 1, 3, 5 or 6 characters"}
	}

	// Except for Z, leading sign is required.
	var mult int
	if start := tzString[0]; start == '+' {
		mult = 1
	} else if start == '-' || int32(start) == '\u2212' {
		// Unicode minus-sign, different from ASCII hyphen.  Allow both.
		// ("hyphen" and "minus" are both mapped onto "hyphen-minus.")
		mult = -1
	} else {
		return tz, &ParseError{tzString, "unrecognized timezone sign"}
	}

	// Hour and minute
	hours, _ := strconv.Atoi(tzString[1:3])
	var minutes int
	if length != 3 {
		// We are down to ±HH:MM and ±HHMM
		if tzString[3] == ':' {
			minutes, _ = strconv.Atoi(tzString[4:])
		} else {
			minutes, _ = strconv.Atoi(tzString[3:])
		}
	}

	if (hours == 0) && (minutes == 0) {
		return time.UTC, nil
	}

	if hours < minHour || hours > maxHour || minutes < minMin || minutes > maxMin {
		return time.Local, &ParseError{tzString, "offset component out of valid range"}
	}

	// We need seconds east of UTC as float64.
	secondsEast := int(mult * 60 * (hours*60 + minutes))

	// We cannot explicitly name the time zone (or determine DST)
	// just based solely on its offset.  This seems to be the next best thing,
	// although it is not ideal because it returns a time.Location where the caller
	// cannot change `.name` (unexported field) from what is given here.
	return time.FixedZone("UTC", secondsEast), nil
}

// Note: an all-out-regex may work for ParseISOTime, such as:
// re := regexp.MustCompile(`(?P<hour>\d{2}):?(?P<minute>\d{2})?:?(?P<second>\d{2})?[\\.,]?(?P<frac>\d{1,9})?(?P<offset>Z|[+-]\d{2}:?\d{2}?)?`)
// However, this would yield "false positives" for times such as "12:", and Go does not support lookahead.
// The time complexity of the existing approach is good, so we stick with that.

// ParseISOTime parses an ISO-8601 time string with no date component.
// Examples: HH, HH:MM or HHMM, HH:MM:SS or HHMMSS, HH:MM:SS.ssssss.  (Plus an optional time zone portion.)
// `components` here represents hour, minute, second, nanosecond.
func ParseISOTime(timeString string) (components [4]int, tz *time.Location, err error) {
	tz = time.Local
	length := len(timeString)
	// `comp` represents the current index for `components` as we proceed through
	pos, comp := 0, -1

	if length < 2 {
		return components, tz, &ParseError{timeString, "length of time string must be >= 2"}
	}

	hasSep := length >= 3 && timeString[2] == timeSep

	// Support for fractional components other than seconds is part of the
	// ISO-8601 standard, but is not currently implemented in this parser.
	// From Wikipedia: "To denote '14 hours, 30 and one half minutes,' do not include a seconds figure.
	// 					Represent it as '14:30,5', '1430,5', '14:30.5', or '1430.5'."
	// These times will return a ParseError.

	for pos < length && comp < 4 {
		comp += 1

		if start := timeString[pos]; start == 'Z' || start == '+' || start == '-' {
			// Timezone "boundary" detected
			tz, err = parseTimezone(timeString[pos:])
			if err != nil {
				return components, tz, err
			}
			pos = length
			break
		}

		if comp < 3 {
			// Hour, minute, second
			components[comp], _ = strconv.Atoi(timeString[pos : pos+2])
			pos += 2
			if hasSep && pos < length && timeString[pos] == timeSep {
				pos += 1
			}
		}

		if comp == 3 {
			// Second fraction (optional)
			frac := fractionRegex.FindStringSubmatch(timeString[pos:])
			if frac == nil {
				continue
			}

			// There is formally no limit on the number of decimal places for the decimal fraction.
			// But Go's time package has nanosecond precision.
			// See also:
			// https://github.com/dateutil/dateutil/commit/9d2edc0e17cc16eaea49dbea379b85ba4f1e610e
			// We do not raise if caller tries to pass 10 or more digits; we simply chop off to 9.
			// For example, .3684000309 seconds becomes 368400030 nanoseconds
			//
			// Unlike Python, need to make sure we aren't slicing the fraction digits
			// outside the string capacity. There is probably a more idiomatic way to do this slicing.
			secondsFrac, _ := strconv.ParseFloat("0."+frac[1][:min(9, len(frac[1]))], 64)
			// We have float seconds, convert to nanoseconds (10 ** 9 seconds).
			// Note that there is no rounding done here, just truncation.
			// time.Nanosecond (and other constants) are available but not really needed.
			components[comp] = int(secondsFrac * 1e9)
			pos += len(frac[0])
		}
	}

	if pos < length {
		return components, tz, &ParseError{timeString, "unused components"}
	}

	if components[0] == 24 {
		for _, i := range components[1:] {
			// Standard supports 00:00 and 24:00 as representations of midnight
			// But this means no minutes may be attached with hour 24
			if i != 0 {
				return components, tz, &ParseError{timeString, "hour == 24 implies 0 for other time units"}
			}
		}
		// Otherwise, we don't need to set to 0.  This is the only time we want to take advantage of
		// go's time.Date rolling over (normalizing/overflowing) components.
		// time.Date(2014, 4, 10, 24, 0, 0, 0, time.Local) becomes 2014-04-11 00:00:00 on its own.

	}
	// Go does not really have the concept of a "naive" datetime with no timezone info.  All times are initialized with a time.Location arg.
	// - time.Local is, roughly, the zero value for time.Location; it is just `var localLoc Location; var Local *Location = &localLoc`
	// - time.UTC is `var utcLoc = Location{name: "UTC"}; var UTC *Location = &utcLoc`
	// - String() for the time.Location zero value will return time.UTC; see also `func (l *Location) get()`
	return components, tz, nil
}

// ParseISODatetime parses an ISO-8601 datetime (combined date and time string).
//
// It can also parse just a date in isolation, but if the user knows that input strings
// contain only dates with no time components, it will be faster to use ParseISODate.
//
// This function is like Go's `time.Parse(layout, value string) (Time, error)`,
// except without the `layout` parameter.  It attempts to otherwise follow the call syntax of
// `time.Parse` closely; if parse error is not nil, the returned Time will be
// 0001-01-01 00:00:00 +0000 UTC, i.e. time.Time{}.
//
// If no timezone/offset is detected (either with 'Z' or an hh[:mm] offset), the result will
// have loc time.Local.
func ParseISODatetime(datetime string) (time.Time, error) {
	// Date first
	// We get position to know where the date stops
	dateParts, pos, err := parseISODate(datetime)
	if err != nil {
		// Stop here, and keep just the dateString in the ParseError message.
		return time.Time{}, err
	}

	var (
		hour, minute, second, nsec int
		tz                         *time.Location // time.UTC zero value
	)

	// If len(datetime) > pos, it appears we have a time portion
	// If len(datetime) < pos, something's gone very wrong with parseISODate
	// If they're equal, we just have a (seemingly valid) date

	if len(datetime) > pos {
		// Make sure the sep between date and time (strictly just "T") is a non-numeric ASCII character.
		// This means: 0 thru 127 except 48 thru 57 in decimal.
		if sep := datetime[pos]; (sep >= 0 && sep < 48) || (sep > 47 && sep <= 127) {
			var (
				timeParts [4]int
				err       error
			)
			timeParts, tz, err = ParseISOTime(datetime[pos+1:])
			if err != nil {
				tz = time.Local
				// Only erring out because we were signaled that a time portion should be there.
				// Note that passing nil for tz will cause time.Date to panic.
				return time.Date(1, 1, 1, 0, 0, 0, 0, tz), err
			}
			hour, minute, second, nsec = timeParts[0], timeParts[1], timeParts[2], timeParts[3]
		} else {
			tz = time.Local
			return time.Date(1, 1, 1, 0, 0, 0, 0, tz), &ParseError{datetime, "date/time separator must be a non-numeric ASCII character"}
		}

	} else if len(datetime) < pos {
		// This really shouldn't be reached, but represents a case where the
		// position cursor moved past the entire string in parsing just the date.
		return time.Time{}, &ParseError{Datetime: datetime}
	}
	// We need to be very careful about passing the zero value for time.Location here
	res, err := strictDate(dateParts[0], time.Month(dateParts[1]), dateParts[2], hour, minute, second, nsec, tz)
	return res, err
}

// Note that this differs from time.Time.In or time.Time.UTC in that it does not change the
// underlying timestamp components; it merely returns a new time.Time with the same
// year, month, ..., nsec components, but a different loc.
func SetLoc(t time.Time, loc *time.Location) time.Time {
	year, month, day, hour, minute, second, nsec := t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()
	return time.Date(year, month, day, hour, minute, second, nsec, loc)
}
