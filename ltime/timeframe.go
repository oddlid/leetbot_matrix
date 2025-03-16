package ltime

import (
	"fmt"
	"time"
)

type TimeFrame struct {
	Hour         uint8
	Minute       uint8
	WindowBefore time.Duration
	WindowAfter  time.Duration
}

type TimeFrameResult struct {
	TF     TimeFrame     // The TimeFrame this result was derived from
	TS     time.Time     // The timestamp used to derive this result
	Code   TimeCode      // Distance and direction indicator
	Offset time.Duration // Offset from target time
}

func (tf TimeFrame) Adjust(t time.Time, adjust time.Duration) TimeFrame {
	then := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		int(tf.Hour),
		int(tf.Minute),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	when := then.Add(adjust)
	return TimeFrame{
		Hour:         uint8(when.Hour()),
		Minute:       uint8(when.Minute()),
		WindowBefore: tf.WindowBefore,
		WindowAfter:  tf.WindowAfter,
	}
}

func (tf TimeFrame) FormatWindowBefore(t time.Time) string {
	wb := tf.Adjust(t, -tf.WindowBefore)
	return time.Date(0, 0, 0, int(wb.Hour), int(wb.Minute), 0, 0, time.Local).Format("15:04")
}

func (tf TimeFrame) FormatWindowAfter(t time.Time) string {
	wa := tf.Adjust(t, tf.WindowAfter)
	return time.Date(0, 0, 0, int(wa.Hour), int(wa.Minute), 0, 0, time.Local).Format("15:04")
}

func (tf TimeFrame) AsCronSpec() string {
	return fmt.Sprintf("0 %d %d * * *", tf.Minute, tf.Hour)
}

// Code returns a TimeFrameResult indicating if the actual time is before or after the target time,
// and the distance to the target time
func (tf TimeFrame) Code(actual time.Time) TimeFrameResult {
	target := time.Date(
		actual.Year(),
		actual.Month(),
		actual.Day(),
		int(tf.Hour),
		int(tf.Minute),
		0, // Should be 0 since that's the target time we're aiming for
		0, // -- || --
		actual.Location(),
	)

	isBefore := actual.Before(target)

	distance := target.Sub(actual)
	if !isBefore {
		distance = actual.Sub(target)
	}

	result := TimeFrameResult{
		TF:     tf,
		TS:     actual,
		Offset: distance,
	}

	if isBefore {
		if distance > tf.WindowBefore {
			result.Code = TCBefore
		} else {
			result.Code = TCEarly
		}
	} else if distance >= tf.WindowAfter*2 {
		result.Code = TCAfter
	} else if distance >= tf.WindowAfter {
		result.Code = TCLate
	} else {
		result.Code = TCOnTime
	}

	return result
}

// GetTargetScore returns how many points needed to win the game, depending on the TimeFrame
// configuration.
func (tf TimeFrame) GetTargetScore() int {
	return int(tf.Hour)*100 + int(tf.Minute)
}
