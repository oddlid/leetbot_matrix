package leet

import (
	"strings"
	"testing"
	"time"

	"github.com/oddlid/leetbot_matrix/ltime"
)

// visual inspection of output
func Test_Leet_Stats(t *testing.T) {
	t.Parallel()

	l := Leet{
		db: DB{
			BotStart: time.Now(),
			Users: UserData{
				Users: map[string]*User{
					"@short:test.com": {
						Name: "@short:test.com",
					},
					"@prettymuchlonger:test.com": {
						Name: "@prettymuchlonger:test.com",
					},
				},
			},
		},
	}

	var buf strings.Builder
	l.Stats(&buf)
	t.Log(buf.String())
}

// visual inspection of output
func Test_Leet_handleFinishedPlayer(t *testing.T) {
	t.Parallel()

	const userName = `TestUser`

	u := User{
		Name:    userName,
		Entries: ltime.EntryTime{Last: time.Date(2025, 3, 18, 13, 37, 0, 0, time.UTC)},
		Scores:  ValueTracker{Total: 1337},
		Done:    true,
	}
	l := Leet{
		db: DB{
			BotStart: time.Date(2018, 1, 13, 13, 37, 0, 0, time.UTC),
			Users: UserData{
				Users: map[string]*User{
					userName: &u,
				},
			},
		},
	}
	var buf strings.Builder
	l.handleFinishedPlayer(&buf, &u, time.Now())
	t.Log(buf.String())
}
