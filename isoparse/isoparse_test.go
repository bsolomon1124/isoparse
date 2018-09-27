package isoparse

// See also:
// https://github.com/dateutil/dateutil/blob/master/dateutil/test/test_isoparser.py
//
// Note that these tests are not exactly mirrored after the tests in the module above,
// though they do largely follow along with them in structure.
//
// Variables are presented first in one batch; tests second.
//
// Important note regarding equality testing from GoDoc: time
// 		Note that the Go == operator compares not just the time instant but also the Location
// 		and the monotonic clock reading. See the documentation for the Time type for a
// 		discussion of equality testing for Time values.

import (
	"reflect"
	"testing"
	"time"
)

var isoToGregorian = map[[3]int]time.Time{
	{1950, 1, 1}:  time.Date(1950, 1, 2, 0, 0, 0, 0, time.Local),
	{1950, 1, 3}:  time.Date(1950, 1, 4, 0, 0, 0, 0, time.Local),
	{1950, 1, 5}:  time.Date(1950, 1, 6, 0, 0, 0, 0, time.Local),
	{1950, 1, 6}:  time.Date(1950, 1, 7, 0, 0, 0, 0, time.Local),
	{1950, 20, 1}: time.Date(1950, 5, 15, 0, 0, 0, 0, time.Local),
	{1950, 20, 3}: time.Date(1950, 5, 17, 0, 0, 0, 0, time.Local),
	{1950, 20, 5}: time.Date(1950, 5, 19, 0, 0, 0, 0, time.Local),
	{1950, 20, 6}: time.Date(1950, 5, 20, 0, 0, 0, 0, time.Local),
	{1950, 27, 1}: time.Date(1950, 7, 3, 0, 0, 0, 0, time.Local),
	{1950, 27, 3}: time.Date(1950, 7, 5, 0, 0, 0, 0, time.Local),
	{1950, 27, 5}: time.Date(1950, 7, 7, 0, 0, 0, 0, time.Local),
	{1950, 27, 6}: time.Date(1950, 7, 8, 0, 0, 0, 0, time.Local),
	{1950, 53, 1}: time.Date(1951, 1, 1, 0, 0, 0, 0, time.Local),
	{1950, 53, 3}: time.Date(1951, 1, 3, 0, 0, 0, 0, time.Local),
	{1950, 53, 5}: time.Date(1951, 1, 5, 0, 0, 0, 0, time.Local),
	{1950, 53, 6}: time.Date(1951, 1, 6, 0, 0, 0, 0, time.Local),
	{1984, 1, 1}:  time.Date(1984, 1, 2, 0, 0, 0, 0, time.Local),
	{1984, 1, 3}:  time.Date(1984, 1, 4, 0, 0, 0, 0, time.Local),
	{1984, 1, 5}:  time.Date(1984, 1, 6, 0, 0, 0, 0, time.Local),
	{1984, 1, 6}:  time.Date(1984, 1, 07, 0, 0, 0, 0, time.Local),
	{1984, 20, 1}: time.Date(1984, 5, 14, 0, 0, 0, 0, time.Local),
	{1984, 20, 3}: time.Date(1984, 5, 16, 0, 0, 0, 0, time.Local),
	{1984, 20, 5}: time.Date(1984, 5, 18, 0, 0, 0, 0, time.Local),
	{1984, 20, 6}: time.Date(1984, 5, 19, 0, 0, 0, 0, time.Local),
	{1984, 27, 1}: time.Date(1984, 7, 2, 0, 0, 0, 0, time.Local),
	{1984, 27, 3}: time.Date(1984, 7, 4, 0, 0, 0, 0, time.Local),
	{1984, 27, 5}: time.Date(1984, 7, 6, 0, 0, 0, 0, time.Local),
	{1984, 27, 6}: time.Date(1984, 7, 7, 0, 0, 0, 0, time.Local),
	{1984, 53, 1}: time.Date(1984, 12, 31, 0, 0, 0, 0, time.Local),
	{1984, 53, 3}: time.Date(1985, 1, 2, 0, 0, 0, 0, time.Local),
	{1984, 53, 5}: time.Date(1985, 1, 4, 0, 0, 0, 0, time.Local),
	{1984, 53, 6}: time.Date(1985, 1, 5, 0, 0, 0, 0, time.Local),
	{2002, 1, 1}:  time.Date(2001, 12, 31, 0, 0, 0, 0, time.Local),
	{2002, 1, 3}:  time.Date(2002, 1, 2, 0, 0, 0, 0, time.Local),
	{2002, 1, 5}:  time.Date(2002, 1, 4, 0, 0, 0, 0, time.Local),
	{2002, 1, 6}:  time.Date(2002, 1, 5, 0, 0, 0, 0, time.Local),
	{2002, 20, 1}: time.Date(2002, 5, 13, 0, 0, 0, 0, time.Local),
	{2002, 20, 3}: time.Date(2002, 5, 15, 0, 0, 0, 0, time.Local),
	{2002, 20, 5}: time.Date(2002, 5, 17, 0, 0, 0, 0, time.Local),
	{2002, 20, 6}: time.Date(2002, 5, 18, 0, 0, 0, 0, time.Local),
	{2002, 27, 1}: time.Date(2002, 7, 1, 0, 0, 0, 0, time.Local),
	{2002, 27, 3}: time.Date(2002, 7, 3, 0, 0, 0, 0, time.Local),
	{2002, 27, 5}: time.Date(2002, 7, 5, 0, 0, 0, 0, time.Local),
	{2002, 27, 6}: time.Date(2002, 7, 6, 0, 0, 0, 0, time.Local),
	{2002, 53, 1}: time.Date(2002, 12, 30, 0, 0, 0, 0, time.Local),
	{2002, 53, 3}: time.Date(2003, 1, 1, 0, 0, 0, 0, time.Local),
	{2002, 53, 5}: time.Date(2003, 1, 3, 0, 0, 0, 0, time.Local),
	{2002, 53, 6}: time.Date(2003, 1, 4, 0, 0, 0, 0, time.Local),
}

