package logger

import (
	"context"

	"go.uber.org/zap"
)

const (
	loggerKey = "logger"

	RequestIDKey = "request_id"
	MethodKey    = "method"
	PathKey      = "path"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	WithFields(fields ...zap.Field) Logger
}

type ZapLogger struct {
	*zap.Logger
}

func NewLogger() Logger {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	return &ZapLogger{
		Logger: zapLogger,
	}
}

func (l *ZapLogger) WithFields(fields ...zap.Field) Logger {
	return &ZapLogger{Logger: l.Logger.With(fields...)}
}

func LoggerWithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return NewLogger()
}
