package interceptors

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ResponseTimeInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Initialize the start time
	start := time.Now()

	// Call the handler to proceed with the client request
	resp, err := handler(ctx, req)

	// Calculate the time elapsed since the request of the request
	duration := time.Since(start)

	// Log the details with duraton
	status, _ := status.FromError(err)
	fmt.Printf("Method: %s, Status: %d, Duration: %v\n", info.FullMethod, status.Code(), duration)

	md := metadata.Pairs("X-Response-Time", duration.String())
	grpc.SendHeader(ctx, md)

	return resp, err
}
