package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestFromContextGlobalLogger(t *testing.T) {
	logger := FromContext(context.Background())

	require.Equal(t, global, logger)
}

func TestFromContextWithLogger(t *testing.T) {
	l := New(zapcore.DebugLevel)
	ctx := ToContext(context.Background(), l)

	require.Equal(t, l, FromContext(ctx))
}

func TestLevelFromContextGlobalLogger(t *testing.T) {
	lvl := LevelFromContext(context.Background())

	require.Equal(t, global.Level(), lvl)
}

func TestLevelFromContextWithLogger(t *testing.T) {
	l := New(zapcore.DPanicLevel)
	ctx := ToContext(context.Background(), l)

	lvl := LevelFromContext(ctx)
	require.Equal(t, zapcore.DPanicLevel, lvl)
}

func TestLoggerWithSpanContext(t *testing.T) {
	buf := bytes.Buffer{}

	tID, err := trace.TraceIDFromHex("55e02c160e0dbd1b441bf1d5dc3ea3d5")
	require.NoError(t, err)
	sID, err := trace.SpanIDFromHex("a48b167265f65931")
	require.NoError(t, err)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: tID,
		SpanID:  sID,
	})

	l := loggerWithSpanContext(loggerWithWriter(&buf), sc)
	l.Debug("hello world")

	var decoded map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}

	require.Equal(t, "55e02c160e0dbd1b441bf1d5dc3ea3d5", decoded["trace_id"])
	require.Equal(t, "a48b167265f65931", decoded["span_id"])
	require.Equal(t, "hello world", decoded["message"])
}

func TestLoggerWithName(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	l := loggerWithWriter(&buf)

	ctx := ToContext(context.Background(), l)
	ctx = WithName(ctx, "test-logger")

	FromContext(ctx).Debug(ctx, "hello world")

	t.Logf("%s", &buf)
	var decoded map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}

	require.Equal(t, "test-logger", decoded["logger"])
}

func TestLoggerWithKV(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	l := loggerWithWriter(&buf)

	ctx := ToContext(context.Background(), l)
	ctx = WithKV(ctx, "apples", 500)

	FromContext(ctx).Debug(ctx, "hello world")

	t.Logf("%s", &buf)
	var decoded map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}

	require.EqualValues(t, 500, decoded["apples"])
}

func TestLoggerWithFields(t *testing.T) {
	t.Parallel()

	buf := bytes.Buffer{}
	l := loggerWithWriter(&buf)

	ctx := ToContext(context.Background(), l)
	ctx = WithFields(ctx,
		zap.String("kafka-topic", "test-topic"),
		zap.Int32("kafka-partition", 420),
	)

	FromContext(ctx).Debug(ctx, "hello world")

	t.Logf("%s", &buf)
	var decoded map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}

	require.EqualValues(t, "test-topic", decoded["kafka-topic"])
	require.EqualValues(t, 420, decoded["kafka-partition"])
}

func loggerWithWriter(w io.Writer) *zap.SugaredLogger {
	sink := zapcore.AddSync(w)
	return zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				TimeKey:        "ts",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "message",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}),
			sink,
			zap.NewAtomicLevelAt(zapcore.DebugLevel),
		),
	).Sugar()
}

func ExampleWithKV() {
	// context, containing logger
	ctx := context.Background()
	// new context with logger fields
	ctx = WithKV(ctx, "my key", "my value")

	_ = ctx
}

func ExampleWithFields() {
	// context, containing logger
	ctx := context.Background()
	// new context with logger fields
	ctx = WithFields(ctx,
		zap.String("kafka-topic", "my topic"),
		zap.Int32("kafka-partition", 1),
	)

	_ = ctx
}

func ExampleWithName() {
	// context, containing logger
	ctx := context.Background()
	ctx = WithName(ctx, "GetApples")    // -> "GetApples"
	ctx = WithName(ctx, "AppleManager") // - > "GetApples.AppleManager"
	ctx = WithName(ctx, "DB")           // -> "GetApples.AppleManager.DB"

	_ = ctx
}

func TestAddKVAndFieldsFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		name string
		kvs  []any
		want []any
		ctx  context.Context
	}{
		{
			name: "empty values list",
			want: nil,
			ctx:  ctx,
		},
		{
			name: "add to empty context",
			kvs: []any{
				"int", 1,
				"string", "2",
				"bool", true,
			},
			want: []any{
				zap.Any("int", 1),
				zap.Any("string", "2"),
				zap.Any("bool", true),
			},
			ctx: ctx,
		},
		{
			name: "add odd number of arguments to empty context",
			kvs: []any{
				"int", 1,
				"string", "2",
				"bool", // No value for `bool` key therfor skipped.
			},
			want: []any{
				zap.Any("int", 1),
				zap.Any("string", "2"),
			},
			ctx: ctx,
		},
		{
			name: "add invalid key type to empty context",
			kvs: []any{
				"int", 1,
				2, "2", // Will be skipped, since it's the key is not a string.
				"bool", true,
			},
			want: []any{
				zap.Any("int", 1),
				zap.Any("bool", true),
			},
			ctx: ctx,
		},
		{
			name: "add to already added keys",
			kvs: []any{
				"string", "2",
				"bool", true,
			},
			want: []any{
				zap.Any("int", 1),
				zap.Any("string", "2"),
				zap.Any("bool", true),
			},
			ctx: AddKV(ctx, "int", 1),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// act
			got := AddKV(tc.ctx, tc.kvs...)

			// require
			require.Equal(t, tc.want, getKvsFromContext(got))
		})
	}
}

func TestMergeKVs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		name string
		ctx  context.Context
		kvs  []any
		want []any
	}{
		{
			name: "empty context, empty kvs",
			ctx:  ctx,
			kvs:  nil,
			want: nil,
		},
		{
			name: "empty context, some kvs",
			ctx:  ctx,
			kvs:  []any{"firstkey", "firstvalue"},
			want: []any{
				zap.Any("firstkey", "firstvalue"),
			},
		},
		{
			name: "user defined field, empty kvs",
			ctx:  AddKV(ctx, "firstkey", "firstvalue"),
			kvs:  nil,
			want: []any{
				zap.Any("firstkey", "firstvalue"),
			},
		},
		{
			name: "user defined field, redefined key",
			ctx:  AddKV(ctx, "firstkey", "firstvalue"),
			kvs:  []any{"firstkey", "othervalue"},
			want: []any{
				zap.Any("firstkey", "othervalue"),
			},
		},
		{
			name: "user defined field, redefined key with invalid type",
			ctx:  AddKV(ctx, "1", "firstvalue"),
			kvs:  []any{2, "othervalue"},
			want: []any{
				zap.Any("1", "firstvalue"),
			},
		},
		{
			name: "user defined field, redefined key, invalid number of arguments",
			ctx:  AddKV(ctx, "firstkey", "firstvalue"),
			kvs:  []any{"firstkey", "othervalue", "secondkey"}, // нет значения для secondkey
			want: []any{
				zap.Any("firstkey", "othervalue"),
			},
		},
		{
			name: "user defined field, redefined key twice",
			ctx:  AddKV(ctx, "firstkey", "firstvalue"),
			kvs:  []any{"secondkey", "secondvalue", "firstkey", "othervalue", "thirdkey", "thirdvalue", "firstkey", "firstvalue"},
			want: []any{
				zap.Any("firstkey", "firstvalue"),
				zap.Any("secondkey", "secondvalue"),
				zap.Any("thirdkey", "thirdvalue"),
			},
		},
		{
			name: "user defined field, redefined key, reverse order",
			ctx:  AddKV(ctx, "firstkey", "firstvalue"),
			kvs:  []any{"secondkey", "secondvalue", "firstkey", "othervalue"},
			want: []any{
				zap.Any("firstkey", "othervalue"),
				zap.Any("secondkey", "secondvalue"),
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// act
			got := mergeKvs(tc.ctx, tc.kvs...)

			// require
			require.Equal(t, tc.want, got)
		})
	}
}

func BenchmarkAddKV(b *testing.B) {
	ctx := context.Background()

	b.Run("empty args", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			AddKV(ctx)
		}
	})

	b.Run("empty context", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			AddKV(ctx, "key1", "value1", "key2", "value2")
		}
	})

	b.Run("populated context", func(b *testing.B) {
		ctx := AddKV(ctx, "key1", "value1", "key2", "value2")
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			AddKV(ctx, "key3", "value3", "key4", "value4")
		}
	})
}

func BenchmarkMergeKVs(b *testing.B) {
	ctx := context.Background()

	b.Run("empty user fields", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			mergeKvs(ctx)
		}
	})

	b.Run("populated user fields, empty final call fields", func(b *testing.B) {
		ctx := AddKV(ctx, "key1", "value1")

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			mergeKvs(ctx)
		}
	})

	b.Run("populated user fields, populated final call fields", func(b *testing.B) {
		ctx := AddKV(ctx, "key1", "value1")

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			mergeKvs(ctx, "key2", "value2")
		}
	})
}

func testNewLogEntry(message string, kvs ...zap.Field) observer.LoggedEntry {
	return observer.LoggedEntry{
		Entry: zapcore.Entry{
			Message: message,
		},
		Context: kvs,
	}
}
