package singboxapi

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewAuthUnaryInterceptor returns a unary client interceptor that attaches bearer auth if secret is non-empty.
func NewAuthUnaryInterceptor(secret string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if secret != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+secret)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NewAuthStreamInterceptor returns a stream client interceptor that attaches bearer auth if secret is non-empty.
func NewAuthStreamInterceptor(secret string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if secret != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+secret)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
