package main

import (
	"fmt"
	"os"

	"github.com/oddlid/leetbot_matrix/bot"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

func botEntryPoint(cCtx *cli.Context) error {
	b := bot.Bot{
		Username: cCtx.String(optUser),
		Password: cCtx.String(optPass),
		Server:   cCtx.String(optServer),
		DBPath:   cCtx.Path(optDB),
		Command:  fmt.Sprintf("!%d%d", cCtx.Int(optHour), cCtx.Int(optMinute)),
		Logger:   zerolog.New(os.Stdout),
	}
	return b.TestLogin(cCtx.Context)
}
