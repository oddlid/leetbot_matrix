package bot

// See eample at: https://github.com/mautrix/go/blob/main/example/main.go

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oddlid/leetbot_matrix/leet"
	"github.com/oddlid/leetbot_matrix/ltime"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const (
	subCmdStats  = `stats`
	subCmdReload = `reload`
)

var (
	ErrNilClient   = errors.New("client is nil")
	ErrNilReceiver = errors.New("receiver is nil")
	ErrNoRoomID    = errors.New("no room ID set")
	ErrNoNTPServer = errors.New("no NTP server specified")
)

type BotConfig struct {
	Username   string
	Password   string
	Server     string
	DBPath     string
	ConfigFile string
	NTPServer  string
	TimeFrame  ltime.TimeFrame
}
type Bot struct {
	client     *mautrix.Client
	cron       *cron.Cron
	leet       *leet.Leet
	command    string
	userID     string
	lastRoomID id.RoomID
	cfg        BotConfig
	logger     zerolog.Logger
}

func New(cfg BotConfig, logger zerolog.Logger) *Bot {
	return &Bot{
		cfg:     cfg,
		command: fmt.Sprintf("!%d%d", cfg.TimeFrame.Hour, cfg.TimeFrame.Minute),
		userID:  fmt.Sprintf("@%s:%s", cfg.Username, cfg.Server),
		logger:  logger, // adjust later
		leet:    leet.New(logger, cfg.ConfigFile, cfg.TimeFrame),
	}
}

func (b *Bot) log() *zerolog.Logger {
	return &b.logger
}

func (b *Bot) scheduleNTPCheck(ctx context.Context) error {
	if b.cfg.NTPServer == "" {
		return ErrNoNTPServer
	}
	if b.cron == nil {
		b.cron = cron.New(cron.WithSeconds())
	}

	llog := b.log().With().Str("ntp_server", b.cfg.NTPServer).Logger()
	cronSpec := b.cfg.TimeFrame.Adjust(time.Now(), -2*time.Minute).AsCronSpec()
	llog.Debug().Str("cron_spec", cronSpec).Msg("Adding cron job for NTP queries")

	_, err := b.cron.AddFunc(
		cronSpec,
		func() {
			llog.Debug().Msg("Querying NTP server...")
			offset, err := ltime.GetNTPOffSet(b.cfg.NTPServer)
			if err != nil {
				llog.Error().Err(err).Msg("Failed to query NTP server")
				b.leet.SetNTPOffset(0)
				return
			}
			b.leet.SetNTPOffset(offset)
			if err = b.send(ctx, fmt.Sprintf("NTP offset from %q: %+v", b.cfg.NTPServer, offset)); err != nil {
				llog.Error().Err(err).Msg("Failed to send NTP info to room")
			}
		},
	)
	return err
}

func (b *Bot) scheduleConfigSave() error {
	if b.cron == nil {
		b.cron = cron.New(cron.WithSeconds())
	}

	llog := b.log().With().Str("config_file", b.cfg.ConfigFile).Logger()
	cronSpec := b.cfg.TimeFrame.Adjust(time.Now(), 3*time.Minute).AsCronSpec()
	llog.Debug().Str("cron_spec", cronSpec).Msg("Adding cron job for saving config")

	_, err := b.cron.AddFunc(
		cronSpec,
		func() {
			llog.Debug().Msg("Saving config file...")
			if err := b.leet.SaveConfigFile(); err != nil {
				llog.Error().Err(err).Msg("Failed to save config!")
			}
		},
	)

	return err
}

func (b *Bot) fromSelf(user string) bool {
	if b == nil {
		return false
	}
	return user != "" && b.userID != "" && user == b.userID
}

func (b *Bot) setRoom(roomID id.RoomID) {
	if b == nil {
		return
	}
	b.lastRoomID = roomID
}

func (b *Bot) send(ctx context.Context, msg string) error {
	if b.lastRoomID == "" {
		return ErrNoRoomID
	}

	if b.client == nil {
		return ErrNilClient
	}

	_, err := b.client.SendText(ctx, b.lastRoomID, msg)
	return err
}

func (b *Bot) getStats(_ context.Context, w io.Writer) error {
	if b.leet.Active() {
		fmt.Fprintf(w, "Calculation in progress, please try later")
		return nil
	}
	fmt.Fprintf(w, "TODO: show stats")
	return nil
}

func (b *Bot) reloadConfig(_ context.Context, w io.Writer) error {
	if b.leet.Active() {
		fmt.Fprintf(w, "Calculation in progress, please try later")
		return nil
	}

	err := b.leet.LoadConfigFile()
	if err != nil {
		fmt.Fprintf(w, "Failed to reload config. Please check logs.")
	} else {
		fmt.Fprintf(w, "Config reloaded successfully.")
	}
	return err
}

func (b *Bot) play(_ context.Context, w io.Writer, ts time.Time, user string) error {
	res := b.cfg.TimeFrame.Code(ts)
	if !res.Code.InsideWindow() {
		ltime.FormatTimeStampFull(w, ts)
		fmt.Fprintf(
			w,
			" Check your watch, %s! I will only respond to this command between %s and %s.",
			user,
			b.cfg.TimeFrame.FormatWindowBefore(ts),
			b.cfg.TimeFrame.FormatWindowAfter(ts),
		)
		return nil
	}

	ltime.FormatTimeStampFull(w, ts)
	fmt.Fprintf(w, ": %s - TODO: spell et spell", user)
	return nil
}

