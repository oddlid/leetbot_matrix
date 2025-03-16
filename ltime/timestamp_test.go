package ltime

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_FormatTimeStampFull(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	FormatTimeStampFull(&buf, time.Time{})
	assert.Equal(t, "[00:00:00:000000000]", buf.String())
}

func Test_FormatTimeStampSubSecond(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	FormatTimeStampSubSecond(&buf, time.Time{})
	assert.Equal(t, "00000000000", buf.String())
}
