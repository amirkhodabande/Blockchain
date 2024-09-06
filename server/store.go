package server

import (
	"encoding/hex"
	"fmt"
	"sync"

	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/types"
)

type TXStorer interface {
	Put(*blockchain.Transaction) error
	Get(hash string) (*blockchain.Transaction, error)
}

type MemoryTxStore struct {
	lock sync.RWMutex
	txx  map[string]*blockchain.Transaction
}

func NewMemoryTxStore() *MemoryTxStore {
	return &MemoryTxStore{
		txx: make(map[string]*blockchain.Transaction),
	}
}

func (store *MemoryTxStore) Put(tx *blockchain.Transaction) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	hash := hex.EncodeToString(types.HashTransaction(tx))
	store.txx[hash] = tx

	return nil
}

func (store *MemoryTxStore) Get(hash string) (*blockchain.Transaction, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	tx, ok := store.txx[hash]
	if !ok {
		return nil, fmt.Errorf("could not find transaction with hash %s", hash)
	}

	return tx, nil
}

type BlockStorer interface {
	Put(block *blockchain.Block) error
	Get(hash string) (*blockchain.Block, error)
}

type MemoryBlockStore struct {
	lock   sync.RWMutex
	blocks map[string]*blockchain.Block
}

func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{
		blocks: make(map[string]*blockchain.Block),
	}
}

func (store *MemoryBlockStore) Put(block *blockchain.Block) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	hash := hex.EncodeToString(types.HashBlock(block))
	store.blocks[hash] = block
	return nil
}

func (store *MemoryBlockStore) Get(hash string) (*blockchain.Block, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	block, ok := store.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash [%s] does not exists", hash)
	}

	return block, nil
}
