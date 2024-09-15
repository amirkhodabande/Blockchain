package types

import (
	"testing"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/util"
	"github.com/stretchr/testify/require"
)

func TestCalculateRootHash(t *testing.T) {
	var (
		privateKey = crypto.GeneratePrivateKey()
		block      = util.RandomBlock()
		tx         = &blockchain.Transaction{
			Version: 1,
		}
	)

	block.Transactions = append(block.Transactions, tx)
	SignBlock(privateKey, block)

	require.True(t, verifyRootHash(block))
	require.Equal(t, 32, len(block.Header.RootHash))
}

func TestSignVerifyBlock(t *testing.T) {
	var (
		block      = util.RandomBlock()
		privateKey = crypto.GeneratePrivateKey()
		publicKey  = privateKey.Public()
		sign       = SignBlock(privateKey, block)
	)

	require.Equal(t, 64, len(sign.Bytes()))
	require.True(t, sign.Verify(publicKey, HashBlock(block)))

	require.Equal(t, block.PublicKey, publicKey.Bytes())
	require.Equal(t, block.Signature, sign.Bytes())

	require.True(t, VerifyBlock(block))

	invalidPrivateKey := crypto.GeneratePrivateKey()
	block.PublicKey = invalidPrivateKey.Public().Bytes()

	require.False(t, VerifyBlock(block))
}

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)

	require.Equal(t, 32, len(hash))
}
