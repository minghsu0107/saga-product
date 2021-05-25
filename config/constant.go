package config

type HTTPContextKey string

var (
	// JWTAuthHeader is the auth header containing customer ID
	JWTAuthHeader = "Authorization"
	// CustomerKey is the key name for retrieving jwt-decoded customer id in a http request context
	CustomerKey HTTPContextKey = "customer_key"
	// IdempotencyKeyHeader is the idempotencyKey header
	IdempotencyKeyHeader = "Idempotency-Key"

	// PurchaseTopic is the subscribed topic for new purchase
	PurchaseTopic = "purchase"
	// PurchaseResultTopic is the topic to which we publish new purchase result
	PurchaseResultTopic = "purchase_result"

	// ReplyTopic is saga step reply topic
	ReplyTopic = "reply"
	// ProductTopic is publish product topic
	ProductTopic = "product"
	// OrderTopic is publish order topic
	OrderTopic = "order"
	// PaymentTopic is publish payment topic
	PaymentTopic = "payment"
)
