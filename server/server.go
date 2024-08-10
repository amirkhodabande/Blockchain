package server

import (
	"context"
	"fmt"

	blockchain "github.com/blockchain/proto"
	"google.golang.org/grpc/peer"
)

type Server struct {
	version string
	// peers map[net.Addr]*grpc.ClientConn
	blockchain.UnimplementedBlockChainServer
}

func NewServer() *Server {
	return &Server{
		version: "blocker-0.1",
	}
}

func (server *Server) Handshake(ctx context.Context, message *blockchain.HandshakeMessage) (*blockchain.HandshakeMessage, error) {
	ourVersion := &blockchain.HandshakeMessage{
		Version: server.version,
		Height:  100,
	}

	peer, _ := peer.FromContext(ctx)
	fmt.Printf("received version from %s: %+v\n", message, peer.Addr)

	return ourVersion, nil
}

func (server *Server) HandleTransaction(ctx context.Context, tx *blockchain.Transaction) (*blockchain.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Println("received tx from:", peer)
	return &blockchain.Ack{}, nil
}
