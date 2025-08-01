// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wire

import (
	"github.com/spf13/viper"
	"klineio/internal/repository"
	"klineio/internal/server"
	"klineio/pkg/app"
	"klineio/pkg/log"
)

// Injectors from wire.go:

func NewWire(conf *viper.Viper, logger *log.Logger) (*app.App, func(), error) {
	db := repository.NewDB(conf, logger)
	migrateServer := server.NewMigrateServer(db, logger)
	appApp := newApp(migrateServer)
	return appApp, func() {
	}, nil
}

// wire.go:

func newApp(
	migrate *server.MigrateServer,
) *app.App {
	return app.NewApp(app.WithServer(migrate), app.WithName("klineio-migration"))
}