var trueDaysBeforeYear = map[int]int{
	5:   1461,   // Number of days *in* 4 years
	101: 36524,  // Number of days *in* 100 years
	401: 146097, // Number of days *in* 400 years
}

var leapYears = []int{1804, 1856, 1892, 1952, 1984, 2008, 2012, 2068, 2096}
var nonLeapYears = []int{1803, 1855, 1891, 1953, 1985, 2009, 2011, 2067, 2097}

var isoMap = map[time.Time][3]int{
	time.Date(2018, 9, 22, 10, 37, 17, 567365, time.Local): {2018, 38, 6},
	time.Date(1, 1, 1, 1, 1, 1, 1, time.Local):             {1, 1, 1},
	time.Date(1912, 4, 12, 0, 0, 0, 0, time.Local):         {1912, 15, 5},
}

var commonDates = map[string]time.Time{
	// Common valid dates

	// YYYY
	// Day defaults to 1
	"1990": time.Date(1990, time.January, 1, 0, 0, 0, 0, time.Local),
	"1999": time.Date(1999, time.January, 1, 0, 0, 0, 0, time.Local),
	"2018": time.Date(2018, time.January, 1, 0, 0, 0, 0, time.Local),

	// YYYY-MM
	// Day defaults to 1
	"1953-01": time.Date(1953, time.January, 1, 0, 0, 0, 0, time.Local),
	"1953-12": time.Date(1953, time.December, 1, 0, 0, 0, 0, time.Local),
	"2010-05": time.Date(2010, time.May, 1, 0, 0, 0, 0, time.Local),

	// YYYY-MM-DD
	"1990-01-01": time.Date(1990, time.January, 1, 0, 0, 0, 0, time.Local),
	"1999-12-13": time.Date(1999, time.December, 13, 0, 0, 0, 0, time.Local),
	"1941-12-07": time.Date(1941, time.December, 7, 0, 0, 0, 0, time.Local),
	"2050-03-30": time.Date(2050, time.March, 30, 0, 0, 0, 0, time.Local),

	// YYYYMMDD
	"19900105": time.Date(1990, time.January, 5, 0, 0, 0, 0, time.Local),
	"19991212": time.Date(1999, time.December, 12, 0, 0, 0, 0, time.Local),
	"19411207": time.Date(1941, time.December, 7, 0, 0, 0, 0, time.Local),
	"20500330": time.Date(2050, time.March, 30, 0, 0, 0, 0, time.Local),
	"18641010": time.Date(1864, time.October, 10, 0, 0, 0, 0, time.Local),
	"19641010": time.Date(1964, time.October, 10, 0, 0, 0, 0, time.Local),
	"20010101": time.Date(2001, time.January, 1, 0, 0, 0, 0, time.Local),
}

