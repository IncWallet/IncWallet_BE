package hdwallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"regexp"
	"strconv"
	"strings"
)

const (
	IncognitoAccountPrefix      = "m/44'/1750'"
	IncognitoPrimaryAccountPath = "m/44'/1750'/0'"
	IncognitoAccountPathFormat  = "m/44'/1750'/%d'"
	FirstHardenedIndex          = 1 << 31 // bip 44, hardened child key mast begin with 2^32
	seedModifier                = "incognito modifier seed"
	defaultPassphrase           = "incognito passphrase"
)

var (
	ErrInvalidPath        = errors.New("invalid derivation path")
	ErrNoPublicDerivation = errors.New("no public derivation for ed25519")

	pathRegex = regexp.MustCompile("^m(\\/[0-9]+')+$")
)

type Key struct {
	Key       []byte
	ChainCode []byte
}

type KeySet struct {
	PrivateKey     PrivateKey
	PaymentAddress PaymentAddress
	ReadonlyKey    ViewingKey
}

type KeyWallet struct {
	Depth       byte
	ChildNumber []byte
	ChainCode   []byte
	Key			[]byte
	KeySet      KeySet
}

// DeriveForPath derives key for chain path in BIP-44 format and chain seed.
// Ed25119 derivation operated on hardened keys only.
func DeriveForPath(path string, seed []byte, masterKey *Key) (*KeyWallet, error) {
	if !isValidPath(path) {
		return nil, ErrInvalidPath
	}

	if masterKey == nil {
		var err error
		if masterKey, err = NewMasterKey(seed); err != nil {
			return nil, err
		}
	}

	keyWallet := new(KeyWallet)
	segments := strings.Split(path, "/")
	for _, segment := range segments[1:] {
		i64, err := strconv.ParseUint(strings.TrimRight(segment, "'"), 10, 32)
		if err != nil {
			return nil, err
		}

		i := uint32(i64) + FirstHardenedIndex
		key, err := masterKey.Derive(i)
		if err != nil {
			return nil, err
		}
		keyWallet.ChildNumber = common.Uint32ToBytes(uint32(i64))
		keyWallet.Depth = 1
		keyWallet.Key = key.Key
		keyWallet.ChainCode = key.ChainCode
		keyWallet.KeySet = *key.GenerateKeySet()
	}

	return keyWallet, nil
}

func DeriveWithIndex(i uint32, seed []byte, masterKey *Key) (*KeyWallet, error) {
	path := fmt.Sprintf(IncognitoAccountPathFormat, i)
	return DeriveForPath(path, seed, masterKey)
}

func NewMasterKey(seed []byte) (*Key, error) {
	hmac := hmac.New(sha512.New, []byte(seedModifier))
	_, err := hmac.Write(seed)
	if err != nil {
		return nil, err
	}
	sum := hmac.Sum(nil)
	key := &Key{
		Key:       sum[:32],
		ChainCode: sum[32:],
	}
	return key, nil
}

func NewMasterKeyFromMnemonic(mnemonic, passphrase string) (*Key, error) {
	seed := bip39.NewSeed(mnemonic, passphrase)
	return NewMasterKey(seed)
}

func (k *Key) Derive(i uint32) (*Key, error) {
	// no public derivation for ed25519
	if i < FirstHardenedIndex {
		return nil, ErrNoPublicDerivation
	}

	iBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(iBytes, i)
	key := append([]byte{0x0}, k.Key...)
	data := append(key, iBytes...)

	hmac := hmac.New(sha512.New, k.ChainCode)
	_, err := hmac.Write(data)
	if err != nil {
		return nil, err
	}
	sum := hmac.Sum(nil)
	newKey := &Key{
		Key:       sum[:32],
		ChainCode: sum[32:],
	}
	return newKey, nil
}

func (k *Key) GenerateKeySet() (*KeySet) {
	keySet := new(KeySet)
	keySet.PrivateKey = GeneratePrivateKey(k.Key)
	keySet.PaymentAddress = GeneratePaymentAddress(keySet.PrivateKey[:])
	keySet.ReadonlyKey = GenerateViewingKey(keySet.PrivateKey[:])
	return keySet
}

