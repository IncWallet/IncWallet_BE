package hdwallet

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"incwallet/app/lib/common"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
	"testing"
)

func TestNewMasterKey(t *testing.T) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("entropy: ", hex.EncodeToString(entropy))

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("mnemonic: ", mnemonic)

	seed := bip39.NewSeed(mnemonic, defaultPassphrase)
	fmt.Println("seed hex: ", hex.EncodeToString(seed))

	key, e := NewMasterKey(seed)
	if e != nil {
		fmt.Println(err)
	}
	fmt.Println("primary key hex: ", hex.EncodeToString(key.Key))

	fmt.Println("Accounts:")
	for i := 0; i < 1; i++ {
		path := fmt.Sprintf(IncognitoAccountPathFormat, i)
		kw, e := DeriveForPath(path, seed, nil)
		if e != nil {
			fmt.Println(err)
		}
		kw.Println()

		index := uint32(0)
		kw, e = DeriveWithIndex(index, []byte{}, key)
		if e != nil {
			fmt.Println(err)
		}
		kw.Println()
	}
}

func TestSafeStore_EncryptMasterKey(t *testing.T) {
	for i:=0 ;i < 10; i++ {
		entropy, _ := bip39.NewEntropy(256)
		m, _ := bip39.NewMnemonic(entropy)
		passphrase := "123456"
		seed := bip39.NewSeed(m, passphrase)

		var masterKey *Key
		var err error
		var ss SafeStore

		if masterKey, err = NewMasterKey(seed); err != nil {
			fmt.Println(err)
		}

		masterKeyBytes, err := json.Marshal(masterKey)

		masterKeyJson, err := ss.EncryptMasterKey(masterKeyBytes, passphrase, "testnet")
		if err != nil {
			fmt.Println(err)
		}

		decryptedMasterKey, err := ss.DecryptMasterKey(masterKeyJson, passphrase)
		if err != nil {
			fmt.Println(err)
		}

		assert.Equal(t, masterKey.Key, decryptedMasterKey.Key)
		assert.Equal(t, masterKey.ChainCode, decryptedMasterKey.ChainCode)
	}
}

func TestNewMasterKey2(t *testing.T) {
	for loop :=0; loop < 10; loop ++ {
		entropy, err := bip39.NewEntropy(256)
		if err != nil {
			fmt.Println(err)
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("mnemonic: ", mnemonic)

		seed := bip39.NewSeed(mnemonic, defaultPassphrase)

		mk, err := NewMasterKey(seed)
		if err != nil {
			fmt.Println(err)
		}

		mkPrime, err := NewMasterKeyFromMnemonic(mnemonic, defaultPassphrase)
		if err != nil {
			fmt.Println(err)
		}
		assert.Equal(t, mk.Key, mkPrime.Key)
		assert.Equal(t, mk.ChainCode, mkPrime.ChainCode)
	}
}

func TestSafeStore_EncryptPrivateKey(t *testing.T) {
	ss := new(SafeStore)
	for i:=0 ;i < 10; i++ {
		entropy, _ := bip39.NewEntropy(256)
		mnemonic, _ := bip39.NewMnemonic(entropy)
		seed := bip39.NewSeed(mnemonic, defaultPassphrase)
		key, _ := NewMasterKey(seed)
		//generate account
		index := uint32(1)
		keyWallet, _ := DeriveWithIndex(index, []byte{}, key)
		privateKeyStr := keyWallet.Base58CheckSerialize(common.PriKeyType)
		cryptoJson, err := ss.EncryptPrivateKey([]byte(privateKeyStr), "123456")
		if err != nil {
			fmt.Println(err)
		}
		privateKeyBytes, err := ss.DecryptPrivateKey(cryptoJson, "123456")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Private Key 1:", privateKeyStr)
		fmt.Println("Private Key 2:", string(privateKeyBytes))
	}
}