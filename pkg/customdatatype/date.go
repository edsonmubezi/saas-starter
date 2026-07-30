package customdatatype

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type DateOnly struct {
	time.Time
}

const dateLayout = "2006-01-02"

// UnmarshalJSON parses JSON string into DateOnly (YYYY-MM-DD).
func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" {
		d.Time = time.Time{}
		return nil
	}

	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date format (expected YYYY-MM-DD)")
	}

	d.Time = t
	return nil
}

// MarshalJSON converts DateOnly to JSON string (YYYY-MM-DD).
func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format(dateLayout) + `"`), nil
}

// Value implements driver.Valuer for database storage.
func (d DateOnly) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format(dateLayout), nil
}

// Scan implements sql.Scanner for database reading.
func (d *DateOnly) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case []byte:
		t, err := time.Parse(dateLayout, string(v))
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	case string:
		t, err := time.Parse(dateLayout, v)
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	default:
		return fmt.Errorf("cannot convert %T to DateOnly", value)
	}
}

func NewDateOnlyFromTime(t time.Time) DateOnly {
	year, month, day := t.Date()
	return DateOnly{Time: time.Date(year, month, day, 0, 0, 0, 0, t.Location())}
}

// DifferenceInDays returns the inclusive number of days between two DateOnly values.
// If d2 is before d1, returns 0.
func DifferenceInDays(d1, d2 DateOnly) int {
	// Create date-only times preserving location
	y1, m1, d1d := d1.Date()
	start := time.Date(y1, m1, d1d, 0, 0, 0, 0, d1.Location())

	y2, m2, d2d := d2.Date()
	end := time.Date(y2, m2, d2d, 0, 0, 0, 0, d2.Location())

	diff := end.Sub(start).Hours() / 24
	if diff < 0 {
		return 0
	}
	return int(diff)
}

// WorkingDaysBetween returns the number of working days (Mon–Fri) between
// two DateOnly values, using the same exclusive-end convention as DifferenceInDays.
// If d2 is before or equal to d1, returns 0.
func WorkingDaysBetween(d1, d2 DateOnly) int {
	y1, m1, d1d := d1.Date()
	start := time.Date(y1, m1, d1d, 0, 0, 0, 0, d1.Location())

	y2, m2, d2d := d2.Date()
	end := time.Date(y2, m2, d2d, 0, 0, 0, 0, d2.Location())

	if !end.After(start) {
		return 0
	}

	count := 0
	for cur := start; cur.Before(end); cur = cur.AddDate(0, 0, 1) {
		wd := cur.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
	}
	return count
}

func ParseDateOnly(s string) (DateOnly, error) {
	if s == "" {
		return DateOnly{time.Time{}}, nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return DateOnly{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	return DateOnly{t}, nil
}

func ParseDateOnlyNow() DateOnly {
	now := time.Now()
	year, month, day := now.Date()
	return DateOnly{Time: time.Date(year, month, day, 0, 0, 0, 0, now.Location())}
}

func ParseDateAndTime() DateOnly {
	now := time.Now()
	return DateOnly{Time: now}
}

func MonthsBetween(start, end time.Time) int {
	// Ensure start <= end
	if start.After(end) {
		start, end = end, start
	}

	years := end.Year() - start.Year()
	months := int(end.Month()) - int(start.Month())

	totalMonths := years*12 + months

	// Adjust if the end day is earlier in the month than the start day
	if end.Day() < start.Day() {
		totalMonths--
	}

	return totalMonths
}

func (d DateOnly) ToTime() time.Time {
	return d.Time
}

func GetMonthRange() (DateOnly, DateOnly) {
	now := time.Now()
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	last := first.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return DateOnly{first}, DateOnly{last}
}

func NewDateOnly(t time.Time) DateOnly {
	year, month, day := t.Date()
	return DateOnly{Time: time.Date(year, month, day, 0, 0, 0, 0, t.Location())}
}

func GetMonthRangeFromDate(input DateOnly) (DateOnly, DateOnly) {
	first := time.Date(input.Year(), input.Month(), 1, 0, 0, 0, 0, input.Location())
	last := first.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return DateOnly{first}, DateOnly{last}
}

func (d DateOnly) IsZero() bool {
	return d.Time.IsZero()
}
