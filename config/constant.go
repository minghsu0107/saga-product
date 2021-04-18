package config

type HTTPContextKey string

var (
	// JWTAuthHeader is the auth header containing customer ID
	JWTAuthHeader = "Authorization"
	// InvalidationTopic is the cache invalidation topic
	InvalidationTopic = "invalidate_cache"
	// CustomerKey is the key name for retrieving jwt-decoded customer id in a http request context
	CustomerKey HTTPContextKey = "customer_key"
)
