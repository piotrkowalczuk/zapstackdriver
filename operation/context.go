package operation

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type key struct{}

var contextKey = key{}

// LINK: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogEntryOperation
type operation struct {
	id          uuid.UUID
	first, last bool
	producer    string
}

// MarshalLogObject implements zapcore ObjectMarshaler.
func (o operation) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("id", o.id.String())
	enc.AddBool("first", o.first)
	enc.AddBool("last", o.last)
	if o.producer != "" {
		enc.AddString("producer", o.producer)
	}

	return nil
}

// WithContext returns a copy of parent in which the value associated with key is random UUID.
func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, uuid.New())
}

func FromContext(ctx context.Context, producer string) zapcore.Field {
	return fromContext(ctx, false, false, producer)
}
func FromContextFirst(ctx context.Context, producer string) zapcore.Field {
	return fromContext(ctx, true, false, producer)
}
func FromContextLast(ctx context.Context, producer string) zapcore.Field {
	return fromContext(ctx, false, true, producer)
}

func fromContext(ctx context.Context, first, last bool, producer string) zapcore.Field {
	const key = "logging.googleapis.com/operation"
	if id, ok := ctx.Value(contextKey).(uuid.UUID); ok {
		return zap.Object(key, operation{
			id:       id,
			first:    first,
			last:     last,
			producer: producer,
		})
	}
	return zap.Object(key, operation{
		id:       uuid.New(),
		first:    first,
		last:     last,
		producer: producer,
	})
}
