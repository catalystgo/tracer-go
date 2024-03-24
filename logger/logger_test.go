package logger

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func ExampleWithLevel() {
	Logger().Desugar().WithOptions(WithLevel(zapcore.DebugLevel)).Sugar()
}

func ExampleAddKV() {
	ctx := AddKV(
		context.Background(),
		zap.String("a", "b"),
		zap.Int("c", 1),
		zap.Error(assert.AnError),
	)

	// В сообщении появятся поля, переданные в AddKV.
	Info(ctx, "some message")
}

func ExampleToContext() {
	logger := Logger()

	customLogger := logger.With(
		zap.String("component", "database"),
	)

	ctx := ToContext(context.Background(), customLogger)

	loggerFromContext := FromContext(ctx)

	// Будет содержать дополнительное поле component.
	loggerFromContext.Info("some message")
}

// func TestMetrics_Increments(t *testing.T) {
// 	cases := []struct {
// 		name    string
// 		counter prometheus.Counter
// 		act     func(logger *zap.SugaredLogger)
// 		want    float64
// 	}{
// 		{
// 			name:    "debug",
// 			counter: debugMessageCounter,
// 			act: func(logger *zap.SugaredLogger) {
// 				logger.Debug("debug")
// 			},
// 			want: 1,
// 		},
// 		{
// 			name:    "info",
// 			counter: infoMessageCounter,
// 			act: func(logger *zap.SugaredLogger) {
// 				logger.Info("info")
// 			},
// 			want: 1,
// 		},
// 		{
// 			name:    "warn",
// 			counter: warnMessageCounter,
// 			act: func(logger *zap.SugaredLogger) {
// 				logger.Warn("warn")
// 			},
// 			want: 1,
// 		},
// 		{
// 			name:    "error",
// 			counter: errorMessageCounter,
// 			act: func(logger *zap.SugaredLogger) {
// 				logger.Error("error")
// 			},
// 			want: 1,
// 		},
// 		{
// 			name:    "dpanic",
// 			counter: panicMessageCounter,
// 			act: func(logger *zap.SugaredLogger) {
// 				logger.DPanic("dpanic")
// 			},
// 			want: 1,
// 		},
// 		{
// 			name:    "panic",
// 			counter: panicMessageCounter,
// 			act: func(logger *zap.SugaredLogger) {
// 				// Panic приводит к панике и НЕ инкрементит метрику
// 				// https://github.com/uber-go/zap/blob/da406e36227eb650c4df4ab5a83dcedc00645ef2/zapcore/entry.go#L197-L198
// 				// https://github.com/uber-go/zap/blob/eae3743bc3e91db68bf977bc563263d4cb60777c/logger.go#L314-L318
// 				assert.Panics(t, func() {
// 					logger.Panic("panic") // паникует и прерывает тест
// 				})
// 			},
// 			want: 1,
// 		},
// 		{
// 			name:    "fatal",
// 			counter: fatalMessageCounter,
// 			act: func(logger *zap.SugaredLogger) {
// 				// logger.Fatal("fatal") // приводит к fatal завершению теста
// 				// Fatal приводит к панике завершению и НЕ инкрементит метрику
// 				// https://github.com/uber-go/zap/blob/da406e36227eb650c4df4ab5a83dcedc00645ef2/zapcore/entry.go#L199-L200
// 				// https://github.com/uber-go/zap/blob/eae3743bc3e91db68bf977bc563263d4cb60777c/logger.go#L314-L318
// 			},
// 			want: 0,
// 		},
// 	}

// 	for _, tc := range cases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			// arrange
// 			log := NewWithSink(zapcore.DebugLevel, io.Discard)
// 			messageCounters.Reset()

// 			// act
// 			tc.act(log)

// 			// assert
// 			require.Equal(t, tc.want, testutil.ToFloat64(tc.counter))
// 		})
// 	}
// }

type testLogLevel struct {
	level zapcore.Level
	want  bool
}

