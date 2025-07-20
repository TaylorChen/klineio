//go:build wireinject
// +build wireinject

package wire

import (
	"klineio/internal/job"
	"klineio/internal/repository"
	"klineio/internal/server"
	"klineio/internal/service"
	"klineio/internal/task"
	"klineio/pkg/app"
	"klineio/pkg/exchange"
	"klineio/pkg/log"
	"klineio/pkg/notifier"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

// provideDingTalkWebhookURL provides the DingTalk webhook URL from config.
func provideDingTalkWebhookURL(conf *viper.Viper) string {
	return conf.GetString("dingtalk.webhook_url")
}

var repositorySet = wire.NewSet(
	repository.NewDB,
	//repository.NewRedis,
	repository.NewRepository,
	repository.NewTransaction,
	repository.NewUserRepository,
	repository.NewExchangePriceRepository,
)

var exchangeClientSet = wire.NewSet(
	exchange.NewBinanceClient,
	exchange.NewOKEXClient,
)

var notifierSet = wire.NewSet(
	notifier.NewDingTalkNotifier,
)

var serviceSet = wire.NewSet(
	service.NewPriceMonitorService,
)

var taskSet = wire.NewSet(
	task.NewTask,
	task.NewUserTask,
	job.NewPriceMonitorJob,
)
var serverSet = wire.NewSet(
	server.NewTaskServer,
)

// build App
func newApp(
	task *server.TaskServer,
) *app.App {
	return app.NewApp(
		app.WithServer(task),
		app.WithName("klineio"), // Modified app name
	)
}

func NewWire(conf *viper.Viper, logger *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		exchangeClientSet,
		notifierSet,
		serviceSet,
		taskSet,
		serverSet,
		// job.NewJob, // Removed unused provider again
		newApp,
		provideDingTalkWebhookURL,
	))
}