func (k *Key) RawSeed() [32]byte {
	var rawSeed [32]byte
	copy(rawSeed[:], k.Key[:])
	return rawSeed
}

func (key *KeySet) InitFromPrivateKey(privateKey *PrivateKey) error {
	if privateKey == nil || len(*privateKey) != common.PrivateKeySize {
		return errors.New(fmt.Sprintf("Cannot init key wallet from private key"))
	}

	key.PrivateKey = *privateKey
	key.PaymentAddress = GeneratePaymentAddress(key.PrivateKey[:])
	key.ReadonlyKey = GenerateViewingKey(key.PrivateKey[:])

	return nil
}

func isValidPath(path string) bool {
	if !pathRegex.MatchString(path) {
		return false
	}

	// Check for overflows
	segments := strings.Split(path, "/")
	for _, segment := range segments[1:] {
		_, err := strconv.ParseUint(strings.TrimRight(segment, "'"), 10, 32)
		if err != nil {
			return false
		}
	}
	return true
}

func deserialize(data []byte) (*KeyWallet, error) {
	var key = &KeyWallet{}
	if len(data) < 2 {
		return nil, errors.New("Cannot deserialized data. Len error")
	}
	keyType := data[0]
	if keyType == common.PriKeyType {
		if len(data) != common.PrivKeySerializedBytesLen {
			return nil, errors.New("Cannot deserialized data. Private Key error")
		}

		key.Depth = data[1]
		key.ChildNumber = data[2:6]
		key.ChainCode = data[6:38]
		keyLength := int(data[38])
		key.KeySet.PrivateKey = make([]byte, keyLength)
		copy(key.KeySet.PrivateKey[:], data[39:39+keyLength])
	} else if keyType == common.PaymentAddressType {
		if !bytes.Equal(common.BurnAddress1BytesDecode, data) {
			if len(data) != common.PaymentAddrSerializedBytesLen {
				return nil, errors.New("Cannot deserialized data. Payment Address error")
			}
		}
		apkKeyLength := int(data[1])
		pkencKeyLength := int(data[apkKeyLength+2])
		key.KeySet.PaymentAddress.Pk = make([]byte, apkKeyLength)
		key.KeySet.PaymentAddress.Tk = make([]byte, pkencKeyLength)
		copy(key.KeySet.PaymentAddress.Pk[:], data[2:2+apkKeyLength])
		copy(key.KeySet.PaymentAddress.Tk[:], data[3+apkKeyLength:3+apkKeyLength+pkencKeyLength])
	} else if keyType == common.ReadonlyKeyType {
		if len(data) != common.ReadOnlyKeySerializedBytesLen {
			return nil, errors.New("Cannot deserialized data. ReadOnlyKey error")
		}

		apkKeyLength := int(data[1])
		if len(data) < apkKeyLength+3 {
			return nil, errors.New("Cannot deserialized data. Unkown error")
		}
		skencKeyLength := int(data[apkKeyLength+2])
		key.KeySet.ReadonlyKey.Pk = make([]byte, apkKeyLength)
		key.KeySet.ReadonlyKey.Rk = make([]byte, skencKeyLength)
		copy(key.KeySet.ReadonlyKey.Pk[:], data[2:2+apkKeyLength])
		copy(key.KeySet.ReadonlyKey.Rk[:], data[3+apkKeyLength:3+apkKeyLength+skencKeyLength])
	}

	// validate checksum
	cs1 := base58.ChecksumFirst4Bytes(data[0 : len(data)-4])
	cs2 := data[len(data)-4:]
	for i := range cs1 {
		if cs1[i] != cs2[i] {
			return nil, errors.New("Cannot deserialized data. Checksum error")
		}
	}
	return key, nil
}

func Base58CheckDeserialize(data string) (*KeyWallet, error) {
	b, _, err := base58.Base58Check{}.Decode(data)
	if err != nil {
		return nil, err
	}
	return deserialize(b)
}

