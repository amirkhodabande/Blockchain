package server

import (
	"context"
	"encoding/hex"
	"log"
	"net"
	"sync"
	"time"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

const blockTime = time.Second * 5

type Mempool struct {
	lock         sync.RWMutex
	transactions map[string]*blockchain.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[string]*blockchain.Transaction),
	}
}

func (pool *Mempool) Clear() []*blockchain.Transaction {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	transactions := make([]*blockchain.Transaction, len(pool.transactions))
	i := 0
	for key, value := range pool.transactions {
		delete(pool.transactions, key)
		transactions[i] = value
		i++
	}

	return transactions
}

func (pool *Mempool) Len() int {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	return len(pool.transactions)
}

func (pool *Mempool) Has(transaction *blockchain.Transaction) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	hash := hex.EncodeToString(types.HashTransaction(transaction))
	_, ok := pool.transactions[hash]
	return ok
}

func (pool *Mempool) Add(transaction *blockchain.Transaction) bool {
	if pool.Has(transaction) {
		return false
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	hash := hex.EncodeToString(types.HashTransaction(transaction))
	pool.transactions[hash] = transaction
	return true
}

type ServerConfig struct {
	Version       string
	ListenAddress string
	PrivateKey    *crypto.PrivateKey
}

type Server struct {
	ServerConfig

	logger *zap.SugaredLogger

	peerLock sync.RWMutex
	peers    map[blockchain.BlockChainClient]*blockchain.HandshakeMessage
	mempool  *Mempool

	blockchain.UnimplementedBlockChainServer
}

func NewServer(config ServerConfig) *Server {
	logger, _ := zap.NewDevelopment()

	return &Server{
		peers:        make(map[blockchain.BlockChainClient]*blockchain.HandshakeMessage),
		logger:       logger.Sugar(),
		mempool:      NewMempool(),
		ServerConfig: config,
	}
}

func (server *Server) Start(listenAddress string, bootstrapServers []string) error {
	server.ListenAddress = listenAddress
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

	if server.PrivateKey != nil {
		go server.validatorLoop()
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
	hash := hex.EncodeToString(types.HashTransaction(tx))

	if server.mempool.Add(tx) {
		server.logger.Debugw("received transaction", "from", peer.Addr, "hash", hash, "we", server.ListenAddress)

		go func() {
			if err := server.broadcast(tx); err != nil {
				server.logger.Errorw("broadcast error", err)
			}
		}()
	}

	return &blockchain.Ack{}, nil
}

func (server *Server) validatorLoop() {
	server.logger.Infow("stating validator loop", "publicKey", server.PrivateKey.Public(), "blockTime", blockTime)
	ticker := time.NewTicker(blockTime)

	for {
		<-ticker.C

		transactions := server.mempool.Clear()

		server.logger.Debugw("time to create new block", "lenTx", len(transactions))
	}
}

func (server *Server) broadcast(message any) error {
	for peer := range server.peers {
		switch v := message.(type) {
		case *blockchain.Transaction:
			_, err := peer.HandleTransaction(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (server *Server) addPeer(client blockchain.BlockChainClient, message *blockchain.HandshakeMessage) {
	server.peerLock.Lock()
	defer server.peerLock.Unlock()

	if len(message.PeerList) > 0 {
		go server.bootstrapNetwork(message.PeerList)
	}

	server.peers[client] = message
	server.logger.Debugf("[%s]: new peer connected (%s) - height (%d)", server.ListenAddress, message.ListenAddress, message.Height)
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
    
		server.logger.Debugw("dialing remote server", "from", server.ListenAddress, "to", address)
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
		ListenAddress: server.ListenAddress,
		PeerList:      server.getPeerList(),
	}
}

func (server *Server) canConnectWith(address string) bool {
	if server.ListenAddress == address {
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