var uncommonDates = map[string]time.Time{
	// Uncommon, but valid.
	"2009-W01-1": time.Date(2008, time.December, 29, 0, 0, 0, 0, time.Local),
	"2009-W53-7": time.Date(2010, time.January, 3, 0, 0, 0, 0, time.Local),
	"1981-095":   time.Date(1981, time.April, 5, 0, 0, 0, 0, time.Local),
	"1981-009":   time.Date(1981, time.January, 9, 0, 0, 0, 0, time.Local),
	"2001-300":   time.Date(2001, time.October, 27, 0, 0, 0, 0, time.Local),
}

var timesWithComponents = map[string][4]int{
	"16:22:12+00:00": {16, 22, 12, 0},
	"16:22:12Z":      {16, 22, 12, 0},
	"162212Z":        {16, 22, 12, 0},
	"134730":         {13, 47, 30, 0},
	"13:47:30":       {13, 47, 30, 0},
	"09:30Z":         {9, 30, 0, 0},
	"0930Z":          {9, 30, 0, 0},
	"14:45:15Z":      {14, 45, 15, 0},
	"144515Z":        {14, 45, 15, 0},
}

var tzStrings = map[string]*time.Location{
	"+0000":  time.UTC,
	"+00:00": time.UTC,
	"-0002":  time.FixedZone("UTC", -120),
	"-00:02": time.FixedZone("UTC", -120),
	"+0002":  time.FixedZone("UTC", 120),
	"+00:02": time.FixedZone("UTC", 120),
	"-0015":  time.FixedZone("UTC", -900),
	"-00:15": time.FixedZone("UTC", -900),
	"+0015":  time.FixedZone("UTC", 900),
	"+00:15": time.FixedZone("UTC", 900),
	"-0500":  time.FixedZone("UTC", -18000),
	"-05:00": time.FixedZone("UTC", -18000),
	"+0500":  time.FixedZone("UTC", 18000),
	"+05:00": time.FixedZone("UTC", 18000),
	"-0502":  time.FixedZone("UTC", -18120),
	"-05:02": time.FixedZone("UTC", -18120),
	"+0502":  time.FixedZone("UTC", 18120),
	"+05:02": time.FixedZone("UTC", 18120),
	"-0515":  time.FixedZone("UTC", -18900),
	"-05:15": time.FixedZone("UTC", -18900),
	"+0515":  time.FixedZone("UTC", 18900),
	"+05:15": time.FixedZone("UTC", 18900),
	"-2300":  time.FixedZone("UTC", -82800),
	"-23:00": time.FixedZone("UTC", -82800),
	"+2300":  time.FixedZone("UTC", 82800),
	"+23:00": time.FixedZone("UTC", 82800),
	"-2302":  time.FixedZone("UTC", -82920),
	"-23:02": time.FixedZone("UTC", -82920),
	"+2302":  time.FixedZone("UTC", 82920),
	"+23:02": time.FixedZone("UTC", 82920),
	"-2315":  time.FixedZone("UTC", -83700),
	"-23:15": time.FixedZone("UTC", -83700),
	"+2315":  time.FixedZone("UTC", 83700),
	"+23:15": time.FixedZone("UTC", 83700),
}

// Invalid ISO strings per 2004 standard
// "By disallowing dates of the form YYYYMM, the standard avoids confusion with the
// truncated representation YYMMDD (still often used)."
// Note that dateutil does allow these, it seems.
var invalidYYYYMM = []string{
	"195301",
	"195312",
	"201005",
	"177607",
	"149201",
}

var invalidDates = []string{
	"243",         // ISO string too short
	"2014-0423",   // Inconsistent date separators
	"201404-23",   // Inconsistent date separators
	"2014日03月14",  // Not ASCII
	"2013-02-29",  // Invalid day
	"2014/12/03",  // Wrong separators
	"2014-04-19T", // Unknown components
}

