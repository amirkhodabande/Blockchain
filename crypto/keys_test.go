package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privateKey := GeneratePrivateKey()

	assert.Equal(t, len(privateKey.Bytes()), privateKeyLen)

	publicKey := privateKey.Public()

	assert.Equal(t, len(publicKey.Bytes()), publicKeyLen)
}

func TestNewPrivateKeyFromString(t *testing.T) {
	var (
		stringKey     = "9cc4f38df849cf7144e33fd8f8a53962eb00038333f6adaca9d0c37be693530c"
		privateKey    = NewPrivateKeyFromString(stringKey)
		addressString = "532714319995af6cac8fdcb39060a6ba0019f603"
	)
	assert.Equal(t, privateKeyLen, len(privateKey.Bytes()))
	address := privateKey.Public().Address()
	assert.Equal(t, addressString, address.String())
}

func TestPrivateKeySign(t *testing.T) {
	privateKey := GeneratePrivateKey()
	publicKey := privateKey.Public()
	msg := []byte("signed message")

	sig := privateKey.Sign(msg)
	assert.True(t, sig.Verify(publicKey, msg))

	assert.False(t, sig.Verify(publicKey, []byte("not signed message")))

	anotherPrivateKey := GeneratePrivateKey()
	anotherPublicKey := anotherPrivateKey.Public()
	assert.False(t, sig.Verify(anotherPublicKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privateKey := GeneratePrivateKey()
	publicKey := privateKey.Public()
	address := publicKey.Address()

	assert.Equal(t, addressLen, len(address.Bytes()))
	fmt.Println(address)
}
