package bot

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Bot_getOwnUserID(t *testing.T) {
	t.Parallel()

	const (
		server = `test.com`
		user   = `bot`
	)
	b := Bot{
		Username: user,
		Server:   server,
	}

	assert.Empty(t, (*Bot)(nil).getOwnUserID())
	assert.Equal(t, fmt.Sprintf("@%s:%s", user, server), b.getOwnUserID())
}

func Test_Bot_fromSelf(t *testing.T) {
	t.Parallel()

	const user = `@test:test.com`

	assert.False(t, (*Bot)(nil).fromSelf(user))

	b := Bot{}

	assert.False(t, b.fromSelf(""))
	assert.False(t, b.fromSelf(user))
	b.userID = "@bot:test.com"
	assert.False(t, b.fromSelf(""))
	assert.False(t, b.fromSelf(user))
	assert.True(t, b.fromSelf(b.userID))
}
