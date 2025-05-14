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

type Leet struct {
	configFilePath string
	db             DB
	logger         zerolog.Logger
	tf             ltime.TimeFrame
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

func (l *Leet) logErr(err error) {
	if l == nil || err == nil {
		return
	}
	l.logger.Error().Err(err).Send()
}

func (l *Leet) logErrFn(f func() error) {
	l.logErr(f())
}

func (l *Leet) Play(_ context.Context, w io.Writer, userName string, tfr ltime.TimeFrameResult) error {
	if l == nil {
		return ErrNilReceiver
	}
	// TODO:: make sure to restore this when the round is over and calculations are done
	// l.active.Store(true)

	user := l.db.Users.getUser(userName)
	if user == nil {
		return fmt.Errorf("no such user: %s", userName)
	}

	if l.handleFinishedPlayer(w, user, tfr.TS) {
		return nil
	}

	if l.checkSpam(w, user, tfr.TS) {
		return nil
	}
	// ...
	return nil
}

func (l *Leet) handleFinishedPlayer(w io.Writer, user *User, ts time.Time) bool {
	if !user.Done {
		return false
	}
	// let the playser see the timestamp of posting, to make it extra annoying if it was a good time ;)
	l.logErr(ltime.FormatTimeStampFull(w, ts))
	tDiff := ltime.Diff(l.db.BotStart, user.Entries.Last)
	l.logErr(util.Fpf(
		w,
		": %s - you're done, as you're #%d, reaching %d points @ %s after %d year(s), %d month(s), %d day(s)",
		user.Name,
		l.db.Users.filterByDone(true).sortByLastEntryAsc().getIndex(user.Name)+1,
		user.Scores.Total,
		ltime.FormatLongDate(user.Entries.Last),
		tDiff.Year,
		tDiff.Month,
		tDiff.Day,
	))
	return true
}

func (l *Leet) checkSpam(w io.Writer, user *User, ts time.Time) bool {
	if user.locked.Load() {
		l.logErr(ltime.FormatTimeStampFull(w, ts))
		l.logErr(util.Fpf(w, ": %s - Stop spamming!", user.Name))
		return true
	}
	return false
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

func (l *Leet) Stats(w io.Writer) error {
	if l == nil {
		return ErrNilReceiver
	}

	greet := func(points int) error {
		has, bc := l.db.BonusCfgs.hasValue(points)
		if !has {
			return nil
		}
		return util.Fpf(w, " - %s", bc.Greeting)
	}

	winners := l.db.Users.filterByDone(true).sortByLastEntryAsc()
	win := func(u *User) error {
		if !u.Done {
			return nil
		}
		return util.Fpf(w, " - Winner #%d!", winners.getIndex(u.Name)+1)
	}

	entryFormat := util.GetPadFormat(
		l.db.Users.maxNameLen(),
		": %04d @ %s Best: %s Bonus: %03dx = %04d Tax: %03dx = -%04d Miss: -%04d",
	)

	if err := util.Fpf(w, "Stats since %s:\n", l.db.BotStart.Format(time.RFC3339)); err != nil {
		return err
	}

	for _, u := range l.db.Users.toSlice().sortByPointsDesc() {
		if err := util.Fpf(
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
		); err != nil {
			return err
		}
		if err := win(u); err != nil {
			return err
		}
		if err := greet(u.Scores.Total); err != nil {
			return err
		}
		if err := util.Fpf(w, "\n"); err != nil {
			return err
		}
	}

	return nil
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
	defer l.logErrFn(file.Close)
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
	defer l.logErrFn(file.Close)
	return l.saveConfig(file)
}
