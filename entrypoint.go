package main

import (
	"os"
	"time"

	"github.com/oddlid/leetbot_matrix/bot"
	"github.com/oddlid/leetbot_matrix/ltime"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

func botEntryPoint(cCtx *cli.Context) error {
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()
	cfg := bot.BotConfig{
		Username:   cCtx.String(optUser),
		Password:   cCtx.String(optPass),
		Server:     cCtx.String(optServer),
		Room:       cCtx.String(optRoom),
		NTPServer:  cCtx.String(optNTPServer),
		DBPath:     cCtx.Path(optDB),
		ConfigFile: cCtx.Path(optConfigFile),
		TimeFrame: ltime.TimeFrame{
			Hour:   uint8(cCtx.Int(optHour)),
			Minute: uint8(cCtx.Int(optMinute)),
			// We currently hard code these, since it's unlikely we'll start this game over with differenct values,
			// but at least it would be easy to add support for other time windows
			WindowBefore: time.Minute,
			WindowAfter:  time.Minute,
		},
	}
	b := bot.New(cfg, l)
	return b.Start(cCtx.Context)
}
