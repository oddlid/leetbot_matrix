package ltime

import (
	"fmt"
	"time"
)

type TimeFrame struct {
	Hour         int
	Minute       int
	WindowBefore time.Duration
	WindowAfter  time.Duration
}

func (tf TimeFrame) Adjust(t time.Time, adjust time.Duration) TimeFrame {
	then := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		tf.Hour,
		tf.Minute,
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	when := then.Add(adjust)
	return TimeFrame{
		Hour:         when.Hour(),
		Minute:       when.Minute(),
		WindowBefore: tf.WindowBefore,
		WindowAfter:  tf.WindowAfter,
	}
}

func (tf TimeFrame) AsCronSpec() string {
	return fmt.Sprintf("%d %d * * *", tf.Hour, tf.Minute)
}

func (tf TimeFrame) Code(t time.Time) TimeCode {
	switch h := t.Hour(); {
	case h < tf.Hour:
		return TCBefore
	case h > tf.Hour:
		return TCAfter
	}

	switch m := t.Minute(); {
	case m < tf.Minute-int(tf.WindowBefore.Minutes()):
		return TCBefore
	case m > tf.Minute+int(tf.WindowAfter.Minutes()):
		return TCAfter
	case m == tf.Minute-int(tf.WindowBefore.Minutes()):
		return TCEarly
	case m == tf.Minute+int(tf.WindowAfter.Minutes()):
		return TCLate
	default:
		return TCOnTime
	}
}

// GetTargetScore returns how many points needed to win the game, depending on the TimeFrame
// configuration.
func (tf TimeFrame) GetTargetScore() int {
	return tf.Hour*100 + tf.Minute
}

// Distance returns how far off the passed time is.
// It does not indicate if the time is before or after, just how close.
// Combine with TimeFrame.Code to get if it's before or after.
func (tf TimeFrame) Distance(actual time.Time) time.Duration {
	target := time.Date(
		actual.Year(),
		actual.Month(),
		actual.Day(),
		tf.Hour,
		tf.Minute,
		0, // Should be 0 since that's the target time we're aiming for
		0, // -- || --
		actual.Location(),
	)
	if actual.After(target) {
		return actual.Sub(target)
	}
	return target.Sub(actual)
}
