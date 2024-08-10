package types

import (
	"crypto/sha256"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"google.golang.org/protobuf/proto"
)

func SignBlock(privateKey *crypto.PrivateKey, block *blockchain.Block) *crypto.Signature {
	return privateKey.Sign(HashBlock(block))
}

func HashBlock(block *blockchain.Block) []byte {
	b, err := proto.Marshal(block)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}
