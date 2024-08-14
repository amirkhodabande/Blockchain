package server

import (
	"context"
	"log"
	"net"
	"sync"

	blockchain "github.com/blockchain/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type Server struct {
	version       string
	listenAddress string
	logger        *zap.SugaredLogger

	peerLock sync.RWMutex
	peers    map[blockchain.BlockChainClient]*blockchain.HandshakeMessage
	blockchain.UnimplementedBlockChainServer
}

func NewServer() *Server {
	logger, _ := zap.NewDevelopment()

	return &Server{
		peers:   make(map[blockchain.BlockChainClient]*blockchain.HandshakeMessage),
		version: "blocker-0.1",
		logger:  logger.Sugar(),
	}
}

func (server *Server) addPeer(client blockchain.BlockChainClient, message *blockchain.HandshakeMessage) {
	server.peerLock.Lock()
	defer server.peerLock.Unlock()

	server.logger.Debugf("[%s]: new peer connected (%s) - height (%d)", server.listenAddress, message.ListenAddress, message.Height)

	server.peers[client] = message
}

func (server *Server) deletePeer(c blockchain.BlockChainClient) {
	server.peerLock.Lock()
	defer server.peerLock.Unlock()

	delete(server.peers, c)
}

func (server *Server) BootstrapNetwork(addresses []string) error {
	for _, address := range addresses {
		client, err := makeBlockChainClient(address)
		if err != nil {
			return err
		}

		version, err := client.Handshake(context.Background(), server.getVersion())
		if err != nil {
			server.logger.Info("handshake error:", err)
			continue
		}

		server.addPeer(client, version)
	}
	return nil
}

func (server *Server) Start(listenAddress string) error {
	server.listenAddress = listenAddress
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	blockchain.RegisterBlockChainServer(grpcServer, server)

	server.logger.Info("node running on port:", listenAddress)

	return grpcServer.Serve(ln)
}

func (server *Server) Handshake(ctx context.Context, message *blockchain.HandshakeMessage) (*blockchain.HandshakeMessage, error) {
	client, err := makeBlockChainClient(message.ListenAddress)
	if err != nil {
		return nil, err
	}

	server.addPeer(client, message)

	return server.getVersion(), nil
}

func (server *Server) HandleTransaction(ctx context.Context, tx *blockchain.Transaction) (*blockchain.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	server.logger.Info("received tx from:", peer)

	return &blockchain.Ack{}, nil
}

func (server *Server) getVersion() *blockchain.HandshakeMessage {
	return &blockchain.HandshakeMessage{
		Version:       "blocker-0.1",
		Height:        0,
		ListenAddress: server.listenAddress,
	}
}

func makeBlockChainClient(listenAddress string) (blockchain.BlockChainClient, error) {
	conn, err := grpc.NewClient(listenAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return blockchain.NewBlockChainClient(conn), nil
}
