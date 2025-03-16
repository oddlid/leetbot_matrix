package leet

import "sync"

type UserData struct {
	Users map[string]*User `json:"users"`
	mu    sync.RWMutex
}

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
