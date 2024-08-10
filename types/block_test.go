package types

import (
	"testing"

	"github.com/blockchain/crypto"
	"github.com/blockchain/util"
	"github.com/stretchr/testify/assert"
)

func TestSignBlock(t *testing.T) {
	var (
		block      = util.RandomBlock()
		privateKey = crypto.GeneratePrivateKey()
		publicKey  = privateKey.Public()
		sign       = SignBlock(privateKey, block)
	)

	assert.Equal(t, 64, len(sign.Bytes()))
	assert.True(t, sign.Verify(publicKey, HashBlock(block)))
}

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)

	assert.Equal(t, 32, len(hash))
}
