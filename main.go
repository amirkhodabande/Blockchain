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
	makeServer(":4300", []string{":3000"})

	// go func() {
	// 	for {
	// 		time.Sleep(2 * time.Second)
	// 		makeTransaction()
	// 	}
	// }()
	select {}
}

func makeServer(listenAddress string, bootstrapServers []string) *server.Server {
	server := server.NewServer()
	go server.Start(listenAddress)

	if len(bootstrapServers) > 0 {
		if err := server.BootstrapNetwork(bootstrapServers); err != nil {
			log.Fatal(err)
		}
	}

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
