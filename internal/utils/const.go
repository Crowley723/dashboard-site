package utils

import (
	"regexp"
	"time"
)

const (
	// Hour is an int based representation of the time unit.
	Hour = time.Minute * 60

	// Day is an int based representation of the time unit.
	Day = Hour * 24

	// Week is an int based representation of the time unit.
	Week = Day * 7

	// Year is an int based representation of the time unit.
	Year = Day * 365

	// Month is an int based representation of the time unit.
	Month = Year / 12
)

var (
	// StandardTimeLayouts is the set of standard time layouts used with ParseTimeString.
	StandardTimeLayouts = []string{
		"Jan 2 15:04:05 2006",
		time.DateTime,
		time.RFC3339,
		time.RFC1123Z,
		time.RubyDate,
		time.ANSIC,
		time.DateOnly,
	}

	standardDurationUnits = []string{"ns", "us", "µs", "μs", "ms", "s", "m", "h"}

	reOnlyNumeric      = regexp.MustCompile(`^\d+$`)
	reDurationStandard = regexp.MustCompile(`(?P<Duration>[1-9]\d*?)(?P<Unit>[^\d\s]+)`)
	reNumeric          = regexp.MustCompile(`\d+`)
)

// Duration unit types.
const (
	DurationUnitDays   = "d"
	DurationUnitWeeks  = "w"
	DurationUnitMonths = "M"
	DurationUnitYears  = "y"
)

// Number of hours in particular measurements of time.
const (
	HoursInDay   = 24
	HoursInWeek  = HoursInDay * 7
	HoursInMonth = HoursInDay * 30
	HoursInYear  = HoursInDay * 365
)
