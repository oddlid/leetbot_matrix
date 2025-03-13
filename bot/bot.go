package bot

// See eample at: https://github.com/mautrix/go/blob/main/example/main.go

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oddlid/leetbot_matrix/ltime"
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
)

type Bot struct {
	Username   string
	Password   string
	Server     string
	DBPath     string
	Command    string
	userID     string
	lastRoomID id.RoomID
	Logger     zerolog.Logger
	client     *mautrix.Client
}

func (b *Bot) getOwnUserID() string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("@%s:%s", b.Username, b.Server)
}

func (b *Bot) fromSelf(user string) bool {
	if b == nil {
		return false
	}
	return user != "" && b.userID != "" && user == b.userID
}

func (b *Bot) send(ctx context.Context, msg string) error {
	if b == nil {
		return ErrNilReceiver
	}

	if b.lastRoomID == "" {
		return ErrNoRoomID
	}

	if b.client == nil {
		return ErrNilClient
	}

	_, err := b.client.SendText(ctx, b.lastRoomID, msg)
	return err
}

func (b *Bot) dispatch(ctx context.Context, ts time.Time, user, cmd string) error {
	if b == nil {
		return ErrNilReceiver
	}

	if b.fromSelf(user) {
		b.Logger.Debug().Str("user", user).Msg("Ignoring message from myself")
		return nil
	}
	if !strings.HasPrefix(cmd, b.Command) {
		b.Logger.Debug().Str("user", user).Str("msg", cmd).Msg("Ignoring message without required prefix")
		return nil
	}

	cmds := strings.Split(cmd, " ")
	b.Logger.Debug().Strs("cmds", cmds).Send()

	if len(cmds) > 1 {
		switch s := cmds[1]; s {
		case subCmdStats:
			return b.send(ctx, "TODO: show stats")
		case subCmdReload:
			return b.send(ctx, "TODO: reload score file")
		default:
			return b.send(ctx, fmt.Sprintf("Invalid subcommand(s): %s", cmds[1:]))
		}
	}

	return b.send(ctx, fmt.Sprintf("%s: %s - TODO: spell et spell", ltime.FormatTimeStampFull(ts), user))
}

func (b *Bot) TestLogin(ctx context.Context) error {
	if b == nil {
		return ErrNilReceiver
	}
	var err error
	b.client, err = mautrix.NewClient(b.Server, "", "")
	if err != nil {
		return err
	}
	b.client.Log = b.Logger
	b.userID = b.getOwnUserID()

	syncer := b.client.Syncer.(*mautrix.DefaultSyncer) // TODO: check cast

	syncer.OnEventType(event.EventMessage, func(ctx context.Context, evt *event.Event) {
		ts := time.Now() // save timestamp asap, before any other processing
		b.lastRoomID = evt.RoomID
		if err := b.dispatch(ctx, ts, evt.Sender.String(), evt.Content.AsMessage().Body); err != nil {
			b.Logger.Error().Err(err).Msg("Dispatch failed")
		}
		// body := evt.Content.AsMessage().Body
		// if strings.HasPrefix(body, "!1337") {
		// 	resp, err := client.SendText(ctx, evt.RoomID, "kvekkk!!!!")
		// 	if err != nil {
		// 		b.Logger.Error().Err(err).Msg("Failed to send text")
		// 	} else {
		// 		b.Logger.Info().Str("event_id", resp.EventID.String()).Msg("Event sent")
		// 	}
		// }
		// b.Logger.Info().
		// 	Str("sender", evt.Sender.String()).
		// 	Str("type", evt.Type.String()).
		// 	Str("id", evt.ID.String()).
		// 	Str("body", body).
		// 	Msg("Received message")
	})

	syncer.OnEventType(event.StateMember, func(ctx context.Context, evt *event.Event) {
		if evt.GetStateKey() == b.client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite {
			_, err := b.client.JoinRoomByID(ctx, evt.RoomID)
			if err != nil {
				b.Logger.Error().Err(err).
					Str("room_id", evt.RoomID.String()).
					Str("inviter", evt.Sender.String()).
					Msg("Failed to join room after invite")
			} else {
				b.lastRoomID = evt.RoomID
				b.Logger.Info().
					Str("room_id", evt.RoomID.String()).
					Str("inviter", evt.Sender.String()).
					Msg("Joined room after invite")
			}
		}
	})

	cryptoHelper, err := cryptohelper.NewCryptoHelper(b.client, []byte("1337"), b.DBPath)
	if err != nil {
		return err
	}

	cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: b.Username},
		Password:   b.Password,
	}

	if err = cryptoHelper.Init(ctx); err != nil {
		return err
	}
	b.client.Crypto = cryptoHelper

	go func() {
		if err := b.client.SyncWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
			b.Logger.Error().Err(err).Msg("SyncWithContext failed")
		}
	}()

	<-ctx.Done()
	b.Logger.Info().Msg("Shutting down...")

	if err = cryptoHelper.Close(); err != nil {
		b.Logger.Error().Err(err).Msg("Failed to close cryptoHelper")
	}

	return nil
}
