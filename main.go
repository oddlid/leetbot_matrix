package main

/*
The intention when starting to write this, was to replicate my IRC `dvdgbot` 1337 module, for playing the same on Matrix.
But after playing a bit around with it, it seems the whole concept is just not suited for Matrix, due to the whole
"eventually consistent" thing, and since federation some times take ages, or fully stop for a while.
When the game concept is to send a specific message as close to a specific timestamp as possible, rewarding the closest,
and also to have the possibility of entry only open within a short time window (2 minutes), it all really falls flat when you
can never know in time if there are still some messages not yet received from other servers.
But, maybe I'll finish it anyways, and it will only be playable for users on the same server as the bot. Just as an exercise.
We'll see, if I have time and nothing better to do.
*/

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oddlid/leetbot_matrix/util"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

const (
	logTimeStampLayout = `2006-01-02T15:04:05.999-07:00`
	appName            = `leetbot_matrix`
	defaultHomeServer  = `oddware.net`
	defaultUser        = `leetbot`
	defaultDB          = `leetbot_matrix.db`
	defaultConfigFile  = `/tmp/leetbot_config.json`
	defaultHour        = 13
	defaultMinute      = 37
	envServer          = `M_HOMESERVER`
	envUser            = `M_USER`
	envPass            = `M_PASS`
	envDB              = `M_DB`
	envLogLevel        = `L_LOGLEVEL`
	envHour            = `L_HOUR`
	envMinute          = `L_MINUTE`
	envConfigFile      = `L_CONFIGFILE`
	optServer          = `server`
	optRoom            = `room`
	optUser            = `user`
	optPass            = `pass`
	optDB              = `db`
	optLogLevel        = `log-level`
	optHour            = `hour`
	optMinute          = `minute`
	optConfigFile      = `config`
)

var (
	Version   string
	BuildDate string
	CommitID  string
)

func getBuildDate() time.Time {
	ts, err := time.Parse(time.RFC3339, BuildDate)
	if err != nil {
		return time.Time{}
	}
	return ts
}

func getVersion() string {
	if Version != "" && CommitID != "" {
		return fmt.Sprintf("%s_%s", Version, CommitID)
	}
	return time.Time{}.Format("2006-01-02_15:04:05")
}

func app() *cli.App {
	return &cli.App{
		Compiled:  getBuildDate(),
		Name:      appName,
		Version:   getVersion(),
		Copyright: fmt.Sprintf("(C) 2025  - %d, Odd E. Ebbesen", time.Now().Year()),
		Authors: []*cli.Author{
			{
				Name:  "Odd E. Ebbesen",
				Email: "git@oddware.net",
			},
		},
		Usage: "Run leet bot for Matrix",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    optServer,
				Aliases: []string{"s"},
				Usage:   "Matrix homeserver `address` to connect to",
				Value:   defaultHomeServer,
				EnvVars: []string{envServer},
			},
			&cli.StringFlag{
				Name:    optRoom,
				Aliases: []string{"r"},
				Usage:   "Which `room` to try to join",
			},
			&cli.StringFlag{
				Name:    optUser,
				Aliases: []string{"u"},
				Usage:   "Username",
				Value:   defaultUser,
				EnvVars: []string{envUser},
			},
			&cli.StringFlag{
				Name:    optPass,
				Aliases: []string{"p"},
				Usage:   "Password",
				EnvVars: []string{envPass},
			},
			&cli.PathFlag{
				Name:    optDB,
				Aliases: []string{"D"},
				Usage:   "SQLite database `path`",
				Value:   defaultDB,
				EnvVars: []string{envDB},
			},
			&cli.StringFlag{
				Name:    optLogLevel,
				Aliases: []string{"l"},
				Usage:   "Log `level`",
				Value:   zerolog.InfoLevel.String(),
				EnvVars: []string{envLogLevel},
			},
			&cli.IntFlag{
				Name:    optHour,
				Aliases: []string{"H"},
				Usage:   "The `hour` to trigger on",
				Value:   defaultHour,
				EnvVars: []string{envHour},
			},
			&cli.IntFlag{
				Name:    optMinute,
				Aliases: []string{"M"},
				Usage:   "The `minute` to trigger on",
				Value:   defaultMinute,
				EnvVars: []string{envMinute},
			},
			&cli.PathFlag{
				Name:    optConfigFile,
				Aliases: []string{"c"},
				Usage:   "Config file `path`",
				Value:   defaultConfigFile,
				EnvVars: []string{envConfigFile},
			},
		},
		Before: func(ctx *cli.Context) error {
			zerolog.TimeFieldFormat = logTimeStampLayout
			if ctx.IsSet(optLogLevel) || ctx.IsSet("l") {
				lvl, err := zerolog.ParseLevel(ctx.String(optLogLevel))
				if err != nil {
					return fmt.Errorf("%w - aborting", err)
				}
				zerolog.SetGlobalLevel(lvl)
			} else {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}
			return nil
		},
		Action: botEntryPoint,
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()
	if err := app().RunContext(ctx, os.Args); err != nil {
		_ = util.Fpf(os.Stderr, "%s\n", err.Error())
	}
}