var invalidDatetimes = []string{
	"201",                          // ISO string too short
	"2012-0425",                    // Inconsistent date separators
	"201204-25",                    // Inconsistent date separators
	"20120425T0120:00",             // Inconsistent time separators
	"20120425T012500-334",          // Wrong microsecond separator
	"2001-1",                       // YYYY-M not valid
	"2012-04-9",                    // YYYY-MM-D not valid
	"201204",                       // YYYYMM not valid
	"20120411T03:30+",              // Time zone too short
	"20120411T03:30+1234567",       // Time zone too long
	"20120411T03:30-25:40",         // Time zone invalid
	"2012-1a",                      // Invalid month
	"20120411T03:30+00:60",         // Time zone invalid minutes
	"20120411T03:30+00:61",         // Time zone invalid minutes
	"20120411T033030.123456012:00", // No sign in time zone
	"2012-W00",                     // Invalid ISO week
	"2012-W55",                     // Invalid ISO week
	"2012-W01-0",                   // Invalid ISO week day
	"2012-W01-8",                   // Invalid ISO week day
	"2013-000",                     // Invalid ordinal day
	"2013-366",                     // Invalid ordinal day
	"2013366",                      // Invalid ordinal day
	"2014-03-12Т12:30:14",          // Cyrillic T
	"2014-04-21T24:00:01",          // Invalid use of 24 for midnight
	"2014_W01-1",                   // Invalid separator
	"1985-102☐10:15Z",              // Invalid separator
	"2014W01-1",                    // Inconsistent use of dashes
	"2014-W011",                    // Inconsistent use of dashes
}

// Note that we don't include stuff like "25" or "14:60" here (invalid components).
// This will be caught in ParseISODatetime, but not in ParseISOTime, because it just reurns components.
var invalidTimes = []string{
	"3",                    //  ISO string too short
	"14時30分15秒",            //  Not ASCII
	"14_30_15",             //  Invalid separators
	"1430:15",              //  Inconsistent separator use
	"14:30:15.34468305:00", //  No sign in time zone
	"14:30:15+",            //  Time zone too short
	"14:30:15+1234567",     //  Time zone invalid
	"14:59:30_344583",      //  Invalid microsecond separator
	"24:01",                //  24 used for non-midnight time
	"24:00:01",             //  24 used for non-midnight time
	"24:00:00.001",         //  24 used for non-midnight time
	"24:00:00.000001",      // 24 used for non-midnight time
}

var invalidTzStrings = []string{
	"00:00",   // No sign
	"05:00",   // No sign
	"_00:00",  // Invalid sign
	"00:0000", // # String too long
}

var zeroTzs = []string{
	"-00:00",
	"+00:00",
	"+00",
	"-00",
	"+0000",
	"-0000",
}

var midnightISODatetimes = map[string]time.Time{
	"2014-04-11T00":              time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-10T24":              time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-11T00:00":           time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-10T24:00":           time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-11T00:00:00":        time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-10T24:00:00":        time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-11T00:00:00.000":    time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-10T24:00:00.000":    time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-11T00:00:00.000000": time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
	"2014-04-10T24:00:00.000000": time.Date(2014, 4, 11, 0, 0, 0, 0, time.Local),
}

var differentSepISODatetimes = map[string]time.Time{
	"2014-01-01 14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
	"2014-01-01a14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
	"2014-01-01T14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
	"2014-01-01_14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
	"2014-01-01-14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),

	"20140101 14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
	"20140101a14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
	"20140101T14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
	"20140101_14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
	"20140101-14:33:09": time.Date(2014, 1, 1, 14, 33, 9, 0, time.Local),
}

// Bad parameters for `strictDate()`
var invalidParams = [][]int{
	{10001, 7, 4, 30, 30, 30, 100},     // Bad year
	{-1, 7, 4, 30, 30, 30, 100},        // Bad year
	{2000, 0, 4, 30, 30, 30, 100},      // Bad month
	{2000, 14, 4, 30, 30, 30, 100},     // Bad month
	{2000, 7, 32, 30, 30, 30, 100},     // Bad day (for given month)
	{2000, 2, 30, 30, 30, 30, 100},     // Bad day (for given month)
	{2000, 7, 4, -1, 30, 30, 100},      // Bad hour
	{2000, 7, 4, 25, 30, 30, 100},      // Bad hour (note: 24 is ok, contingent)
	{2000, 7, 4, 60, 30, 30, 100},      // Bad minute
	{2000, 7, 4, 61, 30, 30, 100},      // Bad minute
	{2000, 7, 4, 30, 61, 30, 100},      // Bad second
	{2000, 7, 4, 30, 60, 30, 100},      // Bad second
	{2000, 7, 4, 30, 30, 30, -1},       // Bad nsec
	{2000, 7, 4, 30, 30, 30, int(1e9)}, // Bad nsec
}

