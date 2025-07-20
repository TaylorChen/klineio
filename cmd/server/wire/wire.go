//go:build wireinject
// +build wireinject

package wire

import (
	"klineio/internal/handler"
	"klineio/internal/job"
	"klineio/internal/repository"
	"klineio/internal/server"
	"klineio/internal/service"
	"klineio/pkg/app"
	"klineio/pkg/exchange"
	"klineio/pkg/jwt"
	"klineio/pkg/log"
	"klineio/pkg/notifier"
	"klineio/pkg/server/http"
	"klineio/pkg/sid"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

// provideDingTalkWebhookURL provides the DingTalk webhook URL from config.
func provideDingTalkWebhookURL(conf *viper.Viper) string {
	return conf.GetString("dingtalk.webhook_url")
}

var repositorySet = wire.NewSet(
	repository.NewDB,
	repository.NewRedis,
	repository.NewMongo,
	repository.NewRepository,
	repository.NewTransaction,
	repository.NewUserRepository,
	repository.NewExchangePriceRepository,
	// repository.NewMonitorConfigRepository, // 暂时注释掉，因为该函数不存在
)

var serviceSet = wire.NewSet(
	service.NewService,
	service.NewUserService,
	service.NewPriceMonitorService,
)

var handlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewUserHandler,
)

var jobSet = wire.NewSet(
	job.NewJob,
	job.NewUserJob,
	job.NewPriceMonitorJob,
)

var exchangeClientSet = wire.NewSet(
	exchange.NewBinanceClient,
	exchange.NewOKEXClient,
)

var notifierSet = wire.NewSet(
	notifier.NewDingTalkNotifier,
)

var serverSet = wire.NewSet(
	server.NewHTTPServer,
	server.NewJobServer,
)

// build App
func newApp(
	httpServer *http.Server,
	jobServer *server.JobServer,
	// task *server.Task,
) *app.App {
	return app.NewApp(
		app.WithServer(httpServer, jobServer),
		app.WithName("klineio-server"), // Modified app name
	)
}

func NewWire(conf *viper.Viper, logger *log.Logger) (*app.App, func(), error) {
	panic(wire.Build(
		repositorySet,
		serviceSet,
		handlerSet,
		jobSet,
		exchangeClientSet,
		notifierSet,
		serverSet,
		newApp,
		jwt.NewJwt,
		sid.NewSid,
		provideDingTalkWebhookURL,
	))
}
