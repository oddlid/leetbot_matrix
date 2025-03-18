package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetPadFormat(t *testing.T) {
	t.Parallel()

	s := GetPadFormat(10, ":")
	ret := fmt.Sprintf(s, "one")
	assert.Equal(t, "one        :", ret)
}
