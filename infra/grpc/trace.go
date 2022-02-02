package grpc

import (
	"context"
	"encoding/hex"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// LogTraceUnary logs trace id from the incoming request context
func LogTraceUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		b := trace.SpanFromContext(ctx).SpanContext().TraceID()
		grpc_ctxtags.Extract(ctx).Set("traceID", hex.EncodeToString(b[:]))
		return handler(ctx, req)
	}
}