func (kw *KeyWallet) serialize(keyType byte) ([]byte, error) {
	// Write fields to buffer in order
	buffer := new(bytes.Buffer)
	buffer.WriteByte(keyType)
	if keyType == common.PriKeyType {
		buffer.WriteByte(kw.Depth)
		buffer.Write(kw.ChildNumber)
		buffer.Write(kw.ChainCode)

		// Private keys should be prepended with a single null byte
		keyBytes := make([]byte, 0)
		keyBytes = append(keyBytes, byte(len(kw.KeySet.PrivateKey))) // set length
		keyBytes = append(keyBytes, kw.KeySet.PrivateKey[:]...)      // set pri-key
		buffer.Write(keyBytes)
	} else if keyType == common.PaymentAddressType {
		keyBytes := make([]byte, 0)
		keyBytes = append(keyBytes, byte(len(kw.KeySet.PaymentAddress.Pk))) // set length PaymentAddress
		keyBytes = append(keyBytes, kw.KeySet.PaymentAddress.Pk[:]...)      // set PaymentAddress

		keyBytes = append(keyBytes, byte(len(kw.KeySet.PaymentAddress.Tk))) // set length Pkenc
		keyBytes = append(keyBytes, kw.KeySet.PaymentAddress.Tk[:]...)      // set Pkenc
		buffer.Write(keyBytes)
	} else if keyType == common.ReadonlyKeyType {
		keyBytes := make([]byte, 0)
		keyBytes = append(keyBytes, byte(len(kw.KeySet.ReadonlyKey.Pk))) // set length PaymentAddress
		keyBytes = append(keyBytes, kw.KeySet.ReadonlyKey.Pk[:]...)      // set PaymentAddress

		keyBytes = append(keyBytes, byte(len(kw.KeySet.ReadonlyKey.Rk))) // set length Skenc
		keyBytes = append(keyBytes, kw.KeySet.ReadonlyKey.Rk[:]...)      // set Pkenc
		buffer.Write(keyBytes)
	} else {
		return []byte{}, errors.New("Cannot serialized KeyWallet. Type is not supported")
	}

	// Append the standard doublesha256 checksum
	checksum := base58.ChecksumFirst4Bytes(buffer.Bytes())
	return append(buffer.Bytes(), checksum...), nil
}

func (kw *KeyWallet) Base58CheckSerialize(keyType byte) string {
	serializedKey, err := kw.serialize(keyType)
	if err != nil {
		return ""
	}
	return base58.Base58Check{}.Encode(serializedKey, byte(0x00))
}

func GetKeyWalletInfoFromPrivateKey(privateKeyStr string) (*KeyWallet, string, string, string, error) {
	keyWallet, err := Base58CheckDeserialize(privateKeyStr)
	if err != nil {
		return nil, "", "", "", errors.New(fmt.Sprintf("Cannot init key wallet. Error %v", err))
	}
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	publicKeyStr := base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, common.ZeroByte)
	paymentAddressStr := keyWallet.Base58CheckSerialize(common.PaymentAddressType)
	viewingKeyStr := keyWallet.Base58CheckSerialize(common.ReadonlyKeyType)
	return keyWallet, publicKeyStr, paymentAddressStr, viewingKeyStr, nil
}

func (kw *KeyWallet) Println() {
	fmt.Println("=======")
	fmt.Println("Child Number", kw.ChildNumber)
	fmt.Println("Chain Code", base58.Base58Check{}.Encode(kw.ChainCode, common.ZeroByte))
	fmt.Println("Key ", base58.Base58Check{}.Encode(kw.Key, common.ZeroByte))
	fmt.Println("KeySet Detail:")
	fmt.Println("--private key:", kw.Base58CheckSerialize(common.PriKeyType))
	fmt.Println("--payment address:", kw.Base58CheckSerialize(common.PaymentAddressType))
	fmt.Println("--shard id: ", common.GetShardIDFromPublicKey(kw.KeySet.PaymentAddress.Pk))
	fmt.Println("--mining key:", GenerateMiningKey(kw.KeySet.PrivateKey) )
}
