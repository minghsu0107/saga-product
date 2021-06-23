# Saga Product
Transaction services of the [saga pattern implementation](https://github.com/minghsu0107/saga-example), including product, order, payment and orchestrator.

Features:
- Entirely event-driven
- Idempotency for all transactions
- Stateless saga orchestrator making transactions scalable
- Sonyflake distributed unique ID generator
- Redis cache for fast data retrieval
- Prometheus metrics
- Distributed tracing exporter
  - HTTP server
  - gRPC server
  - gPRC client
- Four microservices in a monorepo and compiled to a single binary, minimizing deployment efforts
- Comprehensive application struture with domain-driven design (DDD), decoupling service implementations from configurations and transports
- Compile-time dependecy injection using [wire](https://github.com/google/wire)
- Graceful shutdown
- Unit testing and continuous integration using [Drone CI](https://www.drone.io)
## Usage
See [docker-compose example](https://github.com/minghsu0107/saga-example/blob/main/docker-compose.yaml) for details on how to start each service.
## Exported Metrics
| Metric                                                                                                                                   | Description                                                                                                 | Labels                                                           |
| ---------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| subscriber_messages_received_total                                                                                                       | A Prometheus Counter. Counts the number of messages obtained by the subscriber.                             | `acked` ("acked" or "nacked"), `handler_name`, `subscriber_name` |
| handler_execution_time_seconds (handler_execution_time_seconds_count, handler_execution_time_bucket, handler_execution_time_seconds_sum) | A Prometheus Histogram. Records the execution time of the handler function wrapped by the middleware.       | `handler_name`, `success` ("true" or "false")                    |
| publish_time_seconds (publish_time_seconds_count, publish_time_seconds_bucket, publish_time_seconds_sum)                                 | A Prometheus Histogram. Registers the time of execution of the Publish function of the decorated publisher. | `handler_name`, `success` ("true" or "false"), `publisher_name`  |
| http_request_duration_seconds (http_request_duration_seconds_count, http_request_duration_seconds_bucket, http_request_duration_sum)     | A Prometheus histogram. Records the latency of the HTTP requests.                                           | `code`, `handler`, `method`                                      |
| http_requests_inflight                                                                                                                   | A Prometheus gauge. Records the number of inflight requests being handled at the same time.                 | `code`, `handler`, `method`                                      |
| http_response_size_bytes (http_response_size_bytes_count, http_response_size_bytes_bucket, http_response_size_bytes_sum)                 | A Prometheus histogram. Records the size of the HTTP responses.                                             | `handler`                                                        |