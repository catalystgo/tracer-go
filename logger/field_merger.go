package logger

import (
	"context"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type invalidPairs []invalidPair

func mergeKvs(ctx context.Context, otherKVs ...any) []any {
	kvsFromContext := getKvsFromContext(ctx)
	if len(kvsFromContext) == 0 {
		return globalMerger.sweetenFields(otherKVs)
	}
	return mergeFields(kvsFromContext, globalMerger.sweetenFields(otherKVs))
}

type invalidPair struct {
	position   int
	key, value any
}

func (p invalidPair) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("position", int64(p.position))
	zap.Any("key", p.key).AddTo(enc)
	zap.Any("value", p.value).AddTo(enc)

	return nil
}
func (ps invalidPairs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	var err error
	for i := range ps {
		err = multierr.Append(err, enc.AppendObject(ps[i]))
	}
	return err
}

const (
	errMsgOddNumber    = "Ignored key without a value."
	errMsgNonStringKey = "Ignored key-value pairs with non-string keys."
	errMsgMultiple     = "Multiple errors without a key."
)

var globalMerger = newFieldMerger(New(zap.ErrorLevel).Desugar())

type fieldMerger struct {
	logger *zap.Logger
}

func newFieldMerger(logger *zap.Logger) *fieldMerger {
	return &fieldMerger{logger: logger}
}

// sweetenFields Function copied from `zap` source code
// https://github.com/uber-go/zap/blob/master/sugar.go
func (m *fieldMerger) sweetenFields(args []any) []any {
	if len(args) == 0 {
		return nil
	}

	var (
		fields    = make([]any, 0, len(args))
		invalid   invalidPairs
		seenError bool
	)

	for i := 0; i < len(args); {
		if f, ok := args[i].(zap.Field); ok {
			fields = append(fields, f)
			i++
			continue
		}

		if err, ok := args[i].(error); ok {
			if !seenError {
				seenError = true
				fields = append(fields, zap.Error(err))
			} else {
				m.logger.Error(errMsgMultiple, zap.Error(err))
			}
			i++
			continue
		}

		if i == len(args)-1 {
			m.logger.Error(errMsgOddNumber, zap.Any("ignored", args[i]))
			break
		}

		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			if cap(invalid) == 0 {
				invalid = make(invalidPairs, 0, len(args)/2)
			}

			invalid = append(invalid, invalidPair{i, key, val})
		} else {
			fields = append(fields, zap.Any(keyStr, val))
		}
		i += 2
	}

	if len(invalid) > 0 {
		m.logger.Error(errMsgNonStringKey, zap.Array("invalid", invalid))
	}

	return fields
}
