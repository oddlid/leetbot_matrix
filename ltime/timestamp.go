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

func FormatTimeStampFull(w io.Writer, t time.Time) error {
	_, err := fmt.Fprintf(w, tmplTSFull, t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
	return err
}

func FormatTimeStampSubSecond(w io.Writer, t time.Time) error {
	_, err := fmt.Fprintf(w, tmplTSSubSec, t.Second(), t.Nanosecond())
	return err
}
