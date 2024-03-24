package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey int

const (
	loggerContextKey contextKey = iota
)

// ToContext Attaches a logger to context
func ToContext(ctx context.Context, l *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}

// FromContext Gets the logger from contet
func FromContext(ctx context.Context) *zap.SugaredLogger {
	l := getLogger(ctx)

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		// if span is valid - inject trace_id & span_id to logger
		l = loggerWithSpanContext(l, span.SpanContext())
	}

	return l
}

// LevelFromContext Gets the log_level from the context logger
func LevelFromContext(ctx context.Context) zapcore.Level {
	return FromContext(ctx).Level()
}

// getLogger Get from context object
// if none such logger exists then use the global logger instead
func getLogger(ctx context.Context) *zap.SugaredLogger {
	l := global

	if logger, ok := ctx.Value(loggerContextKey).(*zap.SugaredLogger); ok {
		l = logger
	}

	return l
}

// loggerWithSpanContext Inject trace_id & span_id values to logger
func loggerWithSpanContext(l *zap.SugaredLogger, spanCtx trace.SpanContext) *zap.SugaredLogger {
	return l.Desugar().With(
		zap.Stringer("trace_id", spanCtx.TraceID()),
		zap.Stringer("span_id", spanCtx.SpanID()),
	).Sugar()
}

// WithName Set a name for the logger
func WithName(ctx context.Context, name string) context.Context {
	l := FromContext(ctx).Named(name)
	return ToContext(ctx, l)
}

// WithKV Adds KV pair to logger from context
func WithKV(ctx context.Context, key string, value any) context.Context {
	l := FromContext(ctx).With(key, value)
	return ToContext(ctx, l)
}

// WithFields Adds fields to logger from context
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	l := FromContext(ctx).Desugar().With(fields...).Sugar()
	return ToContext(ctx, l)
}

// mergeFields Merges fields passed by caller with context ones
// stored in the context, if a field has the same key in both arrays
// fields from caller override the context ones
func mergeFields(filedFromContext, fieldsFromCaller []any) []any {
	merged := make([]any, 0, len(filedFromContext)+len(fieldsFromCaller))
	merged = append(merged, filedFromContext...)

	for i := 0; i < len(fieldsFromCaller); i++ {
		wasFieldFromContextReplaced := false

		newField := fieldsFromCaller[i].(zap.Field)

		for j := range merged {
			oldField := merged[j].(zap.Field)

			// if both fields have the same key
			// then use the one passed by the user
			if oldField.Key == newField.Key {
				merged[j] = newField
				wasFieldFromContextReplaced = true
			}
		}

		// append only if the field's key
		// didn't exist in the
		if !wasFieldFromContextReplaced {
			merged = append(merged, newField)
		}
	}

	return merged
}

type logFieldKeyType string

var logFieldKey = logFieldKeyType("logger-fields")

// AddKV Adds KV pairs to
func AddKV(ctx context.Context, kvs ...any) context.Context {
	if len(kvs) == 0 {
		return ctx
	}

	kvsFromContext := getKvsFromContext(ctx)
	additionalFields := globalMerger.sweetenFields(kvs)

	return context.WithValue(ctx, logFieldKey, mergeFields(kvsFromContext, additionalFields))
}

// getKvsFromContext Gets the KV values stored by AddKV function
func getKvsFromContext(ctx context.Context) []any {
	if kvs, ok := ctx.Value(logFieldKey).([]any); ok {

		// create a copy from the fields since the
		// array can be modified by AddKV
		kvsCopy := make([]any, len(kvs))
		copy(kvsCopy, kvs)
		return kvsCopy
	}
	return nil
}
