package types

import (
	"fmt"
	"testing"

	"github.com/blockchain/crypto"
	blockchain "github.com/blockchain/proto"
	"github.com/blockchain/util"
	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	var (
		fromPrivateKey = crypto.GeneratePrivateKey()
		fromAddress    = fromPrivateKey.Public().Address().Bytes()

		toPrivateKey = crypto.GeneratePrivateKey()
		toAddress    = toPrivateKey.Public().Address().Bytes()
	)

	input := &blockchain.TxInput{
		PreviousTxHash:   util.RandomHash(),
		PreviousOutIndex: 0,
		PublicKey:        fromPrivateKey.Public().Bytes(),
	}

	output1 := &blockchain.TxOutput{
		Amount:  4,
		Address: toAddress,
	}
	output2 := &blockchain.TxOutput{
		Amount:  4,
		Address: fromAddress,
	}

	tx := &blockchain.Transaction{
		Version: 1,
		Inputs:  []*blockchain.TxInput{input},
		Outputs: []*blockchain.TxOutput{output1, output2},
	}
	sig := SignTransaction(fromPrivateKey, tx)
	input.Signature = sig.Bytes()

	assert.True(t, VerifyTransaction(tx))
	fmt.Println(input.Signature)
}
