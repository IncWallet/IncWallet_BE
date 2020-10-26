package common

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/incognitochain/incognito-chain/privacy"
	"golang.org/x/crypto/sha3"
	"incwallet/app/lib/base58"
	"io"
)

type Hash [HashSize]byte

func GetNetworkURL(network string) string {
	if network == Testnet {
		return URLTestnet
	}
	if network == Mainnet {
		return URLMainnet
	}
	return ""
}

func GetShardIDFromPublicKey(pk []byte) byte {
	lastByte := pk[len(pk)-1]
	return byte(int(lastByte) % MaxShardNumber)
}

func Uint32ToBytes(value uint32) []byte {
	b := make([]byte, Uint32Size)
	binary.BigEndian.PutUint32(b, value)
	return b
}

func BytesToUint32(b []byte) (uint32, error) {
	if len(b) != Uint32Size {
		return 0, errors.New("invalid length of input BytesToUint32")
	}
	return binary.BigEndian.Uint32(b), nil
}

// HashB calculates SHA3-256 hashing of input b
// and returns the result in bytes array.
func HashB(b []byte) []byte {
	hash := sha3.Sum256(b)
	return hash[:]
}

// HashB calculates SHA3-256 hashing of input b
// and returns the result in Hash.
func HashH(b []byte) Hash {
	return Hash(sha3.Sum256(b))
}

func ParseNameFromHash(b []byte) string {
	hash := sha3.Sum256(b)
	hash = sha3.Sum256(hash[:])
	return base58.Base58Check{}.Encode(hash[:], ZeroByte)
}

func GetEntropyCSPRNG(n int) []byte {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(crand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return mainBuff
}

func AesCTRXOR(key, inText, iv []byte) ([]byte, error) {

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

func AesGCMEncrypt(key, inText []byte) (outText, nonce []byte, err error) {

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	stream, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, nil, err
	}

	nonce = GetEntropyCSPRNG(12)

	outText = stream.Seal(nil, nonce, inText, []byte(GcmAdditionData))
	return outText, nonce, err
}

func AesGCMDecrypt(key, cipherText, nonce []byte) ([]byte, error) {

	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	outText, err := stream.Open(nil, nonce, cipherText, []byte(GcmAdditionData))
	if err != nil {
		return nil, err
	}

	return outText, err
}

func ParseString2Point(data string) (*privacy.Point, error) {
	pByte, _, err := base58.Base58Check{}.Decode(data)
	if err != nil {
		return nil, err
	}
	p, err := new(privacy.Point).FromBytesS(pByte)
	if err != nil {
		return nil, err
	}
	return p, nil
}
