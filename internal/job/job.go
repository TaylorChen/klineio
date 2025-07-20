package job

import (
	"klineio/internal/repository"
	"klineio/pkg/jwt"
	"klineio/pkg/log"
	"klineio/pkg/sid"
)

type Job struct {
	logger       *log.Logger
	sid          *sid.Sid
	jwt          *jwt.JWT
	tm           repository.Transaction
	PriceMonitor *PriceMonitorJob // Add PriceMonitorJob
}

func NewJob(
	tm repository.Transaction,
	logger *log.Logger,
	sid *sid.Sid,
	priceMonitorJob *PriceMonitorJob, // Add priceMonitorJob as a parameter
) *Job {
	return &Job{
		logger:       logger,
		sid:          sid,
		tm:           tm,
		PriceMonitor: priceMonitorJob, // Assign priceMonitorJob
	}
}
