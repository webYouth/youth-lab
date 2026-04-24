// Main entrypoint starts Gin HTTP and gRPC servers.
package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"youthlab/server/gen/blogpb"
	"youthlab/server/internal/handler"
	"youthlab/server/internal/service"
)

func main() {
	httpAddr := getEnv("GIN_HTTP_ADDR", ":8080")
	grpcAddr := getEnv("GRPC_ADDR", ":50051")

	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen gRPC on %s: %v", grpcAddr, err)
	}

	grpcServer := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(grpcServer, service.NewBlogService())

	go func() {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if serveErr := grpcServer.Serve(grpcListener); serveErr != nil {
			log.Fatalf("gRPC server error: %v", serveErr)
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	handler.RegisterHealthRoutes(router)

	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if runErr := router.Run(httpAddr); runErr != nil {
			log.Fatalf("HTTP server error: %v", runErr)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down gRPC server")
	grpcServer.GracefulStop()
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
