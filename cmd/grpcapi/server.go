package main

import (
	"ClassConnectRPC/internals/api/handlers"
	"ClassConnectRPC/internals/api/interceptors"
	"ClassConnectRPC/internals/repositories/mongodb"
	"ClassConnectRPC/pkg/utils"
	pb "ClassConnectRPC/proto/gen"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading the .env file", err)
		return
	}
}

func main() {
	mongodb.CreateMongoClient()

	// Start the background goroutine to clean up expired tokens from the blacklist
	go utils.JwtStore.CleanUpExpiredTokens()

	r := interceptors.NewRateLimiter(5, time.Minute)
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(r.RateLimitingInterceptor, interceptors.ResponseTimeInterceptor, interceptors.AuthenticationInterceptor))

	pb.RegisterTeachersServiceServer(s, &handlers.Server{})
	pb.RegisterStudentsSerciesServer(s, &handlers.Server{})
	pb.RegisterExecsServiceServer(s, &handlers.Server{})

	reflection.Register(s)

	port := fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))

	fmt.Printf("gRPC server running on port %s\n", port)

	// The TCP listener acts as a means for our gRPC server to communicate with the outside world
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error listening on the specified port", err)
		return
	}

	err = s.Serve(lis)
	if err != nil {
		log.Fatal("Failed to serve", err)
		return
	}

}
