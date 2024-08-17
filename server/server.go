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

func (server *Server) Start(listenAddress string, bootstrapServers []string) error {
	server.listenAddress = listenAddress
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	blockchain.RegisterBlockChainServer(grpcServer, server)
	server.logger.Info("node running on port:", listenAddress)

	if len(bootstrapServers) > 0 {
		go server.bootstrapNetwork(bootstrapServers)
	}

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

func (server *Server) addPeer(client blockchain.BlockChainClient, message *blockchain.HandshakeMessage) {
	server.peerLock.Lock()
	defer server.peerLock.Unlock()

	if len(message.PeerList) > 0 {
		go server.bootstrapNetwork(message.PeerList)
	}

	server.peers[client] = message
	server.logger.Debugf("[%s]: new peer connected (%s) - height (%d)", server.listenAddress, message.ListenAddress, message.Height)
}

func (server *Server) deletePeer(c blockchain.BlockChainClient) {
	server.peerLock.Lock()
	defer server.peerLock.Unlock()

	delete(server.peers, c)
}

func (server *Server) bootstrapNetwork(addresses []string) error {
	for _, address := range addresses {
		if !server.canConnectWith(address) {
			continue
		}

		server.logger.Debugw("dialing remote server", "from", server.listenAddress, "to", address)
		client, version, err := server.dialRemoteServer(address)
		if err != nil {
			server.logger.Info("handshake error:", err)
			continue
		}

		server.addPeer(client, version)
	}
	return nil
}

func (server *Server) dialRemoteServer(listenAddress string) (blockchain.BlockChainClient, *blockchain.HandshakeMessage, error) {
	client, err := makeBlockChainClient(listenAddress)
	if err != nil {
		return nil, nil, err
	}

	version, err := client.Handshake(context.Background(), server.getVersion())
	if err != nil {
		return nil, nil, err
	}

	return client, version, err
}

func (server *Server) getVersion() *blockchain.HandshakeMessage {
	return &blockchain.HandshakeMessage{
		Version:       "blocker-0.1",
		Height:        0,
		ListenAddress: server.listenAddress,
		PeerList:      server.getPeerList(),
	}
}

func (server *Server) canConnectWith(address string) bool {
	if server.listenAddress == address {
		return false
	}

	connectedPeers := server.getPeerList()
	for _, listenAddress := range connectedPeers {
		if address == listenAddress {
			return false
		}
	}

	return true
}

func (server *Server) getPeerList() []string {
	server.peerLock.RLock()
	defer server.peerLock.RUnlock()

	peers := []string{}
	for _, version := range server.peers {
		peers = append(peers, version.ListenAddress)
	}

	return peers
}

func makeBlockChainClient(listenAddress string) (blockchain.BlockChainClient, error) {
	conn, err := grpc.NewClient(listenAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return blockchain.NewBlockChainClient(conn), nil
}
