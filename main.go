package main

import (
	"context"
	"log"
	"time"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/server"
	"github.com/blockchain/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	makeServer(":3000", []string{}, true)
	time.Sleep(time.Second)
	makeServer(":4000", []string{":3000"}, false)

	time.Sleep(time.Second)
	makeServer(":5000", []string{":4000"}, false)

	for {
		time.Sleep(time.Second * 2)
		makeTransaction()
	}
}

func makeServer(listenAddress string, bootstrapServers []string, isValidator bool) *server.Server {
	serverConfig := server.ServerConfig{
		Version:       "Blocker-1",
		ListenAddress: listenAddress,
	}

	if isValidator {
		serverConfig.PrivateKey = crypto.GeneratePrivateKey()
	}

	server := server.NewServer(serverConfig)
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

	privateKey := crypto.GeneratePrivateKey()
	transaction := &blockchain.Transaction{
		Version: 1,
		Inputs: []*blockchain.TxInput{
			{
				PreviousTxHash:   util.RandomHash(),
				PreviousOutIndex: 0,
				PublicKey:        privateKey.Public().Bytes(),
			},
		},
		Outputs: []*blockchain.TxOutput{
			{
				Amount:  99,
				Address: privateKey.Public().Address().Bytes(),
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.HandleTransaction(ctx, transaction)
	if err != nil {
		log.Fatal(err)
	}
}
