package server

import (
	"encoding/hex"
	"fmt"

	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/types"
)

type HeaderList struct {
	headers []*blockchain.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: []*blockchain.Header{},
	}
}

func (list *HeaderList) Add(header *blockchain.Header) {
	list.headers = append(list.headers, header)
}

func (list *HeaderList) Get(index int) *blockchain.Header {
	if index > list.Height() {
		panic("index  too high")
	}
	return list.headers[index]
}

func (list *HeaderList) Height() int {
	return list.Len() - 1
}

func (list *HeaderList) Len() int {
	return len(list.headers)
}

type Chain struct {
	blockStore BlockStorer
	headers    *HeaderList
}

func NewChain(blockStorer BlockStorer) *Chain {
	return &Chain{
		blockStore: blockStorer,
		headers:    NewHeaderList(),
	}
}

func (chain *Chain) Height() int {
	return chain.headers.Height()
}

func (chain *Chain) AddBlock(block *blockchain.Block) error {
	chain.headers.Add(block.Header)

	return chain.blockStore.Put(block)
}

func (chain *Chain) GetBlockByHash(hash []byte) (*blockchain.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return chain.blockStore.Get(hashHex)
}

func (chain *Chain) GetBlockByHeight(height int) (*blockchain.Block, error) {
	if chain.Height() < height {
		return nil, fmt.Errorf("given height (%d) too high - height (%d)", height, chain.Height())
	}

	header := chain.headers.Get(height)
	hash := types.HashHeader(header)

	return chain.GetBlockByHash(hash)
}
