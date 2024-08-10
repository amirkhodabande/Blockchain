package util

import (
	cryptoRand "crypto/rand"
	"io"
	"math/rand"
	"time"

	blockchain "github.com/blockchain/proto"
)

func RandomHash() []byte {
	hash := make([]byte, 32)
	io.ReadFull(cryptoRand.Reader, hash)

	return hash
}

func RandomBlock() *blockchain.Block {
	header := &blockchain.Header{
		Version:      1,
		Height:       int32(rand.Intn(1000)),
		PreviousHash: RandomHash(),
		RootHash:     RandomHash(),
		Timestamp:    time.Now().UnixNano(),
	}

	return &blockchain.Block{
		Header: header,
	}
}