func (b *Bot) dispatch(ctx context.Context, ts time.Time, user, cmd string) error {
	if b == nil {
		return ErrNilReceiver
	}

	if b.fromSelf(user) {
		b.log().Debug().Str("user", user).Msg("Ignoring message from myself")
		return nil
	}
	if !strings.HasPrefix(cmd, b.command) {
		b.log().Debug().Str("user", user).Str("msg", cmd).Msg("Ignoring message without required prefix")
		return nil
	}

	cmds := strings.Split(cmd, " ")
	b.log().Debug().Strs("cmds", cmds).Send()

	var buf strings.Builder

	if len(cmds) > 1 {
		switch s := cmds[1]; s {
		case subCmdStats:
			if err := b.getStats(ctx, &buf); err != nil {
				return err
			}
			return b.send(ctx, buf.String())
		case subCmdReload:
			if err := b.reloadConfig(ctx, &buf); err != nil {
				return err
			}
			return b.send(ctx, buf.String())
		default:
			buf.WriteString("Invalid subcommand(s): ")
			buf.WriteString(strings.Join(cmds[1:], " "))
			return b.send(ctx, buf.String())
		}
	}

	if err := b.play(ctx, &buf, ts, user); err != nil {
		return err
	}
	return b.send(ctx, buf.String())
}

func (b *Bot) Start(ctx context.Context) error {
	if b == nil {
		return ErrNilReceiver
	}

	b.log().Info().Msg("Initializing...")

	// Find true address of server, in case of delegation.
	cwk, err := mautrix.DiscoverClientAPI(ctx, b.cfg.Server)
	if err != nil {
		return err
	}

	b.client, err = mautrix.NewClient(cwk.Homeserver.BaseURL, "", "")
	if err != nil {
		return err
	}
	b.client.Log = b.logger
	// adjust the bot logger now, after having passed on a clean copy to the client
	b.logger = b.logger.With().Str("bot", b.userID).Logger()

	syncer := b.client.Syncer.(*mautrix.DefaultSyncer) // TODO: check cast

	syncer.OnEventType(event.EventMessage, func(ctx context.Context, evt *event.Event) {
		ts := time.Now().Add(b.leet.GetNTPOffset()) // save timestamp asap, before any other processing
		b.setRoom(evt.RoomID)
		if err := b.dispatch(ctx, ts, evt.Sender.String(), evt.Content.AsMessage().Body); err != nil {
			b.log().Error().Err(err).Msg("Dispatch failed")
		}
	})

	syncer.OnEventType(event.StateMember, func(ctx context.Context, evt *event.Event) {
		if evt.GetStateKey() == b.client.UserID.String() {
			switch evt.Content.AsMember().Membership {
			case event.MembershipInvite:
				_, err := b.client.JoinRoomByID(ctx, evt.RoomID)
				if err != nil {
					b.log().Error().Err(err).
						Str("room_id", evt.RoomID.String()).
						Str("inviter", evt.Sender.String()).
						Msg("Failed to join room after invite")
				} else {
					b.setRoom(evt.RoomID)
					b.log().Info().
						Str("room_id", evt.RoomID.String()).
						Str("inviter", evt.Sender.String()).
						Msg("Joined room after invite")
				}
			case event.MembershipJoin:
				b.setRoom(evt.RoomID)
				b.log().Info().
					Str("room_id", evt.RoomID.String()).
					Str("inviter", evt.Sender.String()).
					Msg("Joined room")
			}
		}
	})

	cryptoHelper, err := cryptohelper.NewCryptoHelper(b.client, []byte("1337"), b.cfg.DBPath)
	if err != nil {
		return err
	}

	cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: b.cfg.Username},
		Password:   b.cfg.Password,
	}

	if err = cryptoHelper.Init(ctx); err != nil {
		return err
	}
	b.client.Crypto = cryptoHelper

	go func() {
		if err := b.client.SyncWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
			b.log().Error().Err(err).Msg("SyncWithContext failed")
		}
	}()

	if err = b.leet.LoadConfigFile(); err != nil {
		b.log().Error().Err(err).Msg("Failed to load config file!")
	}

	if err = b.scheduleNTPCheck(ctx); err != nil {
		b.log().Error().Err(err).Msg("Failed to schedule NTP check!")
	}
	if err = b.scheduleConfigSave(); err != nil {
		b.log().Error().Err(err).Msg("Failed to schedule saving of config!")
	}

	if b.cron != nil {
		b.cron.Start()
		for _, e := range b.cron.Entries() {
			b.log().Debug().Int("cron_entry_id", int(e.ID)).Time("next", e.Next).Msg("Cron entry")
		}
	}

	b.log().Info().Msg("Ready to rock!")
	<-ctx.Done()
	b.log().Info().Msg("Shutting down...")

	if b.cron != nil {
		b.log().Debug().Msg("Stopping cron jobs...")
		b.cron.Stop()
	}

	b.log().Debug().Msg("Closing Crypto Helper...")
	if err = cryptoHelper.Close(); err != nil {
		b.log().Error().Err(err).Msg("Failed to close cryptoHelper")
	}

	b.log().Debug().Msg("Saving config to file...")
	if err = b.leet.SaveConfigFile(); err != nil {
		b.log().Error().Err(err).Msg("Failed to save config, all changes in this session are now lost!")
	}

	b.log().Info().Msg("Done!")
	return nil
}
