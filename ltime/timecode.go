package ltime

type TimeCode uint8

// Constants for signaling offset from time window
const (
	TCInvalid TimeCode = iota // Default/unspecified is not a valid value
	TCBefore                  // more than a minute before
	TCEarly                   // less than a minute before
	TCOnTime                  // within correct minute
	TCLate                    // less than a minute late
	TCAfter                   // more than a minute late
)

const (
	tcNameInvalid = `invalid`
	tcNameBefore  = `before`
	tcNameEarly   = `early`
	tcNameOnTime  = `on time`
	tcNameLate    = `late`
	tcNameAfter   = `after`
)

func (tc TimeCode) String() string {
	switch tc {
	case TCBefore:
		return tcNameBefore
	case TCEarly:
		return tcNameEarly
	case TCOnTime:
		return tcNameOnTime
	case TCLate:
		return tcNameLate
	case TCAfter:
		return tcNameAfter
	default:
		return tcNameInvalid
	}
}

func (tc TimeCode) InsideWindow() bool {
	return tc == TCEarly || tc == TCOnTime || tc == TCLate
}

func (tc TimeCode) NearMiss() bool {
	return tc == TCEarly || tc == TCLate
}
