package main

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func otelConfig() (endpoint, licenseKey, serviceName, serviceEnv string) {
	endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://otlp.nr-data.net"
	}
	licenseKey = os.Getenv("NEW_RELIC_LICENSE_KEY")
	serviceName = os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "go-echo"
	}
	serviceEnv = os.Getenv("SERVICE_ENV")
	if serviceEnv == "" {
		serviceEnv = "production"
	}
	return
}

func newResource(serviceName, serviceEnv string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.DeploymentEnvironmentKey.String(serviceEnv),
	)
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	endpoint, licenseKey, serviceName, serviceEnv := otelConfig()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithHeaders(map[string]string{
			"api-key": licenseKey,
		}),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(newResource(serviceName, serviceEnv)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func initMeter(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	endpoint, licenseKey, serviceName, serviceEnv := otelConfig()

	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(endpoint),
		otlpmetrichttp.WithHeaders(map[string]string{
			"api-key": licenseKey,
		}),
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithResource(newResource(serviceName, serviceEnv)),
	)
	otel.SetMeterProvider(mp)
	return mp, nil
}
