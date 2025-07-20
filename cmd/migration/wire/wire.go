//go:build wireinject
// +build wireinject

package wire

import (
	"klineio/internal/repository"
	"klineio/internal/server"
	"klineio/pkg/app"
	"klineio/pkg/log"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

func newApp(
	migrate *server.MigrateServer,
) *app.App {
	return app.NewApp(
		app.WithServer(migrate),
		app.WithName("klineio-migration"),
	)
}

func NewWire(conf *viper.Viper, logger *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repository.NewDB,
		newApp,
		server.NewMigrateServer,
	))
}
