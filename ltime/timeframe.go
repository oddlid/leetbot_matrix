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

func (tf TimeFrame) AsCronSpec() string {
	return fmt.Sprintf("0 %d %d * * *", tf.Minute, tf.Hour)
}

func (tf TimeFrame) Code(t time.Time) TimeCode {
	switch h := uint8(t.Hour()); {
	case h < tf.Hour:
		return TCBefore
	case h > tf.Hour:
		return TCAfter
	}

	switch m := uint8(t.Minute()); {
	case m < uint8(tf.Minute)-uint8(tf.WindowBefore.Minutes()):
		return TCBefore
	case m > uint8(tf.Minute)+uint8(tf.WindowAfter.Minutes()):
		return TCAfter
	case m == uint8(tf.Minute)-uint8(tf.WindowBefore.Minutes()):
		return TCEarly
	case m == uint8(tf.Minute)+uint8(tf.WindowAfter.Minutes()):
		return TCLate
	default:
		return TCOnTime
	}
}

// GetTargetScore returns how many points needed to win the game, depending on the TimeFrame
// configuration.
func (tf TimeFrame) GetTargetScore() int {
	return int(tf.Hour)*100 + int(tf.Minute)
}

// Distance returns how far off the passed time is.
// It does not indicate if the time is before or after, just how close.
// Combine with TimeFrame.Code to get if it's before or after.
func (tf TimeFrame) Distance(actual time.Time) time.Duration {
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
	if actual.After(target) {
		return actual.Sub(target)
	}
	return target.Sub(actual)
}
