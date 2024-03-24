package logger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestSweeten(t *testing.T) {
	cases := []struct {
		name string
		kvs  []any
		want []any
		logs []observer.LoggedEntry
	}{
		{
			name: "expect nil with empty args",
		},
		{
			name: "expect zap.Any with multiple string args",
			kvs: []any{
				"a", "1",
				"b", "2",
			},
			want: []any{
				zap.Any("a", "1"),
				zap.Any("b", "2"),
			},
		},
		{
			name: "expect zap.Any with string args mixed with zap.Field",
			kvs: []any{
				"a", "1",
				zap.String("b", "2"),
				"c", "3",
			},
			want: []any{
				zap.Any("a", "1"),
				zap.String("b", "2"),
				zap.Any("c", "3"),
			},
		},
		{
			name: "expect zap.Any with multiple string args, and zap.Error with error field",
			kvs: []any{
				"a", "1",
				assert.AnError,
				"b", "2",
			},
			want: []any{
				zap.Any("a", "1"),
				zap.Error(assert.AnError),
				zap.Any("b", "2"),
			},
		},
		{
			name: "expect first error to be added, other errors are skipped",
			kvs: []any{
				"a", "1",
				assert.AnError,
				errors.New("duplicated error"),
				"b", "2",
			},
			want: []any{
				zap.Any("a", "1"),
				zap.Error(assert.AnError),
				zap.Any("b", "2"),
			},
			logs: []observer.LoggedEntry{
				testNewLogEntry(errMsgMultiple, zap.Error(errors.New("duplicated error"))),
			},
		},
		{
			name: "expect dangling key to be skipped",
			kvs: []any{
				"a", "1",
				zap.String("b", "2"),
				"c", "3",
				"d",
			},
			want: []any{
				zap.Any("a", "1"),
				zap.String("b", "2"),
				zap.Any("c", "3"),
			},
			logs: []observer.LoggedEntry{
				testNewLogEntry(errMsgOddNumber, zap.Any("ignored", "d")),
			},
		},
		{
			name: "expect keys with invalid types to be skipped",
			kvs: []any{
				"a", "1",
				zap.String("b", "2"),
				6, "value for invalid key", // this is going to be skipped
				"c", "3",
			},
			want: []any{
				zap.Any("a", "1"),
				zap.String("b", "2"),
				zap.Any("c", "3"),
			},
			logs: []observer.LoggedEntry{
				testNewLogEntry(errMsgNonStringKey, zap.Array("invalid", invalidPairs{
					invalidPair{
						position: 3,
						key:      6,
						value:    "value for invalid key",
					},
				})),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			core, logs := observer.New(zap.InfoLevel)
			logger := zap.New(core)
			merger := newFieldMerger(logger)

			// act
			got := merger.sweetenFields(tc.kvs)

			// assert
			require.Equal(t, tc.want, got)

			// Если в тесте указаны логи, проверяем.
			if len(tc.logs) > 0 {
				// Получаем логи обсервера.
				records := logs.All()
				require.Equal(t, len(tc.logs), len(records))

				// Проходим по каждой записи и сравниваем
				// сообщение и контекст в виде полей.
				for i, wantRecord := range tc.logs {
					gotRecord := records[i]

					require.Equal(t, wantRecord.Entry.Message, gotRecord.Message)
					require.Equal(t, wantRecord.Context, gotRecord.Context)
				}
			}
		})
	}
}
