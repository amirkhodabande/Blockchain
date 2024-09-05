package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratePrivateKey(t *testing.T) {
	privateKey := GeneratePrivateKey()

	require.Equal(t, len(privateKey.Bytes()), privateKeyLen)

	publicKey := privateKey.Public()

	require.Equal(t, len(publicKey.Bytes()), PublicKeyLen)
}

func TestNewPrivateKeyFromString(t *testing.T) {
	var (
		stringKey     = "9cc4f38df849cf7144e33fd8f8a53962eb00038333f6adaca9d0c37be693530c"
		privateKey    = NewPrivateKeyFromString(stringKey)
		addressString = "532714319995af6cac8fdcb39060a6ba0019f603"
	)
	require.Equal(t, privateKeyLen, len(privateKey.Bytes()))
	address := privateKey.Public().Address()
	require.Equal(t, addressString, address.String())
}

func TestPrivateKeySign(t *testing.T) {
	privateKey := GeneratePrivateKey()
	publicKey := privateKey.Public()
	msg := []byte("signed message")

	sig := privateKey.Sign(msg)
	require.True(t, sig.Verify(publicKey, msg))

	require.False(t, sig.Verify(publicKey, []byte("not signed message")))

	anotherPrivateKey := GeneratePrivateKey()
	anotherPublicKey := anotherPrivateKey.Public()
	require.False(t, sig.Verify(anotherPublicKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privateKey := GeneratePrivateKey()
	publicKey := privateKey.Public()
	address := publicKey.Address()

	require.Equal(t, addressLen, len(address.Bytes()))
	fmt.Println(address)
}
