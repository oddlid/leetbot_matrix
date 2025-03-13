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
	Username        string
	Password        string
	Server          string
	DBPath          string
	ScoreFile       string
	BonusConfigFile string
	NTPServer       string
	TimeFrame       ltime.TimeFrame
}
type Bot struct {
	client     *mautrix.Client
	cron       *cron.Cron
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
		b.cron = cron.New()
	}

	b.log().Debug().Msg("Adding cron job for NTP queries...")
	_, err := b.cron.AddFunc(
		b.cfg.TimeFrame.Adjust(time.Now(), -2*time.Minute).AsCronSpec(),
		func() {
			b.log().Debug().Str("ntp_server", b.cfg.NTPServer).Msg("Querying NTP server...")
			offset, err := ltime.GetNTPOffSet(b.cfg.NTPServer)
			if err != nil {
				b.log().Error().Err(err).Str("ntp_server", b.cfg.NTPServer).Msg("Failed to query NTP server")
				return
			}
			if err = b.send(ctx, fmt.Sprintf("NTP offset from %q: %+v", b.cfg.NTPServer, offset)); err != nil {
				b.log().Error().Err(err).Msg("Failed to send NTP info to room")
			}
		},
	)
	if err != nil {
		return err
	}
	b.cron.Start()
	return nil
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
	_, err := fmt.Fprintf(w, "TODO: show stats")
	return err
}

func (b *Bot) reloadConfig(_ context.Context, w io.Writer) error {
	_, err := fmt.Fprintf(w, "TODO: reload config")
	return err
}

func (b *Bot) play(_ context.Context, w io.Writer, ts time.Time, user string) error {
	if err := ltime.FormatTimeStampFull(w, ts); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, ": %s - TODO: spell et spell", user); err != nil {
		return err
	}
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
		ts := time.Now() // save timestamp asap, before any other processing
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

	if err = b.scheduleNTPCheck(ctx); err != nil {
		b.log().Error().Err(err).Msg("Failed to schedule NTP check")
	}

	<-ctx.Done()
	b.log().Info().Msg("Shutting down...")

	if b.cron != nil {
		b.cron.Stop()
	}

	if err = cryptoHelper.Close(); err != nil {
		b.log().Error().Err(err).Msg("Failed to close cryptoHelper")
	}

	return nil
}
