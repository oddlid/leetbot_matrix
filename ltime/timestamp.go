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

func FormatTimeStampFull(w io.Writer, t time.Time) {
	fmt.Fprintf(w, tmplTSFull, t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
}

func FormatTimeStampSubSecond(w io.Writer, t time.Time) {
	fmt.Fprintf(w, tmplTSSubSec, t.Second(), t.Nanosecond())
}

func FormatLongDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000000000")
}

func FormatShortTime(t time.Time) string {
	return t.Format("15:04:05.000000000")
}
