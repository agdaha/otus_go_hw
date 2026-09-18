package internalgrpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func loggingInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		peerAddr := "-"
		if p, ok := peer.FromContext(ctx); ok {
			peerAddr = p.Addr.String()
		}

		logger.Info(fmt.Sprintf("%s [%s] %s %s %dms",
			peerAddr,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			info.FullMethod,
			status.Code(err).String(),
			time.Since(start).Milliseconds(),
		))
		return resp, err
	}
}
