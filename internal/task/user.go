package task

import (
	"klineio/internal/repository"
	"klineio/pkg/log"
)

type UserTask interface {
	// CheckUser(ctx context.Context) error // Removed CheckUser interface method
}

type userTask struct {
	userRepo repository.UserRepository
	logger   *log.Logger
}

func NewUserTask(
	userRepo repository.UserRepository,
	logger *log.Logger,
) UserTask {
	return &userTask{
		userRepo: userRepo,
		logger:   logger,
	}
}

// func (t userTask) CheckUser(ctx context.Context) error {
//	// do something
//	t.logger.Info("CheckUser")
//	return nil
// }
