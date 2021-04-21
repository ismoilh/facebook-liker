package main

import (
	"context"
	"facebook-liker/internal/scrapper"
	"facebook-liker/internal/server"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	// Version of the server
	Version = "0.1"
)

// addr is custom command flag for holding host:port address
type addr struct {
	host net.IP
	port string
}

var (
	defaultHost = net.ParseIP("127.0.0.1")
	defaultPort = "7070"
)

// Set sets address to s
func (a *addr) Set(s string) error {
	ss := strings.Split(s, ":")
	if len(ss) == 1 { // s doesn't contain :, so it's just port
		// if just port is specified, then listen on localhost + port
		a.host = defaultHost
		a.port = ss[0]
	} else if len(ss) == 2 { // s is in form host:port
		// if host is empty, then listen on localhost + port
		if ss[0] == "" {
			a.host = defaultHost
			a.port = ss[1]
		} else { // else use host and port from s
			// check if user supplied correct ip
			a.host = net.ParseIP(ss[0])
			if a.host == nil {
				return errors.Errorf("invalid host format: %s", ss[0])
			}
			a.port = ss[1]
		}
	} else {
		return errors.Errorf("invalid address format")
	}

	return nil
}

// Type returns name of the address type
func (a *addr) Type() string {
	return "addr"
}

// String returns string representation of address
func (a *addr) String() string {
	// default value when address is not specified
	if a == nil || a.host.String() == "<nil>" {
		return defaultHost.String() + ":" + defaultPort
	}

	return a.host.String() + ":" + a.port
}

// Config holds server command flags and environmental variables,
// that is used for configuring server
type Config struct {
	// Addr is an address on which server will listen on
	Addr *addr

	Selenium *struct {
		// Addr of selenium hub on which it listens
		Addr string

		// BrowserName ...
		BrowserName string

		// BrowserArgs are extra args for browser, like --headless
		BrowserArgs []string

		// MaxInstances are maximum amount of parallel instances of same browser
		// https://stackoverflow.com/questions/13723349/selenium-grid-maxsessions-vs-maxinstances
		MaxInstances int

		// MaxSessions are maximum amount of paralles instances for all browsers
		// https://stackoverflow.com/questions/13723349/selenium-grid-maxsessions-vs-maxinstances
		MaxSessions int
	}

	// Logs is a struct that holds configuration for logging
	Logs *struct {
		// Local path of a filename for storing logs
		Path string

		// Logging level:
		// 0 for DEBUG, 1 for INFO, 2 for WARNING, 3 for ERROR, 4 for FATAL, 5 for PANIC
		Level int

		// file for storing logs
		file *os.File
	}
}

var conf *Config

// parse all command flags
func init() {
	conf = &Config{
		Addr: &addr{},
		Selenium: &struct {
			Addr string
			BrowserName string
			BrowserArgs []string
			MaxInstances int
			MaxSessions int
		}{},
		Logs: &struct {
			Path  string
			Level int
			file  *os.File
		}{},
	}

	rootCmd.PersistentFlags().VarP(conf.Addr, "addr", "a", `address on which server should listen, examples: "127.0.0.1:7070", ":7070", "7070"`)

	rootCmd.PersistentFlags().StringVarP(&conf.Selenium.Addr, "selenium-addr", "s", "selenium-hub:4444", `address of selenium hub, examples: "127.0.0.1:4444", ":4444", "4444"`)

	rootCmd.PersistentFlags().StringVarP(&conf.Selenium.BrowserName, "selenium-browser", "b", "chrome", "browser name for scrapping")

	rootCmd.PersistentFlags().StringSliceVar(&conf.Selenium.BrowserArgs, "selenium-args", []string{"--headless"}, "extra args for browser")

	rootCmd.PersistentFlags().IntVar(&conf.Selenium.MaxInstances, "selenium-max-instances", 10, "maximum instances of same browser to run in parallel: https://stackoverflow.com/questions/13723349/selenium-grid-maxsessions-vs-maxinstances")

	rootCmd.PersistentFlags().IntVar(&conf.Selenium.MaxSessions, "selenium-max-sessions", 10, "maximum sessions of all browsers to run in parallel: https://stackoverflow.com/questions/13723349/selenium-grid-maxsessions-vs-maxinstances")

	rootCmd.PersistentFlags().StringVarP(&conf.Logs.Path, "logs-path", "p", "logs.jsonl", "local path of a filename for storing logs, it will be created, if it doesn't exist")

	rootCmd.PersistentFlags().IntVarP(&conf.Logs.Level, "logs-level", "l", 1, "logging level: 0 for DEBUG, 1 for INFO, 2 for WARNING, 3 for ERROR, 4 for FATAL, 5 for PANIC")
}

// rootCmd is main server command
var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "A Golang server that uses Selenium for liking Facebook posts",
	Long:  `A Golang server that uses Selenium for liking Facebook posts`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		defer errors.Wrap(err, "main.rootCmd.RunE")

		// close log file on program exit
		defer func() {
			if cerr := conf.Logs.file.Close(); cerr != nil {
				if err != nil {
					err = errors.WithMessage(cerr, "failed to close log file")
				}
			}
		}()

		// init scrapper config
		scrp, err := scrapper.New(
			conf.Selenium.Addr,
			scrapper.OptBrowserName(conf.Selenium.BrowserName),
			scrapper.OptBrowserArgs(conf.Selenium.BrowserArgs),
			scrapper.OptMaxInstances(conf.Selenium.MaxInstances),
			scrapper.OptMaxSessions(conf.Selenium.MaxSessions),
		)

		// init server
		srv, err := server.New(conf.Addr.String(), scrp)
		if err != nil {
			return errors.WithMessage(err, "failed to construct server")
		}

		// start the server
		log.Logger.Info().Str("addr", conf.Addr.String()).Msg("STARTING the server")
		srvErrChan := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())
		go srv.Run(ctx, srvErrChan)
		log.Logger.Info().Str("addr", conf.Addr.String()).Msg("SUCCESFULLY started the server")

		// catch Ctrl + C (SIGINT) signal
		sigintChan := make(chan os.Signal, 1)
		signal.Notify(sigintChan, syscall.SIGINT)

		log.Logger.Info().Msg("WAITING for SIGINT(Ctrl + C) signal")
		select {
		case <-sigintChan:
			log.Logger.Info().Msg("SUCCESFULLY caught SIGINT(Ctrl + C) signal")
			cancel()
			break
		case err := <-srvErrChan:
			log.Logger.Warn().Msg("RECEIVED error from server")
			cancel()
			return err
		}

		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		defer errors.Wrap(err, "main.rootCmd.PreRunE")

		// create new file for storing logs
		f, err := os.Create(conf.Logs.Path)
		if err != nil {
			return errors.WithMessage(err, "failed to create log file")
		}

		conf.Logs.file = f

		// output logs both in terminal and file
		// and set beautiful logging for terminals
		multi := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.Stamp,
		}, f)

		log.Logger = log.Output(multi).With().Logger()

		// set verbosity level
		zerolog.SetGlobalLevel(zerolog.Level(conf.Logs.Level))

		return nil
	},
	Version: Version,
	Example: "server --addr localhost:7070 --logs-path logs.jsonl --logs-level 3",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Logger.Fatal().Msgf("%v\n", err)
	}
}
