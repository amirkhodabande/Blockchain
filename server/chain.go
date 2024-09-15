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

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	txStore    TXStorer
	blockStore BlockStorer
	utxoStore  UTXOStorer
	headers    *HeaderList
}

func NewChain(blockStorer BlockStorer, txStorer TXStorer, utxoStore UTXOStorer) *Chain {
	chain := &Chain{
		txStore:    txStorer,
		blockStore: blockStorer,
		utxoStore:  utxoStore,
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

		hash := hex.EncodeToString(types.HashTransaction(tx))

		for index, output := range tx.Outputs {
			utxo := &UTXO{
				Hash:     hash,
				Amount:   output.Amount,
				OutIndex: index,
				Spent:    false,
			}

			if err := chain.utxoStore.Put(utxo); err != nil {
				return err
			}
		}

		for index, input := range tx.Inputs {
			key := fmt.Sprintf("%s_%d", hex.EncodeToString(input.PreviousTxHash), index)
			utxo, err := chain.utxoStore.Get(key)
			if err != nil {
				return err
			}

			utxo.Spent = true
			if err := chain.utxoStore.Put(utxo); err != nil {
				return err
			}
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

	for _, tx := range block.Transactions {
		if err := chain.validateTransaction(tx); err != nil {
			return err
		}
	}

	return nil
}

func (chain *Chain) validateTransaction(tx *blockchain.Transaction) error {
	if !types.VerifyTransaction(tx) {
		return fmt.Errorf("invalid transaction signature")
	}

	nInputs := len(tx.Inputs)
	sumInputs := 0

	for i := 0; i < nInputs; i++ {
		previousHash := hex.EncodeToString(tx.Inputs[i].PreviousTxHash)
		key := fmt.Sprintf("%s_%d", previousHash, i)
		utxo, err := chain.utxoStore.Get(key)
		sumInputs += int(utxo.Amount)

		if err != nil {
			return err
		}
		if utxo.Spent {
			return fmt.Errorf("input %d of tx %s is already spent", i, previousHash)
		}
	}

	sumOutputs := 0
	for _, output := range tx.Outputs {
		sumOutputs += int(output.Amount)
	}

	if sumInputs < sumOutputs {
		return fmt.Errorf("insufficient balance got (%d) spending (%d)", sumInputs, sumOutputs)
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
