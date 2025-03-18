package util

import (
	"fmt"
	"io"
)

func GetPadFormat(alignAt int, format string) string {
	return fmt.Sprintf("%s%d%s", "%-", alignAt, "s "+format)
}

func Pad(w io.Writer, align int, str string) {
	fmt.Fprintf(w, GetPadFormat(align, ""), str)
}