// Datetime strings with over 9 digits of second-fraction precision.
var extraPrecision = map[string]time.Time{
	"2018-07-03T14:07:00.123456000001": time.Date(2018, time.Month(7), 3, 14, 7, 0, 123456000, time.Local),
	"2018-07-03T14:07:00.123456999999": time.Date(2018, time.Month(7), 3, 14, 7, 0, 123456999, time.Local),
}

var unequalGregorianISO = map[string]time.Time{
	"2016-W13-7": time.Date(2016, time.Month(4), 3, 0, 0, 0, 0, time.Local),
	"2016W137":   time.Date(2016, time.Month(4), 3, 0, 0, 0, 0, time.Local),
	"2004-W53-7": time.Date(2005, time.Month(1), 2, 0, 0, 0, 0, time.Local), // ISO year != Cal year
	"2004W537":   time.Date(2005, time.Month(1), 2, 0, 0, 0, 0, time.Local),
	"2009-W01-2": time.Date(2008, time.Month(12), 30, 0, 0, 0, 0, time.Local), // ISO year < Cal year
	"2009W012":   time.Date(2008, time.Month(12), 30, 0, 0, 0, 0, time.Local),
	"2009-W53-6": time.Date(2010, time.Month(1), 2, 0, 0, 0, 0, time.Local), // ISO year > Cal year
	"2009W536":   time.Date(2010, time.Month(1), 2, 0, 0, 0, 0, time.Local),
}

type connector struct {
	// Used in example to map between ISO format and actual datetime string + datetime object.
	t time.Time
	f string
}

