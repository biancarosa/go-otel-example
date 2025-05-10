// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const serviceName = "go-otel-api"

// setupOTelTracing initializes an OTLP trace exporter
func setupOTelTracing(ctx context.Context) (*trace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("otel-collector:4318"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(traceProvider)
	return traceProvider, nil
}

// setupOTelMetrics initializes an OTLP metric exporter
func setupOTelMetrics(ctx context.Context) (*metric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint("otel-collector:4318"),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("home-handler")

	_, span := tracer.Start(ctx, "home-operation")
	defer span.End()

	// Add some artificial latency to make metrics more interesting
	sleepTime := rand.Intn(100)
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	span.SetAttributes(attribute.Int("sleep.time.ms", sleepTime))

	// Record a custom metric
	meter := otel.GetMeterProvider().Meter("home-handler")
	latencyHistogram, _ := meter.Int64Histogram("api.home.latency")
	latencyHistogram.Record(ctx, int64(sleepTime), otelmetric.WithAttributes(attribute.String("endpoint", "home")))

	fmt.Fprintf(w, "Hello, OpenTelemetry with Collector!")
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("user-handler")

	ctx, span := tracer.Start(ctx, "user-operation")
	defer span.End()

	// Simulate some backend processing with a child span
	_, childSpan := tracer.Start(ctx, "database-query")
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	childSpan.End()

	// Record a custom metric
	meter := otel.GetMeterProvider().Meter("user-handler")
	counter, _ := meter.Int64Counter("api.user.requests")
	counter.Add(ctx, 1)

	// Add some artificial errors occasionally
	if rand.Intn(10) < 2 {
		span.SetAttributes(attribute.Bool("error", true))
		span.SetAttributes(attribute.String("error.message", "Simulated random error"))

		errorCounter, _ := meter.Int64Counter("api.user.errors")
		errorCounter.Add(ctx, 1, otelmetric.WithAttributes(attribute.String("error.type", "simulated")))

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "An error occurred")
		return
	}

	fmt.Fprintf(w, "User data: {\"id\": 1, \"name\": \"Test User\"}")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC in handler: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize OpenTelemetry
	ctx := context.Background()

	// Set up tracing
	tp, err := setupOTelTracing(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Set up metrics
	mp, err := setupOTelMetrics(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize meter: %v", err)
	}
	defer func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}()

	// Create a counter for total HTTP requests
	meter := otel.GetMeterProvider().Meter(serviceName)
	requestCounter, err := meter.Int64Counter(
		"http.server.request_count",
		otelmetric.WithDescription("Number of HTTP server requests"),
	)
	if err != nil {
		log.Fatalf("Failed to create counter: %v", err)
	}

	// Set up HTTP handlers with OpenTelemetry instrumentation
	http.Handle("/", recoverMiddleware(otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCounter.Add(r.Context(), 1, otelmetric.WithAttributes(attribute.String("endpoint", "home")))
			homeHandler(w, r)
		}),
		"home",
	)))

	http.Handle("/user", recoverMiddleware(otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCounter.Add(r.Context(), 1, otelmetric.WithAttributes(attribute.String("endpoint", "user")))
			userHandler(w, r)
		}),
		"user",
	)))

	http.Handle("/health", recoverMiddleware(otelhttp.NewHandler(
		http.HandlerFunc(healthHandler),
		"health",
	)))

	// Start server
	port := 8080
	log.Printf("Starting server on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
