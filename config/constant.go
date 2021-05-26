package config

type HTTPContextKey string

var (
	// JWTAuthHeader is the auth header containing customer ID
	JWTAuthHeader = "Authorization"

	// CustomerKey is the key name for retrieving jwt-decoded customer id in a http request context
	CustomerKey HTTPContextKey = "customer_key"

	// IdempotencyKeyHeader is the idempotencyKey header in a product event
	IdempotencyKeyHeader = "Idempotency-Key"
	// HandlerHeader identifies a handler in the ReplyTopic
	HandlerHeader = "Handler"

	// UpdateProductInventoryHandler identifier
	UpdateProductInventoryHandler = "update_product_inventory_handler"
	// RollbackProductInventoryHandler identifier
	RollbackProductInventoryHandler = "rollback_product_inventory_handler"
	// CreateOrderHandler identifier
	CreateOrderHandler = "create_order_handler"
	// RollbackOrderHandler identifier
	RollbackOrderHandler = "rollback_order_handler"
	// CreatePaymentHandler identifier
	CreatePaymentHandler = "create_payment_handler"
	// RollbackPaymentHandler identifier
	RollbackPaymentHandler = "rollback_payment_handler"

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
