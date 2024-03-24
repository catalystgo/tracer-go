package logger

import (
	"context"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	global       *zap.SugaredLogger
	defaultLevel = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
)

// init Set the default logger value
func init() {
	SetLogger(New(defaultLevel.Level()))
}

// New create a new logger with specific log_leve & options
func New(level zapcore.Level, options ...zap.Option) *zap.SugaredLogger {
	return NewWithSink(level, os.Stdout, options...)
}

func NewWithSink(level zapcore.LevelEnabler, sink io.Writer, options ...zap.Option) *zap.SugaredLogger {
	if level == nil {
		level = defaultLevel
	}

	core := newZapCore(level, sink)
	return zap.New(core, options...).Sugar()
}

func newZapCore(level zapcore.LevelEnabler, sink io.Writer) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "ts",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    "function  ",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.AddSync(sink),
		level,
	)
}

// Level Get current log_level
func Level() zapcore.Level {
	return defaultLevel.Level()
}

// SetLevel Set current log_level
func SetLevel(l zapcore.Level) {
	defaultLevel.SetLevel(l)
}

// Logger Get global logger
func Logger() *zap.SugaredLogger {
	return global
}

// SetLogger Set global logger (not thread safe)
func SetLogger(l *zap.SugaredLogger) {
	global = l
}

func Debug(ctx context.Context, args ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.DebugLevel) {
		l.Debug(args...)
	}
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.DebugLevel) {
		l.Debugf(format, args...)
	}
}

func DebugKV(ctx context.Context, message string, kvs ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.DebugLevel) {
		l.Debugw(message, mergeKvs(ctx, kvs...)...)
	}
}

func Info(ctx context.Context, args ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.InfoLevel) {
		l.Info(args...)
	}
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.InfoLevel) {
		l.Infof(format, args...)
	}
}

func InfoKV(ctx context.Context, message string, kvs ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.InfoLevel) {
		l.Infow(message, mergeKvs(ctx, kvs...)...)
	}
}

func Warn(ctx context.Context, args ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.WarnLevel) {
		l.Warn(args...)
	}
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.WarnLevel) {
		l.Warnf(format, args...)
	}
}

func WarnKV(ctx context.Context, message string, kvs ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.WarnLevel) {
		l.Warnw(message, mergeKvs(ctx, kvs...)...)
	}
}

func Error(ctx context.Context, args ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.ErrorLevel) {
		l.Error(args...)
	}
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.ErrorLevel) {
		l.Errorf(format, args...)
	}
}

func ErrorKV(ctx context.Context, message string, kvs ...interface{}) {
	if l := FromContext(ctx); l.Level().Enabled(zapcore.ErrorLevel) {
		l.Errorw(message, mergeKvs(ctx, kvs...)...)
	}
}

func Fatal(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Fatal(args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Fatalf(format, args)
}

func FatalKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Fatalw(message, mergeKvs(ctx, kvs...)...)
}

func Panic(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Panic(args...)
}

func Panicf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Panicf(format, args...)
}

func PanicKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Panicw(message, mergeKvs(ctx, kvs...)...)
}
