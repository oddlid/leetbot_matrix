package ltime

import (
	"fmt"
	"io"
	"time"
)

const (
	tmplTSFull   = `[%02d:%02d:%02d:%09d]`
	tmplTSSubSec = `%02d%09d`
)

// don't want to import the util package just for this, so repeating
func fpf(w io.Writer, format string, v ...any) error {
	_, err := fmt.Fprintf(w, format, v...)
	return err
}

func FormatTimeStampFull(w io.Writer, t time.Time) error {
	return fpf(w, tmplTSFull, t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
}

func FormatTimeStampSubSecond(w io.Writer, t time.Time) error {
	return fpf(w, tmplTSSubSec, t.Second(), t.Nanosecond())
}

func FormatLongDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000000000")
}

func FormatShortTime(t time.Time) string {
	return t.Format("15:04:05.000000000")
}

// GetAdjustedTime adjusts the sub-second portion of msgTime to be the first 3 digits of msgTime.Nanosecond()
// plus the last 6 digits of botTime.Nanosecond().
func GetAdjustedTime(msgTime, botTime time.Time) time.Time {
	msgTime = msgTime.Truncate(time.Millisecond) // make sure we set the last 6 digits to 0
	return time.Date(
		msgTime.Year(),
		msgTime.Month(),
		msgTime.Day(),
		msgTime.Hour(),
		msgTime.Minute(),
		msgTime.Second(),
		msgTime.Nanosecond()+botTime.Nanosecond()%int(time.Millisecond),
		msgTime.Location(),
	)
}
