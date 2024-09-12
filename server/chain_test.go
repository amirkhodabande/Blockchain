package server

import (
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
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore(), NewMemoryUTXOStore())
	require.Equal(t, 0, chain.Height())

	_, err := chain.GetBlockByHeight(0)
	require.Nil(t, err)
}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore(), NewMemoryUTXOStore())

	for i := 1; i < 100; i++ {
		block := randomBlock(t, chain)

		require.Nil(t, chain.AddBlock(block))
		require.Equal(t, i, chain.Height())
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore(), NewMemoryUTXOStore())

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

func TestAddBlockWithInsufficientBalance(t *testing.T) {
	var (
		chain      = NewChain(NewMemoryBlockStore(), NewMemoryTxStore(), NewMemoryUTXOStore())
		block      = randomBlock(t, chain)
		privateKey = crypto.NewPrivateKeyFromString(seed)
		recipient  = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)
	previousTransaction, err := chain.txStore.Get("f8dbd676cfebe8608250ef3bbf9b6a46cfb97ff58796329a0135b5511af57440")
	require.Nil(t, err)

	inputs := []*blockchain.TxInput{
		{
			PreviousTxHash:   types.HashTransaction(previousTransaction),
			PreviousOutIndex: 0,
			PublicKey:        privateKey.Public().Bytes(),
		},
	}
	outputs := []*blockchain.TxOutput{
		{
			Amount:  1001,
			Address: recipient,
		},
	}
	tx := &blockchain.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}
	signature := types.SignTransaction(privateKey, tx)
	tx.Inputs[0].Signature = signature.Bytes()

	block.Transactions = append(block.Transactions, tx)
	require.NotNil(t, chain.AddBlock(block))
}

func TestAddBlockWithTx(t *testing.T) {
	var (
		chain      = NewChain(NewMemoryBlockStore(), NewMemoryTxStore(), NewMemoryUTXOStore())
		block      = randomBlock(t, chain)
		privateKey = crypto.NewPrivateKeyFromString(seed)
		recipient  = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)
	previousTransaction, err := chain.txStore.Get("f8dbd676cfebe8608250ef3bbf9b6a46cfb97ff58796329a0135b5511af57440")
	require.Nil(t, err)

	inputs := []*blockchain.TxInput{
		{
			PreviousTxHash:   types.HashTransaction(previousTransaction),
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
			Amount:  900,
			Address: privateKey.Public().Address().Bytes(),
		},
	}
	tx := &blockchain.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}
	signature := types.SignTransaction(privateKey, tx)
	tx.Inputs[0].Signature = signature.Bytes()

	block.Transactions = append(block.Transactions, tx)
	require.Nil(t, chain.AddBlock(block))
}
