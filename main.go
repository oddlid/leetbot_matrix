package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

const (
	logTimeStampLayout     = `2006-01-02T15:04:05.999-07:00`
	appName                = `leetbot_matrix`
	defaultHomeServer      = `oddware.net`
	defaultUser            = `leetbot`
	defaultDB              = `leetbot_matrix.db`
	defaultScoreFile       = `/tmp/leetbot_scores.json`
	defaultBonusConfigFile = `/tmp/leetbot_bonusconfigs.json`
	defaultHour            = 13
	defaultMinute          = 37
	defaultNTPServer       = `0.se.pool.ntp.org`
	envServer              = `M_HOMESERVER`
	envUser                = `M_USER`
	envPass                = `M_PASS`
	envDB                  = `M_DB`
	envLogLevel            = `L_LOGLEVEL`
	envHour                = `L_HOUR`
	envMinute              = `L_MINUTE`
	envScoreFile           = `L_SCOREFILE`
	envBonusConfigFile     = `L_BONUSCONFIGFILE`
	envNTPServer           = `L_NTP_SERVER`
	optServer              = `server`
	optRoom                = `room` // can we use this, or must the bot be invited always?
	optUser                = `user`
	optPass                = `pass`
	optDB                  = `db`
	optLogLevel            = `log-level`
	optHour                = `hour`
	optMinute              = `minute`
	optScoreFile           = `score-file`
	optBonusConfigFile     = `bonus-config-file`
	optNTPServer           = `ntp-server`
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
		Usage: "Run 1337 bot for Matrix",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    optServer,
				Aliases: []string{"s"},
				Usage:   "Matrix homeserver `address` to connect to",
				Value:   defaultHomeServer,
				EnvVars: []string{envServer},
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
				Usage:   "The `hour` to trigger on",
				Value:   defaultHour,
				EnvVars: []string{envHour},
			},
			&cli.IntFlag{
				Name:    optMinute,
				Usage:   "The `minute` to trigger on",
				Value:   defaultMinute,
				EnvVars: []string{envMinute},
			},
			&cli.PathFlag{
				Name:    optScoreFile,
				Aliases: []string{"f"},
				Usage:   "Score file `path`",
				Value:   defaultScoreFile,
				EnvVars: []string{envScoreFile},
			},
			&cli.PathFlag{
				Name:    optBonusConfigFile,
				Aliases: []string{"b"},
				Usage:   "Bonus config file `path`",
				Value:   defaultBonusConfigFile,
				EnvVars: []string{envBonusConfigFile},
			},
			&cli.StringFlag{
				Name:    optNTPServer,
				Aliases: []string{"n"},
				Usage:   "NTP server `address`",
				Value:   defaultNTPServer,
				EnvVars: []string{envNTPServer},
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
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
