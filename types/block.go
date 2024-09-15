package types

import (
	"bytes"
	"crypto/sha256"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"github.com/cbergoon/merkletree"
	"google.golang.org/protobuf/proto"
)

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{hash: hash}
}

func (txHash TxHash) CalculateHash() ([]byte, error) {
	return txHash.hash, nil
}

func (txHash TxHash) Equals(other merkletree.Content) (bool, error) {
	return bytes.Equal(txHash.hash, other.(TxHash).hash), nil
}

func VerifyBlock(block *blockchain.Block) bool {
	if len(block.Transactions) > 0 {
		if !verifyRootHash(block) {
			return false
		}
	}

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
	if len(block.Transactions) > 0 {
		tree, err := getMerkleTree(block)
		if err != nil {
			panic(err)
		}

		block.Header.RootHash = tree.MerkleRoot()
	}

	hash := HashBlock(block)
	signature := privateKey.Sign(hash)
	block.PublicKey = privateKey.Public().Bytes()
	block.Signature = signature.Bytes()

	return signature
}

func verifyRootHash(block *blockchain.Block) bool {
	tree, err := getMerkleTree(block)
	if err != nil {
		return false
	}

	valid, err := tree.VerifyTree()
	if err != nil {
		return false
	}

	if !valid {
		return false
	}

	return bytes.Equal(block.Header.RootHash, tree.MerkleRoot())
}

func getMerkleTree(block *blockchain.Block) (*merkletree.MerkleTree, error) {
	list := make([]merkletree.Content, len(block.Transactions))

	for index, transaction := range block.Transactions {
		list[index] = NewTxHash(HashTransaction(transaction))
	}

	tree, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}

	return tree, nil
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
