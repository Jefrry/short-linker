package logger

import (
	"time"

	"go.uber.org/zap"
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	Sync() error
}

// Field is a wrapper around zap.Field to avoid direct dependency on zap in other packages.
type Field struct {
	zapField zap.Field
}

// zapFields converts internal Fields to zap.Fields.
func zapFields(fields []Field) []zap.Field {
	zf := make([]zap.Field, len(fields))
	for i, f := range fields {
		zf[i] = f.zapField
	}
	return zf
}

type ZapLogger struct {
	logger *zap.Logger
}

func NewLogger(l *zap.Logger) *ZapLogger {
	return &ZapLogger{logger: l}
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, zapFields(fields)...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, zapFields(fields)...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, zapFields(fields)...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, zapFields(fields)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, zapFields(fields)...)
}

func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{logger: l.logger.With(zapFields(fields)...)}
}

func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

func NewProduction() (*zap.Logger, error) {
	return zap.NewProduction()
}

// Field helpers
func String(key, value string) Field {
	return Field{zapField: zap.String(key, value)}
}

func Int(key string, value int) Field {
	return Field{zapField: zap.Int(key, value)}
}

func Duration(key string, value time.Duration) Field {
	return Field{zapField: zap.Duration(key, value)}
}

func Error(err error) Field {
	return Field{zapField: zap.Error(err)}
}

func Any(key string, value any) Field {
	return Field{zapField: zap.Any(key, value)}
}
