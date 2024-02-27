package log

import (
	"context"
	"log/slog"
	"os"

	"github.com/lukeshay/records/pkg/config"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DatadogContextHandler struct {
	slog.Handler
}

type ctxKey string

const slogFields ctxKey = "slog_fields"

// Handle adds contextual attributes to the Record before calling the underlying
// handler
func (h DatadogContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if span, found := tracer.SpanFromContext(ctx); found {
		r.Add(
			slog.Group(
				"dd",
				"span_id", span.Context().SpanID(),
				"trace_id", span.Context().TraceID(),
				"service", "deployer",
				"version", config.Version,
				"env", config.Environment,
			),
		)
	}

	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}

	return h.Handler.Handle(ctx, r)
}

// AppendCtx adds an slog attribute to the provided context so that it will be
// included in any Record created with such context
func AppendCtx(parent context.Context, attr slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(slogFields).([]slog.Attr); ok {
		v = append(v, attr)
		return context.WithValue(parent, slogFields, v)
	}

	v := []slog.Attr{}
	v = append(v, attr)

	return context.WithValue(parent, slogFields, v)
}

func InitLogger() {
	zerologL := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
	h := &DatadogContextHandler{slogzerolog.Option{Logger: &zerologL}.NewZerologHandler()}

	logger := slog.New(h)

	slog.SetDefault(logger)
}