func TestKV(t *testing.T) {
	t.Parallel()

	const message = "test logger message"

	want := testNewLogEntry(
		message,
		zap.String("key-1", "val-1"),
		zap.Any("key-2", "val-2"),
	)
	kvs := []interface{}{
		zap.String("key-1", "val-1"),
		"key-2", "val-2",
	}

	cases := []struct {
		fn     func(ctx context.Context, message string, kvs ...interface{})
		levels []testLogLevel
	}{
		{
			fn: ErrorKV,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: true},
				{level: zapcore.ErrorLevel, want: true},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: InfoKV,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: false},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: DebugKV,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: false},
				{level: zapcore.WarnLevel, want: false},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: WarnKV,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: true},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		for _, level := range tc.levels {
			level := level

			t.Run(formatTestKVCase(tc.fn, level.level), func(t *testing.T) {
				t.Parallel()

				// arrange
				core, logs := observer.New(level.level)
				logger := zap.New(core).Sugar()
				ctx := ToContext(context.Background(), logger)

				// act
				tc.fn(
					ctx,
					message,
					kvs...,
				)

				// assert
				records := logs.All()
				if level.want {
					require.Len(t, records, 1)
					gotRecord := records[0]

					require.Equal(t, want.Entry.Message, gotRecord.Message)
					require.Equal(t, want.Context, gotRecord.Context)
				} else {
					require.Len(t, records, 0)
				}
			})
		}
	}
}

func TestFormat(t *testing.T) {
	t.Parallel()

	messageTemplate := "string: %s, int: %d"
	messageArgs := []interface{}{"string", 1}
	wantMessage := fmt.Sprintf(messageTemplate, messageArgs...)

	want := testNewLogEntry(wantMessage)

	cases := []struct {
		fn     func(ctx context.Context, format string, args ...interface{})
		levels []testLogLevel
	}{
		{
			fn: Errorf,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: true},
				{level: zapcore.ErrorLevel, want: true},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: Infof,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: false},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: Debugf,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: false},
				{level: zapcore.WarnLevel, want: false},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: Warnf,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: true},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		for _, level := range tc.levels {
			level := level

			t.Run(formatTestKVCase(tc.fn, level.level), func(t *testing.T) {
				t.Parallel()

				// arrange
				core, logs := observer.New(level.level)
				logger := zap.New(core).Sugar()
				ctx := ToContext(context.Background(), logger)

				// act
				tc.fn(
					ctx,
					messageTemplate,
					messageArgs...,
				)

				// assert
				records := logs.All()
				if level.want {
					require.Len(t, records, 1)
					gotRecord := records[0]

					require.Equal(t, want.Entry.Message, gotRecord.Message)
				} else {
					require.Len(t, records, 0)
				}
			})
		}
	}
}

func TestSimple(t *testing.T) {
	t.Parallel()

	message := "test simple log message"

	want := testNewLogEntry(message)

	cases := []struct {
		fn     func(ctx context.Context, args ...interface{})
		levels []testLogLevel
	}{
		{
			fn: Error,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: true},
				{level: zapcore.ErrorLevel, want: true},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: Info,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: false},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: Debug,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: false},
				{level: zapcore.WarnLevel, want: false},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
		{
			fn: Warn,
			levels: []testLogLevel{
				{level: zapcore.DebugLevel, want: true},
				{level: zapcore.InfoLevel, want: true},
				{level: zapcore.WarnLevel, want: true},
				{level: zapcore.ErrorLevel, want: false},
				{level: zapcore.PanicLevel, want: false},
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		for _, level := range tc.levels {
			level := level

			t.Run(formatTestKVCase(tc.fn, level.level), func(t *testing.T) {
				t.Parallel()

				// arrange
				core, logs := observer.New(level.level)
				logger := zap.New(core).Sugar()
				ctx := ToContext(context.Background(), logger)

				// act
				tc.fn(ctx, message)

				// assert
				records := logs.All()
				if level.want {
					require.Len(t, records, 1)
					gotRecord := records[0]

					require.Equal(t, want.Entry.Message, gotRecord.Message)
				} else {
					require.Len(t, records, 0)
				}
			})
		}
	}
}

func formatTestKVCase(fn interface{}, level zapcore.Level) string {
	return fmt.Sprintf("fn:%s_loglevel:%s", getFunctionName(fn), level.String())
}

func getFunctionName(i interface{}) string {
	fullname := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	return fullname[strings.LastIndex(fullname, "/")+1:]
}
