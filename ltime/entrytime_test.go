package ltime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_EntryTime_Update(t *testing.T) {
	t.Parallel()

	assert.NotPanics(
		t,
		func() {
			(*EntryTime)(nil).Update(TimeFrame{}, time.Time{})
		},
	)

	tf := TimeFrame{
		Hour:         13,
		Minute:       37,
		WindowBefore: time.Minute,
		WindowAfter:  time.Minute,
	}
	tm := time.Date(0, 0, 0, 13, 35, 0, 0, time.UTC)
	et := EntryTime{}

	et.Update(tf, tm)
	assert.True(t, et.Last.IsZero())

	tm = time.Date(0, 0, 0, 13, 36, 0, 0, time.UTC)
	et.Update(tf, tm)
	assert.True(t, et.Last.Equal(tm))
	assert.True(t, et.Best.Equal(tm))
	// Second time with same values should not change anything, but bail out on near miss
	et.Update(tf, tm)
	assert.True(t, et.Last.Equal(tm))
	assert.True(t, et.Best.Equal(tm))

	// Now we need to be on time to make a change
	tm = time.Date(0, 0, 0, 13, 37, 10, 0, time.UTC)
	et.Update(tf, tm)
	assert.True(t, et.Last.Equal(tm))
	assert.True(t, et.Best.Equal(tm))

	// And now we must score closer then before to get an update
	// First round - worse time
	tm = time.Date(0, 0, 0, 13, 37, 11, 0, time.UTC)
	et.Update(tf, tm)
	assert.True(t, et.Last.Equal(tm))
	assert.False(t, et.Best.Equal(tm))
	// Second round - better time
	tm = time.Date(0, 0, 0, 13, 37, 9, 0, time.UTC)
	et.Update(tf, tm)
	assert.True(t, et.Last.Equal(tm))
	assert.True(t, et.Best.Equal(tm))
}
