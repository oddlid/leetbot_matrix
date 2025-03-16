package leet

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/oddlid/leetbot_matrix/ltime"
	"github.com/rs/zerolog"
)

type LeetConfig struct {
	InspectionTax int  `json:"inspection_tax"`
	OvershootTax  int  `json:"overshoot_tax"`
	InspectAlways bool `json:"inspect_always"`
	TaxLoners     bool `json:"tax_loners"`
}

type DB struct {
	BonusCfgs BonusConfigs `json:"bonus_configs"`
	GameCfg   LeetConfig   `json:"game_config"`
	Users     UserData     `json:"users"`
	BotStart  time.Time    `json:"botstart"`
}

type Leet struct {
	configFilePath string
	db             DB
	logger         zerolog.Logger
	tf             ltime.TimeFrame
	ntpOffset      atomic.Int64
	active         atomic.Bool // true when between the time of first score giving entry and round calculation done
}

var (
	ErrNilReceiver  = errors.New("receiver is nil")
	ErrNoConfigFile = errors.New("no config file path given")
)

func New(logger zerolog.Logger, configFilePath string, tf ltime.TimeFrame) *Leet {
	return &Leet{
		configFilePath: configFilePath,
		logger:         logger.With().Str("module", "leet").Logger(),
		tf:             tf,
		db: DB{
			BotStart: time.Now(), // will be overwritten on config load, but needs to be set on first run
		},
	}
}

func (l *Leet) Play(_ context.Context, w io.Writer, user string, ts time.Time) error {
	if l == nil {
		return ErrNilReceiver
	}
	l.active.Store(true)
	// ...
	return nil
}

func (l *Leet) GetNTPOffset() time.Duration {
	if l == nil {
		return 0
	}
	return time.Duration(l.ntpOffset.Load())
}

func (l *Leet) SetNTPOffset(d time.Duration) {
	if l == nil {
		return
	}
	l.ntpOffset.Store(int64(d))
}

func (l *Leet) Active() bool {
	if l == nil {
		return false
	}
	return l.active.Load()
}

func (l *Leet) loadConfig(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &l.db)
}

func (l *Leet) LoadConfigFile() error {
	if l == nil {
		return ErrNilReceiver
	}
	if l.configFilePath == "" {
		return ErrNoConfigFile
	}
	file, err := os.Open(l.configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return l.loadConfig(file)
}

func (l *Leet) saveConfig(w io.Writer) error {
	data, err := json.MarshalIndent(&l.db, "", "  ") // save in pretty format, to make it easier to update config by hand
	if err != nil {
		return err
	}
	// add newline
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

func (l *Leet) SaveConfigFile() error {
	if l == nil {
		return ErrNilReceiver
	}
	if l.configFilePath == "" {
		return ErrNoConfigFile
	}
	file, err := os.Create(l.configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return l.saveConfig(file)
}
