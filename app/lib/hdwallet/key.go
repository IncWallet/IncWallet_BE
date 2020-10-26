package hdwallet

import (
	"encoding/hex"
	"encoding/json"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"incwallet/app/lib/crypto"
)

// 32-byte spending key
type PrivateKey []byte

// 32-byte public key
type PublicKey []byte

// 32-byte receiving key
type ReceivingKey []byte

// 32-byte transmission key
type TransmissionKey []byte

// ViewingKey is a public/private key pair to encrypt coins in an outgoing transaction
// and decrypt coins in an incoming transaction
type ViewingKey struct {
	Pk PublicKey    // 33 bytes, use to receive coin
	Rk ReceivingKey // 32 bytes, use to decrypt pointByte
}

func (viewKey ViewingKey) GetPublicSpend() *crypto.Point {
	pubSpend, _ := new(crypto.Point).FromBytesS(viewKey.Pk)
	return pubSpend
}

func (viewKey ViewingKey) GetPrivateView() *crypto.Scalar {
	return new(crypto.Scalar).FromBytesS(viewKey.Rk)
}

// PaymentAddress is an address of a payee
type PaymentAddress struct {
	Pk PublicKey       // 33 bytes, use to receive coin
	Tk TransmissionKey // 33 bytes, use to encrypt pointByte
}

func (addr PaymentAddress) GetPublicSpend() *crypto.Point {
	pubSpend, _ := new(crypto.Point).FromBytesS(addr.Pk)
	return pubSpend
}

func (addr PaymentAddress) GetPublicView() *crypto.Point {
	pubView, _ := new(crypto.Point).FromBytesS(addr.Tk)
	return pubView
}

// PaymentInfo contains an address of a payee and a value of coins he/she will receive
type PaymentInfo struct {
	PaymentAddress PaymentAddress
	Amount         uint64
	Message        []byte // 512 bytes
}

func InitPaymentInfo(addr PaymentAddress, amount uint64, message []byte) *PaymentInfo {
	return &PaymentInfo{
		PaymentAddress: addr,
		Amount: amount,
		Message: message,
	}
}

// GeneratePrivateKey generates a random 32-byte spending key
func GeneratePrivateKey(seed []byte) PrivateKey {
	bip32PrivKey := crypto.HashToScalar(seed)
	privateKey := bip32PrivKey.ToBytesS()
	return privateKey
}

// GeneratePublicKey computes a 32-byte public-key corresponding to a spending key
func GeneratePublicKey(privateKey []byte) PublicKey {
	privScalar := new(crypto.Scalar).FromBytesS(privateKey)
	publicKey := new(crypto.Point).ScalarMultBase(privScalar)
	return publicKey.ToBytesS()
}

// GenerateReceivingKey generates a 32-byte receiving key
func GenerateReceivingKey(privateKey []byte) ReceivingKey {
	receivingKey := crypto.HashToScalar(privateKey[:])
	return receivingKey.ToBytesS()
}

// GenerateTransmissionKey computes a 33-byte transmission key corresponding to a receiving key
func GenerateTransmissionKey(receivingKey []byte) TransmissionKey {
	receiScalar := new(crypto.Scalar).FromBytesS(receivingKey)
	transmissionKey := new(crypto.Point).ScalarMultBase(receiScalar)
	return transmissionKey.ToBytesS()
}

// GenerateViewingKey generates a viewingKey corresponding to a spending key
func GenerateViewingKey(privateKey []byte) ViewingKey {
	var viewingKey ViewingKey
	viewingKey.Pk = GeneratePublicKey(privateKey)
	viewingKey.Rk = GenerateReceivingKey(privateKey)
	return viewingKey
}

// GeneratePaymentAddress generates a payment address corresponding to a spending key
func GeneratePaymentAddress(privateKey []byte) PaymentAddress {
	var paymentAddress PaymentAddress
	paymentAddress.Pk = GeneratePublicKey(privateKey)
	paymentAddress.Tk = GenerateTransmissionKey(GenerateReceivingKey(privateKey))
	return paymentAddress
}

// Payment address funtions

func (addr *PaymentAddress) Bytes() []byte {
	return append(addr.Pk[:], addr.Tk[:]...)
}

func (addr *PaymentAddress) SetBytes(bytes []byte) *PaymentAddress {
	// the first 33 bytes are public key
	addr.Pk = bytes[:crypto.Ed25519KeySize]
	// the last 33 bytes are transmission key
	addr.Tk = bytes[crypto.Ed25519KeySize:]
	return addr
}

func (addr PaymentAddress) String() string {
	byteArrays := addr.Bytes()
	return hex.EncodeToString(byteArrays[:])
}

func (addr *PaymentAddress) GetLastByte() byte {
	return addr.Pk[len(addr.Pk) -1]
}

// Mining Key
func GenerateMiningKey(key PrivateKey) string {
	hash := common.HashB(common.HashB(key))
	return base58.Base58Check{}.Encode(hash, common.ZeroByte)
}

func GetMiningPubKey(validatorKey, paymentAddress string) (string, error) {
	committeePublicKey, err := GenerateKeySet(validatorKey, paymentAddress)
	if err != nil {
		return "", err
	}
	keyBytes, err := json.Marshal(committeePublicKey)
	if err != nil {
		return "", err
	}
	base58PubKey := base58.Base58Check{}.Encode(keyBytes, byte(0x00))
	return base58PubKey, nil
}