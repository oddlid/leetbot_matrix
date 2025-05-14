package leet

import (
	"context"
	"io"
	"time"

	"github.com/oddlid/leetbot_matrix/ltime"
)

type LeetConfig struct {
	InspectionTax int  `json:"inspection_tax"`
	OvershootTax  int  `json:"overshoot_tax"`
	InspectAlways bool `json:"inspect_always"`
	TaxLoners     bool `json:"tax_loners"`
}

type DB struct {
	Room      string       `json:"room"`
	BonusCfgs BonusConfigs `json:"bonus_configs"`
	GameCfg   LeetConfig   `json:"game_config"`
	Users     UserData     `json:"users"`
	BotStart  time.Time    `json:"botstart"`
}

func (db *DB) handleEntry(_ context.Context, w io.Writer, user *User, tfr ltime.TimeFrameResult) {
	//
}
