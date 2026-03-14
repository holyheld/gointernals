package typeutil

import "time"

// Timestamp converts time.Time to unix timestamp. if time.Time is a zero time,
// returned value is 0 (zeroish return).
func Timestamp(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	return t.Unix()
}

// EndOfDay returns the last instance of the same day in current location.
func EndOfDay(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		23,
		59,
		59,
		999,
		t.Location(),
	)
}

// FirstOfMonth returns the first instant of the month in
// current location.
func FirstOfMonth(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		1,
		0,
		0,
		0,
		0,
		t.Location(),
	)
}

// LastOfMonth returns the last instant of month in current location.
func LastOfMonth(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		DaysIn(t.Month(), t.Year()),
		23,
		59,
		59,
		int((time.Second - time.Nanosecond).Nanoseconds()),
		t.Location(),
	)
}

// DaysIn returns the day count in month, accounting for a leap year.
func DaysIn(m time.Month, year int) int {
	return time.Date(
		year,
		m+1,
		0,
		0,
		0,
		0,
		0,
		time.UTC,
	).Day()
}

// IsWeekday returns if the specified time is within standard week days
// ([time.Monday] through [time.Friday] including).
func IsWeekday(t time.Time) bool {
	switch t.Weekday() {
	case time.Saturday, time.Sunday:
		return false
	default:
		return true
	}
}
