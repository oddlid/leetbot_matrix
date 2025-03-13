package ltime

import (
	"time"

	"github.com/beevik/ntp"
)

func GetNTPOffSet(server string) (time.Duration, error) {
	res, err := ntp.Query(server)
	if err != nil {
		return 0, err
	}
	return res.ClockOffset, nil
}
