package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/getoutreach/gobox/pkg/app"
	"github.com/pkg/errors"
	jsonnetplayground "github.com/rgst-io/jsonnet-playground/internal/jsonnet-playground"
	"github.com/sirupsen/logrus"
	"github.com/tritonmedia/pkg/service"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := logrus.New().WithContext(ctx)

	//nolint:gocritic importShadow
	app := cli.App{
		Name:    "jsonnet-playground",
		Version: app.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-format",
				Usage:   "set the formatter for logs",
				EnvVars: []string{"LOG_FORMAT"},
				Value:   "JSON",
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		logFormat := c.String("log-format")
		switch strings.ToLower(logFormat) {
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{})
		case "text":
			logrus.SetFormatter(&logrus.TextFormatter{})
		default:
			return fmt.Errorf("unknown log format: %s", logFormat)
		}

		// update our logger's formatter, since we created before
		// we set it
		log.Logger.Formatter = logrus.StandardLogger().Formatter

		// load the config
		conf, err := jsonnetplayground.LoadConfig()
		if err != nil {
			return errors.Wrap(err, "failed to load config")
		}

		// start the service runner, which handles context cancellation
		// and threading
		r := service.NewServiceRunner(ctx, []service.Service{jsonnetplayground.NewPublicHTTPService(conf)})

		sigC := make(chan os.Signal)

		// listen for signals that we want to cancel on, and cancel
		// the context if one is passed
		signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
		go func() {
			sig := <-sigC
			log.WithField("signal", sig.String()).Info("got signal, shutting down")
			cancel()
		}()

		return r.Run(ctx, log)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}
