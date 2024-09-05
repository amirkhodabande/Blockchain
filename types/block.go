package types

import (
	"crypto/sha256"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"google.golang.org/protobuf/proto"
)

func VerifyBlock(block *blockchain.Block) bool {
	if len(block.PublicKey) != crypto.PublicKeyLen {
		return false
	}

	if len(block.Signature) != crypto.SignatureLen {
		return false
	}

	sig := crypto.SignatureFromBytes(block.Signature)
	publicKey := crypto.PublicKeyFromBytes(block.PublicKey)
	hash := HashBlock(block)

	return sig.Verify(publicKey, hash)
}

func SignBlock(privateKey *crypto.PrivateKey, block *blockchain.Block) *crypto.Signature {
	hash := HashBlock(block)
	signature := privateKey.Sign(hash)
	block.PublicKey = privateKey.Public().Bytes()
	block.Signature = signature.Bytes()

	return signature
}

func HashBlock(block *blockchain.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(header *blockchain.Header) []byte {
	b, err := proto.Marshal(header)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)
	return hash[:]
}
