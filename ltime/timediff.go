package ltime

import "time"

type TimeDiff struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

// adapted from: https://stackoverflow.com/a/36531443/1705598
func Diff(a, b time.Time) TimeDiff {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	d := TimeDiff{
		Year:   int(y2 - y1),
		Month:  int(M2 - M1),
		Day:    int(d2 - d1),
		Hour:   int(h2 - h1),
		Minute: int(m2 - m1),
		Second: int(s2 - s1),
	}

	// Normalize negative values
	if d.Second < 0 {
		d.Second += 60
		d.Minute--
	}
	if d.Minute < 0 {
		d.Minute += 60
		d.Hour--
	}
	if d.Hour < 0 {
		d.Hour += 24
		d.Day--
	}
	if d.Day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		d.Day += 32 - t.Day()
		d.Month--
	}
	if d.Month < 0 {
		d.Month += 12
		d.Year--
	}

	return d
}
