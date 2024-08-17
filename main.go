package main

import (
	"context"
	"log"
	"time"

	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	makeServer(":3000", []string{})
	time.Sleep(time.Second)
	makeServer(":4000", []string{":3000"})

	time.Sleep(time.Second)
	makeServer(":5000", []string{":4000"})

	// makeTransaction()
	select {}
}

func makeServer(listenAddress string, bootstrapServers []string) *server.Server {
	server := server.NewServer()
	go server.Start(listenAddress, bootstrapServers)

	return server
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
		Version:       "blocker-0.1",
		Height:        1,
		ListenAddress: "1.1.1.1:4000",
	})
	if err != nil {
		log.Fatal(err)
	}
}
