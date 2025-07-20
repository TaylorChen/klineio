package zapgorm2

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	pkglog "klineio/pkg/log"
)

type gormLogger struct {
	pkgLogger                 *pkglog.Logger // Change to our custom logger type
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

// NewGormLogger creates a new GORM Logger compatible with klineio/pkg/log.
// It renames the constructor to avoid conflict with existing New function in the package.
func NewGormLogger(pkgLogger *pkglog.Logger) *gormLogger {
	return &gormLogger{
		pkgLogger:                 pkgLogger,
		LogLevel:                  logger.Warn, // Default log level
		SlowThreshold:             time.Second, // Default slow threshold
		IgnoreRecordNotFoundError: true,
	}
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.pkgLogger.WithContext(ctx).Logger.Info(fmt.Sprintf(msg, data...))
	}
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.pkgLogger.WithContext(ctx).Logger.Warn(fmt.Sprintf(msg, data...))
	}
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.pkgLogger.WithContext(ctx).Logger.Error(fmt.Sprintf(msg, data...))
	}
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		// zap.String("line", utils.FileWithLine()), // Temporarily comment out due to import issue
	}

	if err != nil {
		if l.IgnoreRecordNotFoundError && err == gorm.ErrRecordNotFound {
			l.pkgLogger.WithContext(ctx).Logger.Debug("record not found", fields...)
			return
		}
		if l.LogLevel >= logger.Error {
			fields = append(fields, zap.Error(err))
			l.pkgLogger.WithContext(ctx).Logger.Error("trace", fields...)
		}
	} else if elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn {
		l.pkgLogger.WithContext(ctx).Logger.Warn("trace", fields...)
	} else if l.LogLevel >= logger.Info {
		l.pkgLogger.WithContext(ctx).Logger.Info("trace", fields...)
	}
}
