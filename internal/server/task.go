package server

import (
	"context"
	"klineio/internal/job"
	"klineio/internal/task"
	"klineio/pkg/log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type TaskServer struct {
	log             *log.Logger
	conf            *viper.Viper // Add conf field
	scheduler       *gocron.Scheduler
	userTask        task.UserTask
	priceMonitorJob *job.PriceMonitorJob // Add PriceMonitorJob
}

func NewTaskServer(
	log *log.Logger,
	conf *viper.Viper, // Add conf parameter
	userTask task.UserTask,
	priceMonitorJob *job.PriceMonitorJob, // Add priceMonitorJob as a parameter
) *TaskServer {
	return &TaskServer{
		log:             log,
		conf:            conf, // Assign conf
		userTask:        userTask,
		priceMonitorJob: priceMonitorJob, // Assign priceMonitorJob
	}
}
func (t *TaskServer) Start(ctx context.Context) error {
	gocron.SetPanicHandler(func(jobName string, recoverData interface{}) {
		t.log.Error("TaskServer Panic", zap.String("job", jobName), zap.Any("recover", recoverData))
	})

	// Initialize a new scheduler
	s := gocron.NewScheduler(time.UTC)

	// Use the configured timeout for the context passed to jobs
	jobTimeout := time.Duration(t.conf.GetInt("price_monitor.timeout_seconds")) * time.Second

	// Schedule the PriceMonitorJob to run every 5 minutes
	_, err := s.Every(5).Minutes().Do(func() {
		jobCtx, cancel := context.WithTimeout(context.Background(), jobTimeout)
		defer cancel()

		err := t.priceMonitorJob.Run(jobCtx)
		if err != nil {
			t.log.WithContext(jobCtx).Error("PriceMonitorJob error", zap.Error(err))
		}
	})

	if err != nil {
		return err
	}

	// Start the scheduler asynchronously
	s.StartAsync()

	// Keep the server running until context is cancelled
	<-ctx.Done()
	s.Stop()
	return nil
}
func (t *TaskServer) Stop(ctx context.Context) error {
	t.scheduler.Stop()
	t.log.Info("TaskServer stop...")
	return nil
}
