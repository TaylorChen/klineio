package task

import (
	"klineio/internal/repository"
	"klineio/pkg/jwt"
	"klineio/pkg/log"
	"klineio/pkg/sid"
)

type Task struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
}

func NewTask(
	tm repository.Transaction,
	logger *log.Logger,
	sid *sid.Sid,
	jwt *jwt.JWT,
) *Task {
	return &Task{
		logger: logger,
		sid:    sid,
		jwt:    jwt,
		tm:     tm,
	}
}
