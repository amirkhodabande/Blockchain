package types

import (
	"crypto/sha256"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"google.golang.org/protobuf/proto"
)

func SignTransaction(privatekey *crypto.PrivateKey, tx *blockchain.Transaction) *crypto.Signature {
	return privatekey.Sign(HashTransaction(tx))
}

func HashTransaction(tx *blockchain.Transaction) []byte {
	b, err := proto.Marshal(tx)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)
	return hash[:]
}

func VerifyTransaction(tx *blockchain.Transaction) bool {
	for _, input := range tx.Inputs {
		var (
			signature = crypto.SignatureFromBytes(input.Signature)
			publicKey = crypto.PublicKeyFromBytes(input.PublicKey)
		)

		// TODO: check issues of nil signature
		tmpSignature := input.Signature
		input.Signature = nil

		if !signature.Verify(publicKey, HashTransaction(tx)) {
			return false
		}
		input.Signature = tmpSignature
	}
	return true
}
