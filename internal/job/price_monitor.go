package job

import (
	"context"
	"time"

	"klineio/internal/service"
	"klineio/pkg/log"

	"go.uber.org/zap"
)

// PriceMonitorJob defines the job for monitoring cryptocurrency prices.
type PriceMonitorJob struct {
	priceMonitorSvc *service.PriceMonitorService
	logger          *log.Logger
	interval        time.Duration
}

// NewPriceMonitorJob creates a new PriceMonitorJob.
func NewPriceMonitorJob(
	priceMonitorSvc *service.PriceMonitorService,
	logger *log.Logger,
) *PriceMonitorJob {
	return &PriceMonitorJob{
		priceMonitorSvc: priceMonitorSvc,
		logger:          logger,
		interval:        1 * time.Minute, // Default to 1 minute, can be configured
	}
}

// Run executes the price monitoring job once.
func (j *PriceMonitorJob) Run(ctx context.Context) error {
	j.logger.Info("Running PriceMonitorJob once")
	err := j.priceMonitorSvc.RunMonitor(ctx)
	if err != nil {
		j.logger.Error("Error running price monitor service", zap.Error(err))
		return err
	}
	return nil
}
