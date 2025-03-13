package leet

// ValueTracker is used to keep track of bonuses, taxes, misses and scores
type ValueTracker struct {
	Times int `json:"times"`
	Total int `json:"total"`
}

// Add adds the given value to total (use negative value to subtract),
// and increases the counter for how many updates there have been.
func (vt *ValueTracker) Add(value int) {
	if vt == nil {
		return
	}
	if value != 0 {
		vt.Total += value
		vt.Times++
	}
}
