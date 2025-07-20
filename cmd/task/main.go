package main

import (
	"context"
	"fmt"
	"klineio/cmd/task/wire"
	"klineio/pkg/config"
	"klineio/pkg/log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
)

var (
	cfg = pflag.StringP("config", "c", "config/local.yml", "config file path.")
)

func main() {
	pflag.Parse()

	// Check for -conf parameter manually
	configPath := *cfg
	for i, arg := range os.Args {
		if arg == "-conf" && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
			break
		}
	}

	fmt.Printf("final config path: %s\n", configPath)
	conf := config.NewConfig(configPath)
	logger := log.NewLog(conf)
	logger.Info("start task")

	// Create a context that can be cancelled by OS signals
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app, cleanup, err := wire.NewWire(conf, logger)
	defer cleanup()

	if err != nil {
		panic(err)
	}

	// Pass the cancellable context to app.Run
	if err = app.Run(ctx); err != nil {
		panic(err)
	}
}
