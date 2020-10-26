package hdwallet

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"incwallet/app/lib/common"
	"incwallet/app/models"

	"golang.org/x/crypto/scrypt"
	"time"
)

const (
	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on chain modern processor.
	StandardScryptN = 1 << 18

	// StandardScryptP is the P parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on chain modern processor.
	StandardScryptP = 1

	scryptR      = 8
	scryptKeyLen = 32

	aesMode    = "aes-256-gcm"
	scryptName = "scrypt"
)

type SafeStore struct {
}

func parseJson(mk *models.Wallet) (cipherData, nonce, salt []byte, err error) {
	if mk.Version != common.CryptoStoreVersion {
		return  nil, nil, nil, fmt.Errorf("version number error : %v", mk.Version)
	}

	// parse and check  cryptoJSON params
	if mk.Crypto.CipherName != aesMode {
		return  nil, nil, nil, fmt.Errorf("cipherName  error : %v", mk.Crypto.CipherName)
	}
	if mk.Crypto.KDF != scryptName {
		return  nil, nil, nil, fmt.Errorf("scryptName  error : %v", mk.Crypto.KDF)
	}
	cipherData, err = hex.DecodeString(mk.Crypto.CipherText)
	if err != nil {
		return nil, nil, nil, err
	}
	nonce, err = hex.DecodeString(mk.Crypto.Nonce)
	if err != nil {
		return nil, nil, nil, err
	}

	// parse and check  scryptParams params
	scryptParams := mk.Crypto.ScryptParams
	salt, err = hex.DecodeString(scryptParams.Salt)
	if err != nil {
		return  nil, nil, nil, err
	}

	return cipherData, nonce, salt, nil
}

func (ss SafeStore) EncryptPrivateKey(privateKey []byte, passphrase string) (*models.CryptoJSON, error) {
	n := StandardScryptN
	p := StandardScryptP
	pwdArray := []byte(passphrase)
	salt := common.GetEntropyCSPRNG(32)
	derivedKey, err := scrypt.Key(pwdArray, salt, n, scryptR, p, scryptKeyLen)
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:32]

	ciphertext, nonce, err := common.AesGCMEncrypt(encryptKey, privateKey)
	if err != nil {
		return nil, err
	}

	scryptParams := models.ScryptParams{
		N:      n,
		R:      scryptR,
		P:      p,
		KeyLen: scryptKeyLen,
		Salt:   hex.EncodeToString(salt),
	}

	cryptoJSON := models.CryptoJSON{
		CipherName:   aesMode,
		CipherText:   hex.EncodeToString(ciphertext),
		Nonce:        hex.EncodeToString(nonce),
		KDF:          scryptName,
		ScryptParams: scryptParams,
	}

	return &cryptoJSON, nil
}

func (ss SafeStore) DecryptPrivateKey(cipherJson *models.CryptoJSON, passphrase string) ([]byte, error) {
	cipherData, err := hex.DecodeString(cipherJson.CipherText)
	if err != nil {
		return nil, err
	}
	nonce, err := hex.DecodeString(cipherJson.Nonce)
	if err != nil {
		return nil, err
	}
	scryptParams := cipherJson.ScryptParams
	salt, err := hex.DecodeString(scryptParams.Salt)
	if err != nil {
		return nil, err
	}

	// begin decrypt
	derivedKey, err := scrypt.Key([]byte(passphrase), salt, scryptParams.N, scryptParams.R, scryptParams.P, scryptParams.KeyLen)
	if err != nil {
		return nil, err
	}

	privateKeyBytes, err := common.AesGCMDecrypt(derivedKey[:32], cipherData, []byte(nonce))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot decrpyt private key. Error %v", err))
	}
	return privateKeyBytes, nil
}

func (ss SafeStore) EncryptMasterKey(masterKeyBytes []byte, passphrase string, network string) (*models.Wallet, error) {
	n := StandardScryptN
	p := StandardScryptP
	pwdArray := []byte(passphrase)
	salt := common.GetEntropyCSPRNG(32)
	derivedKey, err := scrypt.Key(pwdArray, salt, n, scryptR, p, scryptKeyLen)
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:32]

	ciphertext, nonce, err := common.AesGCMEncrypt(encryptKey, masterKeyBytes)
	if err != nil {
		return nil, err
	}

	scryptParams := models.ScryptParams{
		N:      n,
		R:      scryptR,
		P:      p,
		KeyLen: scryptKeyLen,
		Salt:   hex.EncodeToString(salt),
	}

	cryptoJSON := models.CryptoJSON{
		CipherName:   aesMode,
		CipherText:   hex.EncodeToString(ciphertext),
		Nonce:        hex.EncodeToString(nonce),
		KDF:          scryptName,
		ScryptParams: scryptParams,
	}

	masterKeyJSON := models.Wallet{
		WalletId:  common.ParseNameFromHash(append(masterKeyBytes, []byte(network)...)),
		ShardID:   int(common.GetShardIDFromPublicKey(common.HashB(masterKeyBytes))),
		Crypto:    cryptoJSON,
		Version:   common.CryptoStoreVersion,
		Network:   network,
		Timestamp: time.Now().UTC().Unix(),
	}

	return &masterKeyJSON, nil
}

func (ss SafeStore) DecryptMasterKey(mk *models.Wallet, passphrase string) (*Key, error) {
	cipherData, nonce, salt, err := parseJson(mk)
	if err != nil {
		return nil, err
	}
	scryptParams := mk.Crypto.ScryptParams

	// begin decrypt
	derivedKey, err := scrypt.Key([]byte(passphrase), salt, scryptParams.N, scryptParams.R, scryptParams.P, scryptParams.KeyLen)
	if err != nil {
		return nil, err
	}

	masterKeyByte, err := common.AesGCMDecrypt(derivedKey[:32], cipherData, []byte(nonce))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot decrypt. Error %v", err))
	}

	masterKey := new(Key)
	if err := json.Unmarshal(masterKeyByte, masterKey); err != nil {
		return nil, errors.New("Cannot parse Master Key")
	}
	return masterKey, nil
}