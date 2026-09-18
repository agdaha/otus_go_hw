package internalgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type capturingLogger struct {
	messages []string
}

func (l *capturingLogger) Info(msg string) { l.messages = append(l.messages, msg) }
func (l *capturingLogger) Warn(string)     {}
func (l *capturingLogger) Error(string)    {}

func TestLoggingInterceptorSuccess(t *testing.T) {
	logger := &capturingLogger{}
	interceptor := loggingInterceptor(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/event.EventService/CreateEvent"}

	resp, err := interceptor(context.Background(), "req", info, func(_ context.Context, _ any) (any, error) {
		return "resp", nil
	})

	require.NoError(t, err)
	require.Equal(t, "resp", resp)
	require.Len(t, logger.messages, 1)
	require.Contains(t, logger.messages[0], "/event.EventService/CreateEvent")
	require.Contains(t, logger.messages[0], codes.OK.String())
}

func TestLoggingInterceptorError(t *testing.T) {
	logger := &capturingLogger{}
	interceptor := loggingInterceptor(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/event.EventService/DeleteEvent"}
	wantErr := status.Error(codes.NotFound, "event not found")

	_, err := interceptor(context.Background(), "req", info, func(_ context.Context, _ any) (any, error) {
		return nil, wantErr
	})

	require.ErrorIs(t, err, wantErr)
	require.Len(t, logger.messages, 1)
	require.Contains(t, logger.messages[0], codes.NotFound.String())
}

func TestToGRPCError(t *testing.T) {
	require.Equal(t, codes.Internal, status.Code(toGRPCError(context.DeadlineExceeded)))
}
