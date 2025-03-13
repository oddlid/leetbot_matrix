package ltime

import (
	"fmt"
	"time"
)

func FormatTimeStampFull(t time.Time) string {
	return fmt.Sprintf("[%02d:%02d:%02d:%09d]", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
}

func FormatTimeStampSubSecond(t time.Time) string {
	return fmt.Sprintf("%02d%09d", t.Second(), t.Nanosecond())
}
