package util

import (
	"fmt"
	"io"
)

func GetPadFormat(alignAt int, format string) string {
	return fmt.Sprintf("%s%d%s", "%-", alignAt, "s "+format)
}

// func Pad(w io.Writer, align int, str string) {
// 	_ = Fpf(w, GetPadFormat(align, ""), str)
// }

func Fpf(w io.Writer, format string, v ...any) error {
	_, err := fmt.Fprintf(w, format, v...)
	return err
}
