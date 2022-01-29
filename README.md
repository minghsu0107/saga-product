# Saga Product
Transaction services of the [saga pattern implementation](https://github.com/minghsu0107/saga-example), including product, order, payment and orchestrator.

Features:
- Entirely event-driven
- Idempotency for all transactions
- Stateless saga orchestrator making transactions scalable
- Sonyflake distributed unique ID generator
- Redis cache for fast data retrieval
- Bloom filters to prevent cache penatration
- Prometheus metrics
- Distributed tracing exporter
  - HTTP server
  - gRPC server
  - gPRC client
  - Event handlers
- Four microservices in a monorepo and compiled to a single binary, minimizing deployment efforts
- Comprehensive application struture with domain-driven design (DDD), decoupling service implementations from configurations and transports
- Compile-time dependecy injection using [wire](https://github.com/google/wire)
- Graceful shutdown
- Unit testing and continuous integration using [Drone CI](https://www.drone.io)
## Usage
See [docker-compose example](https://github.com/minghsu0107/saga-example/blob/main/docker-compose.yaml) for details on how to start each service.
## Exported Metrics
- `APP` could be `product`, `order`, `payment`, or `orchestrator`.
- `HTTPAPP` could be `product`, `order`, or `payment`.

| Metric                                                                                                                                   | Description                                                                                                 | Labels                                                           |
| ---------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| APP_pubsub_subscriber_messages_received_total                                                                                                       | A Prometheus Counter. Counts the number of messages obtained by the subscriber.                             | `acked` ("acked" or "nacked"), `handler_name`, `subscriber_name` |
| APP_pubsub_handler_execution_time_seconds (APP_pubsub_handler_execution_time_seconds_count, APP_pubsub_handler_execution_time_bucket, APP_pubsub_handler_execution_time_seconds_sum) | A Prometheus Histogram. Records the execution time of the handler function wrapped by the middleware.       | `handler_name`, `success` ("true" or "false")                    |
| APP_pubsub_publish_time_seconds (APP_pubsub_publish_time_seconds_count, APP_pubsub_publish_time_seconds_bucket, APP_pubsub_publish_time_seconds_sum)                                 | A Prometheus Histogram. Registers the time of execution of the Publish function of the decorated publisher. | `handler_name`, `success` ("true" or "false"), `publisher_name`  |
| HTTPAPP_http_request_duration_seconds (HTTPAPP_http_request_duration_seconds_count, HTTPAPP_http_request_duration_seconds_bucket, HTTPAPP_http_request_duration_sum)     | A Prometheus histogram. Records the latency of the HTTP requests.                                           | `code`, `handler`, `method`                                      |
| HTTPAPP_http_requests_inflight                                                                                                                   | A Prometheus gauge. Records the number of inflight requests being handled at the same time.                 | `code`, `handler`, `method`                                      |
| HTTPAPP_http_response_size_bytes (HTTPAPP_http_response_size_bytes_count, HTTPAPP_http_response_size_bytes_bucket, HTTPAPP_http_response_size_bytes_sum)                 | A Prometheus histogram. Records the size of the HTTP responses.                                             | `handler`                                                        |
