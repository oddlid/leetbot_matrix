package ltime

import "time"

type EntryTime struct {
	Last time.Time `json:"last_entry"`
	Best time.Time `json:"best_entry"`
}

func (et *EntryTime) Update(tf TimeFrame, t time.Time) {
	if et == nil {
		return
	}

	res := tf.Code(t)

	// Don't update anything if outside time window
	if !res.Code.InsideWindow() {
		return
	}

	et.Last = t

	// First time, nothing to compare to
	if et.Best.IsZero() {
		et.Best = t
		return
	}

	// We don't want to overwrite the previous best with this one, if this one
	// is closer, but a near miss
	if res.Code.NearMiss() {
		return
	}

	// If the previous best was set regardless because it was the first entry, it might
	// be a near miss, and if so, we know that this entry is better without comparing more
	resBest := tf.Code(et.Best)
	if resBest.Code.NearMiss() {
		et.Best = t
		return
	}

	// Now we know both the previous best and this entry is on time, so we must compare
	// which is closest, and set the best
	if res.Offset < resBest.Offset {
		et.Best = t
	}
}
