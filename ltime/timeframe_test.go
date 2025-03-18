package ltime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_TimeFrame_Adjust(t *testing.T) {
	t.Parallel()

	tf1 := TimeFrame{
		Hour:         13,
		Minute:       37,
		WindowBefore: 1 * time.Minute,
		WindowAfter:  1 * time.Minute,
	}
	tf2 := tf1.Adjust(time.Now(), 1*time.Minute)
	assert.Equal(t, tf1.Hour, tf2.Hour)
	assert.Equal(t, tf1.Minute+1, tf2.Minute)

	tf2 = tf1.Adjust(time.Now(), 30*time.Minute)
	assert.Equal(t, tf1.Hour+1, tf2.Hour)
	assert.Equal(t, tf1.Minute-30, tf2.Minute)
}

func Test_TimeFrame_AsCronSpec(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "0 37 13 * * *", TimeFrame{Hour: 13, Minute: 37}.AsCronSpec())
}

func Test_TimeFrame_Code(t *testing.T) {
	t.Parallel()

	tf := TimeFrame{
		Hour:         13,
		Minute:       37,
		WindowBefore: time.Minute,
		WindowAfter:  time.Minute,
	}

	tm := time.Date(0, 0, 0, 12, 37, 0, 0, time.UTC)
	res := tf.Code(tm)
	assert.Equal(t, TCBefore, res.Code)
	assert.Equal(t, time.Hour, res.Offset)

	tm = time.Date(0, 0, 0, 14, 37, 0, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCAfter, res.Code)
	assert.Equal(t, time.Hour, res.Offset)

	tm = time.Date(0, 0, 0, 13, 35, 0, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCBefore, res.Code)
	assert.Equal(t, 2*time.Minute, res.Offset)

	tm = time.Date(0, 0, 0, 13, 39, 0, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCAfter, res.Code)
	assert.Equal(t, 2*time.Minute, res.Offset)

	tm = time.Date(0, 0, 0, 13, 36, 0, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCEarly, res.Code)
	assert.Equal(t, time.Minute, res.Offset)

	tm = time.Date(0, 0, 0, 13, 36, 59, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCEarly, res.Code)
	assert.Equal(t, time.Second, res.Offset)

	tm = time.Date(0, 0, 0, 13, 38, 0, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCLate, res.Code)
	assert.Equal(t, time.Minute, res.Offset)

	tm = time.Date(0, 0, 0, 13, 38, 59, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCLate, res.Code)
	assert.Equal(t, 119*time.Second, res.Offset)

	tm = time.Date(0, 0, 0, 13, 37, 0, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCOnTime, res.Code)
	assert.Equal(t, time.Duration(0), res.Offset)

	tm = time.Date(0, 0, 0, 13, 37, 59, 0, time.UTC)
	res = tf.Code(tm)
	assert.Equal(t, TCOnTime, res.Code)
	assert.Equal(t, 59*time.Second, res.Offset)
}

func Test_TimeFrame_GetTargetScore(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 1337, TimeFrame{Hour: 13, Minute: 37}.GetTargetScore())
	assert.Equal(t, 1214, TimeFrame{Hour: 12, Minute: 14}.GetTargetScore())
	assert.Equal(t, 214, TimeFrame{Hour: 2, Minute: 14}.GetTargetScore())
}
