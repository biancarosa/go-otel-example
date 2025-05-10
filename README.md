# Go HTTP API with OpenTelemetry Collector and Prometheus

A simple Go HTTP API instrumented with OpenTelemetry, using a Collector to process telemetry data before sending it to Prometheus.

## Architecture

This setup demonstrates a production-grade observability pipeline:

1. **Go Application**: Instrumented with OpenTelemetry SDK
   - Generates both metrics and traces
   - Sends telemetry to the OpenTelemetry Collector

2. **OpenTelemetry Collector**: Central telemetry processing
   - Receives data from the application via OTLP
   - Processes data (batching, filtering, etc.)
   - Exports metrics to Prometheus
   - Exports traces to Jaeger and Honeycomb

3. **Prometheus**: Metrics storage
   - Scrapes metrics from the OpenTelemetry Collector
   - Stores time-series data
   - Provides query capabilities

4. **Jaeger**: Traces visualization
   - Receives traces from the OpenTelemetry Collector
   - Provides visualization and exploration of trace data

5. **Honeycomb**: Cloud-based observability
   - Receives traces from the OpenTelemetry Collector
   - Provides advanced querying and visualization features

## Benefits of Using the Collector

- **Protocol translation**: Accepts OTLP and exports to Prometheus format
- **Processing pipeline**: Can filter, transform, and enrich telemetry data
- **Buffer and retry**: Handles network issues and backpressure
- **Scaling**: Can be deployed as an agent or gateway depending on needs
- **Future extensibility**: Can easily add new exporters (e.g., for logs or other backends)

## Getting Started

1. Clone this repository and navigate to the directory

2. Start the Docker containers:
   ```bash
   docker-compose up -d
   ```

3. Access the services:
   - Go API: http://localhost:8080
   - Prometheus: http://localhost:9090

## Explore the Metrics

In Prometheus, try these queries:
- `go_otel_http_server_request_count` - Total number of HTTP requests
- `go_otel_api_user_errors` - Count of errors in the user endpoint
- `go_otel_api_home_latency` - Latency histogram for the home endpoint

## Files in this Project

- `main.go`: Go API with OpenTelemetry instrumentation
- `go.mod`: Go module dependencies
- `Dockerfile`: Container definition for the Go API
- `otel-collector-config.yaml`: Configuration for the OpenTelemetry Collector
- `prometheus.yml`: Prometheus scrape configuration
- `docker-compose.yml`: Docker Compose configuration for all services

## More Information

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
- [Prometheus Documentation](https://prometheus.io/docs/introduction/overview/)

## Honeycomb Integration

This project also supports sending telemetry data to [Honeycomb](https://www.honeycomb.io/), a cloud-based observability platform.

To configure Honeycomb:

1. Get your API key from the Honeycomb UI

2. Set environment variables before starting the services:
   ```bash
   export HONEYCOMB_API_KEY=your_api_key
   export HONEYCOMB_DATASET=go-otel-example
   docker-compose up -d
   ```

3. Alternatively, create a `.env` file in the project root:
   ```
   HONEYCOMB_API_KEY=your_api_key
   HONEYCOMB_DATASET=go-otel-example
   ```

4. Access your traces in the Honeycomb UI