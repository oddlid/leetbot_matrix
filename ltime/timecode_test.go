package ltime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TimeCode_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, tcNameInvalid, TCInvalid.String())
	assert.Equal(t, tcNameBefore, TCBefore.String())
	assert.Equal(t, tcNameEarly, TCEarly.String())
	assert.Equal(t, tcNameOnTime, TCOnTime.String())
	assert.Equal(t, tcNameLate, TCLate.String())
	assert.Equal(t, tcNameAfter, TCAfter.String())
}

func Test_TimeCode_InsideWindow(t *testing.T) {
	t.Parallel()

	assert.False(t, TCInvalid.InsideWindow())
	assert.False(t, TCBefore.InsideWindow())
	assert.False(t, TCAfter.InsideWindow())
	assert.True(t, TCEarly.InsideWindow())
	assert.True(t, TCOnTime.InsideWindow())
	assert.True(t, TCLate.InsideWindow())
}

func Test_TimeCode_NearMiss(t *testing.T) {
	t.Parallel()

	assert.False(t, TCInvalid.NearMiss())
	assert.False(t, TCBefore.NearMiss())
	assert.False(t, TCAfter.NearMiss())
	assert.False(t, TCOnTime.NearMiss())
	assert.True(t, TCEarly.NearMiss())
	assert.True(t, TCLate.NearMiss())
}
