package leet

import (
	"sort"
	"sync"
)

type UserData struct {
	Users map[string]*User `json:"users"`
	mu    sync.RWMutex
}

type UserSlice []*User

// getUser returns an existing user, or creates a new one with the given ID.
// It returns a nil User if the receiver is nil.
func (ud *UserData) getUser(id string) *User {
	if ud == nil {
		return nil
	}
	// try get existing
	ud.mu.RLock()
	u, ok := ud.Users[id]
	ud.mu.RUnlock()
	if ok {
		return u
	}

	// create new
	ud.mu.Lock()
	defer ud.mu.Unlock()
	u = &User{Name: id}
	ud.Users[id] = u

	return u
}

func (ud *UserData) maxNameLen() int {
	if ud == nil {
		return 0
	}
	max := 0
	for _, v := range ud.Users {
		nlen := len(v.Name)
		if nlen > max {
			max = nlen
		}
	}
	return max
}

func (ud *UserData) toSlice() UserSlice {
	s := make(UserSlice, 0, len(ud.Users))
	for _, v := range ud.Users {
		s = append(s, v)
	}
	return s
}

func (ud *UserData) filterByDone(done bool) UserSlice {
	us := make(UserSlice, 0, len(ud.Users))
	for _, v := range ud.Users {
		if done == v.Done {
			us = append(us, v)
		}
	}
	return us
}

func (us UserSlice) sortByPointsDesc() UserSlice {
	sort.Slice(
		us,
		func(i, j int) bool {
			return us[i].Scores.Total > us[j].Scores.Total
		},
	)
	return us
}

func (us UserSlice) sortByLastEntryAsc() UserSlice {
	sort.Slice(
		us,
		func(i, j int) bool {
			return us[i].Entries.Last.Before(us[j].Entries.Last)
		},
	)
	return us
}

// getIndex returns the position of the user with the given name in the slice, if it exits.
// Used for positioning after having sorted the slice by date, points or whatever
func (us UserSlice) getIndex(name string) int {
	for i, u := range us {
		if u.Name == name {
			return i
		}
	}
	return -1
}