// A nearly exhaustive list of anything and everything from the ISO 8601:2004 standard.
// (With the exception of bare times.)
// Probably duplicates some other test vars we have here, but is a sanity check at the very least.
var allFormats = map[string]connector{
	// Dates
	"19850412":   {t: time.Date(1985, time.Month(4), 12, 0, 0, 0, 0, time.Local), f: "YYYYMMDD"},
	"1985-04-12": {t: time.Date(1985, time.Month(4), 12, 0, 0, 0, 0, time.Local), f: "YYYY-MM-DD"},
	"1985-04":    {t: time.Date(1985, time.Month(4), 1, 0, 0, 0, 0, time.Local), f: "YYYY-MM"},
	"1985":       {t: time.Date(1985, time.Month(1), 1, 0, 0, 0, 0, time.Local), f: "YYYY"},
	// "19"(YY) is disallowed by this package.
	"1985102":    {t: time.Date(1985, time.Month(4), 12, 0, 0, 0, 0, time.Local), f: "YYYYDDD"},
	"1985-102":   {t: time.Date(1985, time.Month(4), 12, 0, 0, 0, 0, time.Local), f: "YYYY-DDD"},
	"1985W155":   {t: time.Date(1985, time.Month(4), 12, 0, 0, 0, 0, time.Local), f: "YYYYWwwD"},
	"1985-W15-5": {t: time.Date(1985, time.Month(4), 12, 0, 0, 0, 0, time.Local), f: "YYYY-Www-D"},
	"1985W15":    {t: time.Date(1985, time.Month(4), 8, 0, 0, 0, 0, time.Local), f: "YYYYWww"},
	"1985-W15":   {t: time.Date(1985, time.Month(4), 8, 0, 0, 0, 0, time.Local), f: "YYYY-Www"},

	// Datetimes
	// There are not exhausitve; they are basically multiplicative
	// (combinatorial product) between valid dates and valid times.
	"19850412T101530":           {t: time.Date(1985, time.Month(4), 12, 10, 15, 30, 0, time.Local), f: "YYYYMMDDTHHMMSS"},
	"19850412T101530Z":          {t: time.Date(1985, time.Month(4), 12, 10, 15, 30, 0, time.UTC), f: "YYYYMMDDTHHMMSSZ"},
	"19850412T101530+0400":      {t: time.Date(1985, time.Month(4), 12, 10, 15, 30, 0, time.FixedZone("UTC", int(4*60*60))), f: "YYYYMMDDTHHMMSS±hhmm"},
	"19850412T101530+04":        {t: time.Date(1985, time.Month(4), 12, 10, 15, 30, 0, time.FixedZone("UTC", int(4*60*60))), f: "YYYYMMDDTHHMMSS±hh"},
	"1985-04-12T10:15:30":       {t: time.Date(1985, time.Month(4), 12, 10, 15, 30, 0, time.Local), f: "YYYYMMDDTHH:MM:SS"},
	"1985-04-12T10:15:30Z":      {t: time.Date(1985, time.Month(4), 12, 10, 15, 30, 0, time.UTC), f: "YYYYMMDDTHH:MM:SSZ"},
	"1985-04-12T10:15:30+04:00": {t: time.Date(1985, time.Month(4), 12, 10, 15, 30, 0, time.FixedZone("UTC", int(4*60*60))), f: "YYYYMMDDTHH:MM:SS±hhmm"},
	"1985-04-12T10:15:30+04":    {t: time.Date(1985, time.Month(4), 12, 10, 15, 30, 0, time.FixedZone("UTC", int(4*60*60))), f: "YYYYMMDDTHH:MM:SS±hh"},
	"19850412T1015":             {t: time.Date(1985, time.Month(4), 12, 10, 15, 0, 0, time.Local), f: "YYYY-MM-DDTHHMM"},
	"1985-04-12T10:15":          {t: time.Date(1985, time.Month(4), 12, 10, 15, 0, 0, time.Local), f: "YYYY-MM-DDTHHMM"},
	"1985102T1015Z":             {t: time.Date(1985, time.Month(4), 12, 10, 15, 0, 0, time.UTC), f: "YYYYDDDDTHHMMZ"},
	"1985-102T10:15Z":           {t: time.Date(1985, time.Month(4), 12, 10, 15, 0, 0, time.UTC), f: "YYYY-DDD:MMZ"},
	"1985W155T1015+0400":        {t: time.Date(1985, time.Month(4), 12, 10, 15, 0, 0, time.FixedZone("UTC", int(4*60*60))), f: "YYYY-WwwTHH:MMZ"},
	"1985-W15-5T10:15+04":       {t: time.Date(1985, time.Month(4), 12, 10, 15, 0, 0, time.FixedZone("UTC", int(4*60*60))), f: "YYYY-Www-DTHH:MM±hh"},
}

// //////////////////////////////////////////////////
// Make sure helper functions check out.
// //////////////////////////////////////////////////

func TestStrictDate(t *testing.T) {
	for _, c := range invalidParams {
		year, month, day, hour, minute, second, nsec := c[0], c[1], c[2], c[3], c[4], c[5], c[6]
		tm, err := strictDate(year, time.Month(month), day, hour, minute, second, nsec, time.Local)
		if err == nil {
			t.Errorf(`strictDate(%v) -> %v (for invalid unit) returned nil error`, c, tm)
		}
	}
}

func TestIsLeapYear(t *testing.T) {
	for _, year := range leapYears {
		if !isLeapYear(year) {
			t.Errorf(`isLeapYear(%d) returned false for valid leap year`, year)
		}
	}
	for _, year := range nonLeapYears {
		if isLeapYear(year) {
			t.Errorf(`isLeapYear(%d) returned true for non-leap year`, year)
		}
	}
}

func TestDaysBeforeYear(t *testing.T) {
	for year, trueDays := range trueDaysBeforeYear {
		if days := daysBeforeYear(year); days != trueDays {
			t.Errorf(`daysBeforeYear(%d) -> %v (should be %d)`, year, days, trueDays)
		}
	}
}

func TestCalcWeekdate(t *testing.T) {
	for arr, trueDate := range isoToGregorian {
		if tm, err := calcWeekdate(arr[0], arr[1], arr[2]); err != nil {
			t.Errorf(`calcWeekdate(%d, %d, %d) -> -> non-nill error (%v) for valid input`, arr[0], arr[1], arr[2], err)
		} else if !tm.Equal(trueDate) {
			t.Errorf(`calcWeekdate(%d, %d, %d) -> %v (should be %v)`, arr[0], arr[1], arr[2], tm, trueDate)
		}
	}
}

