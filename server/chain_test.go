package server

import (
	"encoding/hex"
	"testing"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/types"
	"github.com/blockchain/util"
	"github.com/stretchr/testify/require"
)

func randomBlock(t *testing.T, chain *Chain) *blockchain.Block {
	privateKey := crypto.GeneratePrivateKey()

	block := util.RandomBlock()

	previousBlock, err := chain.GetBlockByHeight(chain.Height())
	require.Nil(t, err)

	block.Header.PreviousHash = types.HashBlock(previousBlock)

	types.SignBlock(privateKey, block)
	return block
}

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
	require.Equal(t, 0, chain.Height())

	_, err := chain.GetBlockByHeight(0)
	require.Nil(t, err)
}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())

	for i := 1; i < 100; i++ {
		block := randomBlock(t, chain)

		require.Nil(t, chain.AddBlock(block))
		require.Equal(t, i, chain.Height())
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())

	for i := 1; i < 100; i++ {
		block := randomBlock(t, chain)

		hashedBlock := types.HashBlock(block)

		require.Nil(t, chain.AddBlock(block))

		fetchedBlock, err := chain.GetBlockByHash(hashedBlock)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlockByHeight)
	}
}

func TestAddBlockWithTx(t *testing.T) {
	var (
		chain      = NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
		block      = randomBlock(t, chain)
		privateKey = crypto.NewPrivateKeyFromString(seed)
		recipient  = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)
	fetchedTransaction, err := chain.txStore.Get("f8dbd676cfebe8608250ef3bbf9b6a46cfb97ff58796329a0135b5511af57440")
	require.Nil(t, err)

	inputs := []*blockchain.TxInput{
		{
			PreviousTxHash:   types.HashTransaction(fetchedTransaction),
			PreviousOutIndex: 0,
			PublicKey:        privateKey.Public().Bytes(),
		},
	}
	outputs := []*blockchain.TxOutput{
		{
			Amount:  100,
			Address: recipient,
		},
		{
			Amount: 900,
			Address: privateKey.Public().Address().Bytes(),
		},
	}
	tx := &blockchain.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}
	block.Transactions = append(block.Transactions, tx)
	require.Nil(t, chain.AddBlock(block))

	txHash := hex.EncodeToString(types.HashTransaction(tx))

	fetchedTx, err := chain.txStore.Get(txHash)

	require.Nil(t, err)
	require.Equal(t, tx, fetchedTx)
}
