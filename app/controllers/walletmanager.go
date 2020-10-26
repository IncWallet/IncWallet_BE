package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/common"
	"incwallet/app/lib/hdwallet"
	"incwallet/app/models"
)

type WalletManager struct {
	WalletID  string
	SafeStore hdwallet.SafeStore
	Wallet    *models.Wallet
}

func (wm *WalletManager) Init(walletId string) error {
	wallet := &models.Wallet{}
	if err := database.Wallet.Find(bson.M{"walletid": walletId}).One(&wallet); err != nil {
		return errors.New(fmt.Sprintf("Cannot find wallet ID %v in database from Init WM", walletId))
	}
	wm.WalletID = wallet.WalletId
	wm.Wallet = wallet
	return nil
}

func (wm *WalletManager) CreateNewWallet(security int, passphrase, network string) (string, error) {
	if security != 128 && security != 256 {
		return "", errors.New("[WM] Cannot init from invalid security level")
	}
	if network != common.Testnet && network != common.Mainnet && network != common.Local {
		return "", errors.New("[WM] Cannot init from invalid network info (local, testnet or mainnet)")
	}
	if count, err := database.Wallet.Find(bson.M{}).Count(); err == nil && count > 0 {
		fmt.Println(count, err)
		return "", errors.New(fmt.Sprintf("[WM] Cannot create more than one wallet"))
	}

	entropy, err := bip39.NewEntropy(security)
	if err != nil {
		return "", errors.New("[WM] Cannot new entropy")
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", errors.New("[WM] Cannot new mnemonic")
	}
	seed := bip39.NewSeed(mnemonic, passphrase)
	mk, err := hdwallet.NewMasterKey(seed)
	if err != nil {
		return "", errors.New("[WM] Cannot new master key")
	}

	if mkByte, err := json.Marshal(mk); err != nil {
		return "", errors.New("[WM] Cannot create ciphertext for master key. Marshal error")
	} else {
		mkJson, err1 := wm.SafeStore.EncryptMasterKey(mkByte, passphrase, network)
		if err1 != nil {
			return "", errors.New("[WM] Cannot create ciphertext for master key. Encrypted error")
		}
		if err2 := database.Wallet.Insert(mkJson); err2 != nil {
			return "", errors.New("[WM] Cannot store Wallet to database. Insert error")
		}
		wm.Wallet = mkJson
		wm.WalletID = mkJson.WalletId
		StateM.NetworkManager.Init(network, "")
		StateM.RpcCaller.Init(network)
		if err := StateM.SaveState(); err != nil {
			return "", errors.New(fmt.Sprintf("Cannnot save State from Create wallet. Error %v",err))
		}
	}
	return mnemonic, nil
}

func (wm *WalletManager) ImportWallet(mnemonic, passphrase, network string) error {
	if network != common.Testnet && network != common.Mainnet && network != common.Local {
		return errors.New("[WM] Cannot init from invalid network info (local, testnet or mainnet)")
	}
	if count, err := database.Wallet.Find(bson.M{}).Count(); err == nil && count > 0 {
		return errors.New(fmt.Sprintf("[WM] Cannot create more than one wallet"))
	}

	seed := bip39.NewSeed(mnemonic, passphrase)
	mk, err := hdwallet.NewMasterKey(seed)
	if err != nil {
		return errors.New("[WM] Cannot new master key")
	}

	if mkByte, err := json.Marshal(mk); err != nil {
		return errors.New("[WM] Cannot create ciphertext for master key. Marshal error")
	} else {
		mkJson, err1 := wm.SafeStore.EncryptMasterKey(mkByte, passphrase, network)
		if err1 != nil {
			return errors.New("[WM] Cannot create ciphertext for master key. Encrypted error")
		}
		if err2 := database.Wallet.Insert(mkJson); err2 != nil {
			return errors.New("[WM] Cannot store Wallet to database. Insert error")
		}
		wm.Wallet = mkJson
		wm.WalletID = mkJson.WalletId
		StateM.NetworkManager.Init(network, "")
		StateM.RpcCaller.Init(network)
		if err := StateM.SaveState(); err != nil {
			return errors.New("Cannnot update State from Importwallet")
		}
	}
	return nil
}