func TestISOCalendar(t *testing.T) {
	for dt, arr := range isoMap {
		if isoCalendar(dt) != arr {
			t.Errorf(`isoCalendar(%v) error: produced %v, should be %v`, dt, isoCalendar(dt), arr)
		}
	}
}

// //////////////////////////////////////////////////
// Tests of the core parsing functions:
// - parseISODateCommon
// - parseISODateUncommon
// - ParseISOTime
// - parseTimezone
// - ParseISODatetime (exported)
// - ParseISODate     (exported)
//
// Note that the unexported functions may "ignore" some cases purposefully: these are cases that
// they, by design, should not catch independently, but only deal with in coordination with other
// functions when wrapped together in `ParseISODatetime()` and `ParseISODate()`
// //////////////////////////////////////////////////

func TestParseISODateCommon(t *testing.T) {
	for dateString, trueDate := range commonDates {
		components, _, err := parseISODateCommon(dateString)
		if err != nil {
			t.Errorf(`parseISODateCommon(%q) -> non-nill error (%v) for valid time string`, dateString, err)
		} else if (components[0] != trueDate.Year()) || (components[1] != int(trueDate.Month())) || (components[2] != trueDate.Day()) {
			t.Errorf(`parseISODateCommon(%q) -> %v (should be %v)`, dateString, components, trueDate)
		}
	}
}

func TestParseISODateUncommon(t *testing.T) {
	for dateString, trueDate := range uncommonDates {
		components, _, err := parseISODateUncommon(dateString)
		if err != nil {
			t.Errorf(`parseISODateUncommon(%q) -> non-ok response produced %v`, dateString, components)
		} else if (components[0] != trueDate.Year()) || (components[1] != int(trueDate.Month())) || (components[2] != trueDate.Day()) {
			t.Errorf(`parseISODateUncommon(%q) -> %v (should be %v)`, dateString, components, trueDate)
		}
	}
}

func TestParseISODate(t *testing.T) {
	for dateString, trueDate := range commonDates {
		if components, _, err := parseISODate(dateString); err != nil {
			t.Errorf(`non-nil error`)
		} else if (components[0] != trueDate.Year()) || (components[1] != int(trueDate.Month())) || (components[2] != trueDate.Day()) {
			t.Errorf(`parseISODate(%q) -> %v (should be %v)`, dateString, components, trueDate)
		}
	}
}

func TestParseISOTime(t *testing.T) {
	for timeString, trueComp := range timesWithComponents {
		components, _, err := ParseISOTime(timeString)
		if err != nil {
			t.Errorf(`ParseISOTime(%q) -> non-nil error (%v) for valid time string`, timeString, err)
		} else {
			for i := 0; i <= 3; i++ {
				if components[i] != trueComp[i] {
					t.Errorf(`ParseISOTime(%q) -> %v (should be %v)`, timeString, components, trueComp)
				}
			}
		}
	}
}

// See dateutil.test.test_isoparser.test_parse_tzstr
func TestParseTimezone(t *testing.T) {
	for tzString, trueTZ := range tzStrings {
		if tz, err := parseTimezone(tzString); err != nil {
			t.Errorf(`parseTimezone(%q) -> non-nil error (%v) for valid tzString`, tzString, err)
		} else if !reflect.DeepEqual(tz, trueTZ) {
			// Google's go-cmp seems to be a better choice here, but reflect should
			// suffice in this specific case.
			t.Errorf(`parseTimezone(%q) -> %v (should be %v)`, tzString, tz, trueTZ)
		}
	}
}

func TestParseISODatetime(t *testing.T) {
	for datetime, c := range allFormats {
		if dt, err := ParseISODatetime(datetime); err != nil {
			t.Errorf(`ParseISODatetime(%q) -> non-nil error (%v) for valid datetime string`, datetime, err)
		} else if !dt.Equal(c.t) {
			t.Errorf(`ParseISODatetime(%q) -> %v (should be %v)`, datetime, dt, c.t)
		}
	}
}

