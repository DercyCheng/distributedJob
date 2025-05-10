package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds tracing configuration
type Config struct {
	ServiceName    string
	JaegerEndpoint string
	SamplingRate   float64
	Enabled        bool
}

// Tracer is the wrapper around OpenTelemetry tracing
type Tracer struct {
	tracer     trace.Tracer
	config     Config
	shutdown   func(context.Context) error
	propagator propagation.TextMapPropagator
}

// NewTracer creates a new tracer
func NewTracer(config Config) (*Tracer, error) {
	if !config.Enabled {
		return &Tracer{
			config: config,
			tracer: trace.NewNoopTracerProvider().Tracer("noop"),
			shutdown: func(context.Context) error {
				return nil
			},
			propagator: propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			),
		}, nil
	}

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(config.SamplingRate)),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ServiceName),
		)),
	)

	otel.SetTracerProvider(tp)
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)

	tracer := tp.Tracer(config.ServiceName)

	return &Tracer{
		tracer: tracer,
		config: config,
		shutdown: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
		propagator: propagator,
	}, nil
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name)
}

// StartSpanWithAttributes starts a new span with attributes
func (t *Tracer) StartSpanWithAttributes(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, trace.WithAttributes(attrs...))
}

// Inject injects span context into carrier
func (t *Tracer) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	t.propagator.Inject(ctx, carrier)
}

// Extract extracts span context from carrier
func (t *Tracer) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return t.propagator.Extract(ctx, carrier)
}

// Shutdown shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	return t.shutdown(ctx)
}

// WithSpan wraps a function with a span
func (t *Tracer) WithSpan(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	ctx, span := t.StartSpan(ctx, name)
	defer span.End()

	return fn(ctx)
}

// RecordError records an error on the current span
func RecordError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
