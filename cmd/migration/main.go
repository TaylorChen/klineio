package main

import (
	"klineio/internal/model"
	"klineio/internal/repository"
	"klineio/pkg/config"
	"klineio/pkg/log"
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
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

	conf := config.NewConfig(configPath)
	logger := log.NewLog(conf)

	// Initialize DB with a GORM logger that outputs info level (including SQL)
	db := repository.NewDB(conf, logger) // NewDB already uses our custom zapgorm2 logger

	// Set GORM logger level to Info to see auto-migration SQL statements
	// db.Logger = db.Logger.LogMode(gorm.Info) // Set LogMode to Info to see SQL

	err := db.AutoMigrate(&model.ExchangePrice{}, &model.MonitorConfig{}) // AutoMigrate the models
	if err != nil {
		logger.Fatal("failed to auto migrate database", zap.Error(err))
	}

	logger.Info("Database migration completed successfully!")
}
