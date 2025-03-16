package leet

import (
	"sync/atomic"

	"github.com/oddlid/leetbot_matrix/ltime"
)

type User struct {
	Name    string          `json:"name"`    // needed to format messages to the user
	Entries ltime.EntryTime `json:"entries"` // imcompatible with old format
	Taxes   ValueTracker    `json:"taxes"`
	Bonuses ValueTracker    `json:"bonuses"`
	Missees ValueTracker    `json:"misses"`
	Scores  ValueTracker    `json:"scores"` // imcompatible with old format
	Done    bool            `json:"done"`   // true when user has reached the target score (was locked in the old format)
	locked  atomic.Bool     // temp lock for spamming in a round
}

// Might be useful, if done a lot
// func (u *User) lock(val bool) {
// 	if u == nil {
// 		return
// 	}
// 	u.locked.Swap(val)
// }
