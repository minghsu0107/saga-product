package config

type HTTPContextKey string

var (
	// JWTAuthHeader is the auth header containing customer ID
	JWTAuthHeader = "Authorization"
	// CustomerKey is the key name for retrieving jwt-decoded customer id in a http request context
	CustomerKey HTTPContextKey = "customer_key"
	// IdempotencyKeyHeader is the idempotencyKey header
	IdempotencyKeyHeader = "Idempotency-Key"
)
