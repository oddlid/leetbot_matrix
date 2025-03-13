package ltime

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type FakeWriter struct {
	Err error
}

func (fw *FakeWriter) Write(b []byte) (int, error) {
	if fw.Err != nil {
		return 0, fw.Err
	}
	return len(b), nil
}

func Test_FormatTimeStampFull_whenError(t *testing.T) {
	t.Parallel()

	fw := FakeWriter{
		Err: errors.New("test error"),
	}

	err := FormatTimeStampFull(&fw, time.Time{})
	assert.ErrorIs(t, err, fw.Err)
}

func Test_FormatTimeStampFull(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	assert.NoError(t, FormatTimeStampFull(&buf, time.Time{}))
	assert.Equal(t, "[00:00:00:000000000]", buf.String())
}

func Test_FormatTimeStampSubSecond_whenError(t *testing.T) {
	t.Parallel()

	fw := FakeWriter{
		Err: errors.New("test error"),
	}

	err := FormatTimeStampSubSecond(&fw, time.Time{})
	assert.ErrorIs(t, err, fw.Err)
}

func Test_FormatTimeStampSubSecond(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	assert.NoError(t, FormatTimeStampSubSecond(&buf, time.Time{}))
	assert.Equal(t, "00000000000", buf.String())
}
