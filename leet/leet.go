package leet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/oddlid/leetbot_matrix/ltime"
	"github.com/oddlid/leetbot_matrix/util"
	"github.com/rs/zerolog"
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

func New(logger zerolog.Logger, configFilePath, room string, tf ltime.TimeFrame) *Leet {
	return &Leet{
		configFilePath: configFilePath,
		logger:         logger.With().Str("module", "leet").Logger(),
		tf:             tf,
		db: DB{
			BotStart: time.Now(), // will be overwritten on config load, but needs to be set on first run
			Room:     room,
		},
	}
}

func (l *Leet) Play(_ context.Context, w io.Writer, userName string, tfr ltime.TimeFrameResult) error {
	if l == nil {
		return ErrNilReceiver
	}
	// TODO:: make sure to restore this when the round is over and calculations are done
	l.active.Store(true)

	user := l.db.Users.getUser(userName)
	if user == nil {
		return fmt.Errorf("no such user: %s", userName)
	}

	if user.Done {
		l.handleFinishedPlayer(w, user, tfr.TS)
		return nil
	}
	// ...
	return nil
}

func (l *Leet) handleFinishedPlayer(w io.Writer, user *User, ts time.Time) {
	// let the playser see the timestamp of posting, to make it extra annoying if it was a good time ;)
	ltime.FormatTimeStampFull(w, ts)
	tDiff := ltime.Diff(l.db.BotStart, user.Entries.Last)
	fmt.Fprintf(
		w,
		": %s - you're done, as you're #%d, reaching %d points @ %s after %d year(s), %d month(s), %d day(s)",
		user.Name,
		l.db.Users.filterByDone(true).sortByLastEntryAsc().getIndex(user.Name)+1,
		user.Scores.Total,
		ltime.FormatLongDate(user.Entries.Last),
		tDiff.Year,
		tDiff.Month,
		tDiff.Day,
	)
}

func (l *Leet) SetRoom(id string) error {
	if l == nil {
		return ErrNilReceiver
	}
	// might need a lock on this
	l.db.Room = id
	return nil
}

func (l *Leet) GetRoom() (string, error) {
	if l == nil {
		return "", ErrNilReceiver
	}
	return l.db.Room, nil
}

func (l *Leet) Stats(w io.Writer) {
	if l == nil {
		return
	}

	greet := func(points int) {
		has, bc := l.db.BonusCfgs.hasValue(points)
		if !has {
			return
		}
		fmt.Fprintf(w, " - %s", bc.Greeting)
	}

	winners := l.db.Users.filterByDone(true).sortByLastEntryAsc()
	win := func(u *User) {
		if !u.Done {
			return
		}
		fmt.Fprintf(w, " - Winner #%d!", winners.getIndex(u.Name)+1)
	}

	entryFormat := util.GetPadFormat(
		l.db.Users.maxNameLen(),
		": %04d @ %s Best: %s Bonus: %03dx = %04d Tax: %03dx = -%04d Miss: -%04d",
	)

	fmt.Fprintf(w, "Stats since %s:\n", l.db.BotStart.Format(time.RFC3339))

	for _, u := range l.db.Users.toSlice().sortByPointsDesc() {
		fmt.Fprintf(
			w,
			entryFormat,
			u.Name,
			u.Scores.Total,
			ltime.FormatLongDate(u.Entries.Last),
			ltime.FormatLongDate(u.Entries.Best),
			u.Bonuses.Times,
			u.Bonuses.Total,
			u.Taxes.Times,
			u.Taxes.Total,
			u.Missees.Total,
		)
		win(u)
		greet(u.Scores.Total)
		fmt.Fprintf(w, "\n")
	}
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
