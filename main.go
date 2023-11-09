package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/benhenryhunter/grpc-server-client-testing/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	s := NewServer()

	// init server
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 5010))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	proto.RegisterStreamerServer(server, s)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatal("grpc server closed")
		}
	}()

	// init client connection
	conn, err := grpc.Dial("localhost:5010", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to create mev-relay-proxy client connection", zap.Error(err))
	}

	clientOne := proto.NewStreamerClient(conn)

	// init client connection
	conn2, err := grpc.Dial("localhost:5010", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to create mev-relay-proxy client connection", zap.Error(err))
	}

	clientTwo := proto.NewStreamerClient(conn2)

	c1 := &Client{StreamerClient: clientOne, ID: 1, DisconnectModulo: 5}
	c2 := &Client{StreamerClient: clientTwo, ID: 2, DisconnectModulo: 3}

	go c1.StreamStuff(context.Background())
	go c2.StreamStuff(context.Background())

	c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("Shutting down")
	// server.Stop()

}
