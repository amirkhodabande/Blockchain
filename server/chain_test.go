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
	chain := NewChain(NewMemoryBlockStore())
	require.Equal(t, 0, chain.Height())

	_, err := chain.GetBlockByHeight(0)
	require.Nil(t, err)
}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

	for i := 1; i < 100; i++ {
		block := randomBlock(t, chain)

		require.Nil(t, chain.AddBlock(block))
		require.Equal(t, i, chain.Height())
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore())

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
