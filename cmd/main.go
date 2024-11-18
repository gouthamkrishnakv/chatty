package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gouthamkrishnakv/chatty/database"
	"github.com/gouthamkrishnakv/chatty/proto"
	"github.com/gouthamkrishnakv/chatty/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if dbErr := database.Init(); dbErr != nil {
		log.Fatalf("failed to initialize database: %v", dbErr)
	}
	log.Printf("database connected")

	signalCtx, cancel := setupSignalContext()
	defer cancel()

	grpcServer, netListener := setupGRPCServer()

	grpcWaitGroup := sync.WaitGroup{}
	grpcWaitGroup.Add(1)

	go serveGRPC(&grpcWaitGroup, grpcServer, netListener)
	defer grpcWaitGroup.Wait()
	defer shutdownGRPCServer(grpcServer)

	<-signalCtx.Done()
	log.Printf("Caught signal, shutting down")
}

// -- setup methods --

func setupSignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
}

func setupGRPCServer() (*grpc.Server, net.Listener) {
	netListener, listenErr := net.Listen("tcp", ":8000")
	if listenErr != nil {
		log.Fatalf("failed to listen: %v", listenErr)
	}

	// gRPC Server
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	proto.RegisterChatServiceServer(grpcServer, server.NewServer())

	return grpcServer, netListener
}

// -- serve methods --
func serveGRPC(wg *sync.WaitGroup, server *grpc.Server, listener net.Listener) {
	defer wg.Done()
	log.Printf("Starting gRPC server")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC server: %v", err)
	}
	log.Printf("shutting down gRPC server")
}

func shutdownGRPCServer(server *grpc.Server) {
	server.GracefulStop()
}
