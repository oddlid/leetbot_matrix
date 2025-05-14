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
	assert.NoError(t, FormatTimeStampFull(&buf, time.Time{}))
	assert.Equal(t, "[00:00:00:000000000]", buf.String())
}

func Test_FormatTimeStampSubSecond(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	assert.NoError(t, FormatTimeStampSubSecond(&buf, time.Time{}))
	assert.Equal(t, "00000000000", buf.String())
}

func Test_GetAdjustedTime(t *testing.T) {
	t.Parallel()

	// Time of message from the server
	mt := time.Date(2025, 5, 12, 13, 37, 0, 111234567, time.UTC)
	// Time of receiving by the bot
	bt := time.Date(2025, 5, 13, 13, 37, 0, 987456789, time.UTC)

	// The naonsecond portion of the returned time should be the first 3 digits of mt subseconds,
	// and the last 6 digits of bt subseconds
	assert.Equal(t, 111456789, GetAdjustedTime(mt, bt).Nanosecond())
}
