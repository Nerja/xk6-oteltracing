package oteltracing

import (
	"context"
	"os"
	"time"

	"go.k6.io/k6/js/modules"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type OtelTracingClient struct {
	tracer trace.Tracer
}

func NewOtelTracingClient() (*OtelTracingClient, error) {
	endpoint := "localhost:4319"
	envEndpoint := os.Getenv("K6_OTEL_GRPC_EXPORTER_ENDPOINT")
	if envEndpoint != "" {
		endpoint = envEndpoint
	}

	exporter, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint(endpoint))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("k6"),
		),
	)
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter, sdktrace.WithBatchTimeout(1*time.Second))
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	return &OtelTracingClient{
		tracer: otel.Tracer("k6"),
	}, nil
}

func (c *OtelTracingClient) CreateSpan(spanName string) trace.Span {
	_, span := c.tracer.Start(context.Background(), spanName, trace.WithSpanKind(trace.SpanKindProducer))
	return span
}

func init() {
	client, _ := NewOtelTracingClient()
	modules.Register("k6/x/oteltracing", client)
}
