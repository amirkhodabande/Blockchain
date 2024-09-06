package server

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/types"
)

const seed = "ca2c1cdf74722ada1e4d152c96a8d2b184a656907b697bd3fd2e1e8abc377da9"

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
	txStore    TXStorer
	blockStore BlockStorer
	headers    *HeaderList
}

func NewChain(blockStorer BlockStorer, txStorer TXStorer) *Chain {
	chain := &Chain{
		txStore:    txStorer,
		blockStore: blockStorer,
		headers:    NewHeaderList(),
	}
	chain.addBlock(createGenesisBlock())

	return chain
}

func (chain *Chain) Height() int {
	return chain.headers.Height()
}

func (chain *Chain) AddBlock(block *blockchain.Block) error {
	if err := chain.ValidateBlock(block); err != nil {
		return err
	}

	return chain.addBlock(block)
}

func (chain *Chain) addBlock(block *blockchain.Block) error {
	chain.headers.Add(block.Header)

	for _, tx := range block.Transactions {
		if err := chain.txStore.Put(tx); err != nil {
			return err
		}
	}

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

func (chain *Chain) ValidateBlock(block *blockchain.Block) error {
	if !types.VerifyBlock(block) {
		return fmt.Errorf("invalid block signature")
	}

	currentBlock, err := chain.GetBlockByHeight(chain.Height())
	if err != nil {
		return err
	}

	hash := types.HashBlock(currentBlock)
	if !bytes.Equal(hash, block.Header.PreviousHash) {
		return fmt.Errorf("invalid previous block hash")
	}

	return nil
}

func createGenesisBlock() *blockchain.Block {
	privateKey := crypto.NewPrivateKeyFromString(seed)

	block := &blockchain.Block{
		Header: &blockchain.Header{
			Version: 1,
		},
	}

	tx := &blockchain.Transaction{
		Version: 1,
		Inputs:  []*blockchain.TxInput{},
		Outputs: []*blockchain.TxOutput{
			{
				Amount:  1000,
				Address: privateKey.Public().Address().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privateKey, block)

	return block
}
