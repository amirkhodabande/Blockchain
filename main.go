package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	server := server.NewServer()

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	blockchain.RegisterBlockChainServer(grpcServer, server)
	fmt.Println("node running on port:", 3000)

	go func() {
		for {
			time.Sleep(2 * time.Second)
			makeTransaction()
		}
	}()

	grpcServer.Serve(ln)
}

func makeTransaction() {
	conn, err := grpc.NewClient(":3000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := blockchain.NewBlockChainClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// _, err = c.HandleTransaction(ctx, &blockchain.Transaction{})
	_, err = c.Handshake(ctx, &blockchain.HandshakeMessage{
		Version: "blocker-0.1",
		Height:  1,
	})
	if err != nil {
		log.Fatal(err)
	}
}