// //////////////////////////////////////////////////
// Confirm that invalid inputs raise non-nil errors.
// //////////////////////////////////////////////////

// Make sure we're not allowing what is a disallowed ambiguous pattern (confusable with YYMMDD).
func TestInvalidYYYYMM(t *testing.T) {
	for _, dateString := range invalidYYYYMM {
		if components, _, err := parseISODate(dateString); err == nil {
			t.Errorf(`parseISODate(%q) -> %v (for invalid YYYYMM) returned nil error`, dateString, components)
		}
	}
}

func TestInvalidDate(t *testing.T) {
	for _, dateString := range invalidDates {
		if dt, err := ParseISODate(dateString); err == nil {
			t.Errorf(`ParseISODate(%q) -> %v returned nil error (invalid dateString should error)`, dateString, dt)
		}
	}

}

func TestInvalidTime(t *testing.T) {
	for _, timeString := range invalidTimes {
		if _, _, err := ParseISOTime(timeString); err == nil {
			t.Errorf(`ParseISOTime(%q) returned nil error (invalid timeString should error)`, timeString)
		}
	}
}

func TestAssertInvalidTz(t *testing.T) {
	for _, tzString := range invalidTzStrings {
		if _, err := parseTimezone(tzString); err == nil {
			t.Errorf(`parseTimezone(%q) returned nil error (invalid tzString should error)`, tzString)
		}
	}
}

func TestInvalidDatetime(t *testing.T) {
	for _, datetime := range invalidDatetimes {
		if _, err := ParseISODatetime(datetime); err == nil {
			t.Errorf(`ParseISODatetime(%q) returned nil error (invalid datetime should error)`, datetime)
		}
	}
}

// //////////////////////////////////////////////////
// Stress-test a number of other edge cases.
// //////////////////////////////////////////////////

func TestUnequalISOWeekDay(t *testing.T) {
	for datetime, trueDate := range unequalGregorianISO {
		if tm, err := ParseISODatetime(datetime); err != nil {
			t.Errorf(`ParseISODatetime(%q) -> non-nil error (%v) for valid datetime`, datetime, err)
		} else if !tm.Equal(trueDate) {
			t.Errorf(`ParseISODatetime(%q) -> %v should (should be %v)`, datetime, tm, trueDate)
		}
	}
}

func TestTzZeroUTC(t *testing.T) {
	for _, tzString := range zeroTzs {
		if _, err := parseTimezone(tzString); err != nil {
			t.Errorf(`parseTimezone(%q) -> non-nil error (%v) for valid tzString`, tzString, err)
		} else if tz, _ := parseTimezone(tzString); tz != time.UTC {
			t.Errorf(`parseTimezone(%v) error: produced %v, should be time.UTC`, tzString, tz)
		}
	}
}

func TestMidnight(t *testing.T) {
	for datetime, trueDate := range midnightISODatetimes {
		tm, err := ParseISODatetime(datetime)
		if err != nil {
			t.Errorf(`ParseISODatetime(%q) -> non-nil error (%v) for valid datetime`, datetime, err)
		} else if !tm.Equal(trueDate) {
			t.Errorf(`ParseISODatetime(%q) -> %v (should be %v)`, datetime, tm, trueDate)
		}
	}
}

func TestSep(t *testing.T) {
	for datetime, trueDate := range differentSepISODatetimes {
		tm, err := ParseISODatetime(datetime)
		if err != nil {
			t.Errorf(`ParseISODatetime(%q) -> non-nil error (%v) for valid datetime`, datetime, err)
		} else if !tm.Equal(trueDate) {
			t.Errorf(`ParseISODatetime(%q) -> %v (should be %v)`, datetime, tm, trueDate)
		}

	}
}

// Make sure we truncate anything beyond 9 digits of precision for fraction component of time.
func TestTruncateNsec(t *testing.T) {
	for datetime, trueDate := range extraPrecision {
		if tm, err := ParseISODatetime(datetime); err != nil {
			t.Errorf(`ParseISODatetime(%q) -> non-nil error (%v) for valid datetime`, datetime, err)
		} else if !tm.Equal(trueDate) {
			t.Errorf(`ParseISODatetime(%q) -> %v should truncate seconds fraction (should be %v)`, datetime, tm, trueDate)
		}
	}
}
