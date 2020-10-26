package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"incwallet/app/lib/crypto"
	"incwallet/app/lib/hdwallet"
	"incwallet/app/models"
	"sort"
	"strconv"
)

type AccountManager struct {
	AccountID string
	Account   *models.Account
}

func getBalanceByTokenID(publickey, tokenID string) (uint64, error) {
	coins := make([]models.Coins, 0)
	if err := database.Coins.Find(bson.M{"publickey": publickey,
		"tokenid": tokenID,
		"isspent": false}).All(&coins); err != nil {
		return 0, errors.New(fmt.Sprintf("Cannot get init wallet. Error: %v", err))
	}

	balance := uint64(0)
	for _, coin := range coins {
		if value, err := strconv.ParseUint(coin.Value, 10, 64); err != nil {
			return 0, errors.New("Cannot parse coin value")
		} else {
			balance += value
		}
	}
	return balance, nil
}

func getTotalValueInPRV(mapBalance map[string]uint64) (uint64, error) {
	if len(mapBalance) == 0 {
		return 0, nil
	}

	totalPRV := uint64(0)
	for tokenID, balance := range mapBalance {
		if tokenID == common.PRVID {
			totalPRV += balance
			continue
		}
		tmpPrvAmount, _, err := getExchangeRate(tokenID, common.PRVID, balance, 0, true)
		if err != nil {
			revel.AppLog.Warnf("cannot get exchange amount for pair PRV-%v. Error %v", tokenID, err)
			continue
		}
		totalPRV += tmpPrvAmount
	}
	return totalPRV, nil
}

func (am *AccountManager) Init(accountID string) error {
	account := &models.Account{}
	if err := database.Accounts.Find(bson.M{"publickey": accountID}).One(&account); err != nil {
		return errors.New(fmt.Sprintf("Cannot find account ID %v in database from Init AM", accountID))
	}
	am.AccountID = account.PublicKey
	am.Account = account
	return nil
}

func (am *AccountManager) ImportAccount(name, privateKeyStr, passphrase string) error {
	accountKW, err := hdwallet.Base58CheckDeserialize(privateKeyStr)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot init key wallet. Error %v", err))
	}
	accountKW.KeySet.InitFromPrivateKey(&accountKW.KeySet.PrivateKey)

	privateKeyJson, err := StateM.WalletManager.SafeStore.EncryptPrivateKey([]byte(privateKeyStr), passphrase)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot encrypt private key. Error %v", err))
	}

	newAccount := &models.Account{
		Name:           name,
		Index:          0,
		PublicKey:      base58.Base58Check{}.Encode(accountKW.KeySet.PaymentAddress.Pk, common.ZeroByte),
		PaymentAddress: accountKW.Base58CheckSerialize(common.PaymentAddressType),
		ViewKey:        accountKW.Base58CheckSerialize(common.ReadonlyKeyType),
		MiningKey:      hdwallet.GenerateMiningKey(accountKW.KeySet.PrivateKey),
		Wallet:         StateM.WalletManager.WalletID,
		Crypto:         *privateKeyJson,
	}

	tmpAccount := new(models.Account)
	if err := database.Accounts.Find(bson.M{"publickey":newAccount.PublicKey}).One(&tmpAccount); err == nil {
		return errors.New("duplicate account in database")
	}

	if err = database.Accounts.Insert(newAccount); err != nil {
		return err
	}

	newAddressBook := &models.AddressBook{
		Name:           name,
		PaymentAddress: newAccount.PaymentAddress,
		ChainName:      "Incognito Chain",
		ChainType:      StateM.NetworkManager.NetworkName,
	}
	if err = database.AddressBook.Insert(newAddressBook); err !=nil {
		revel.AppLog.Warnf("cannot add new info to address book. Error %v", err)
	}

	am.Account = newAccount
	am.AccountID = newAccount.PublicKey
	if err := StateM.SaveState(); err != nil {
		return errors.New("cannnot update State from Addaccount")
	}
	return nil
}

func (am *AccountManager) AddAccount(name string, masterKey *hdwallet.Key, walletID string, shardID int, passphrase string) (string, error) {
	lastAccount := &models.Account{}
	query := bson.M{
		"wallet": walletID,
	}

	if count, err := database.Accounts.Find(bson.M{"name":name}).Count(); err == nil && count > 0 {
		revel.AppLog.Error("account name existed")
		return "", errors.New(fmt.Sprintf("account name %v existed", name))
	}
	startIndex := uint32(0)
	if err := database.Accounts.Find(query).Sort("-index").Limit(1).One(&lastAccount); err != nil {
		revel.AppLog.Warn("Account table is empty")
	}

	startIndex = lastAccount.Index

	for i := startIndex + 1; i < common.MaxIndex; i++ {
		accountKW, err := hdwallet.DeriveWithIndex(uint32(i), nil, masterKey)
		if err != nil {
			return "", err
		}
		publicKey := accountKW.KeySet.PaymentAddress.Pk
		if int(common.GetShardIDFromPublicKey(publicKey)) == shardID {
			privateKeyStr := accountKW.Base58CheckSerialize(common.PriKeyType)
			privateKeyJson, err := StateM.WalletManager.SafeStore.EncryptPrivateKey([]byte(privateKeyStr), passphrase)
			if err != nil {
				return "", errors.New(fmt.Sprintf("Cannot encrypt private key. Error %v", err))
			}
			newAccount := &models.Account{
				Name:           name,
				Index:          i,
				PublicKey:      base58.Base58Check{}.Encode(publicKey, common.ZeroByte),
				PaymentAddress: accountKW.Base58CheckSerialize(common.PaymentAddressType),
				ViewKey:        accountKW.Base58CheckSerialize(common.ReadonlyKeyType),
				MiningKey:      hdwallet.GenerateMiningKey(accountKW.KeySet.PrivateKey),
				Wallet:         walletID,
				Crypto:         *privateKeyJson,
			}
			if err = database.Accounts.Insert(newAccount); err != nil {
				return "", err
			}

			newAddressBook := &models.AddressBook{
				Name:           name,
				PaymentAddress: newAccount.PaymentAddress,
				ChainName:      "Incognito Chain",
				ChainType:      StateM.NetworkManager.NetworkName,
			}
			if err = database.AddressBook.Insert(newAddressBook); err !=nil {
				revel.AppLog.Warnf("cannot add new info to address book. Error %v", err)
			}

			am.Account = newAccount
			am.AccountID = newAccount.PublicKey
			if err := StateM.SaveState(); err != nil {
				return "", errors.New("Cannnot update State from Addaccount")
			}
			return privateKeyStr, nil
		}
	}
	return "", errors.New("[WM] Cannot add account. MaxIndex error")
}

func (am *AccountManager) SwitchAccount(name, passphrase string) (string, error) {
	acc := &models.Account{}
	if err := database.Accounts.Find(bson.M{"name": name}).Sort("-index").Limit(1).One(&acc); err != nil {
		revel.AppLog.Errorf("cannot load account %v. Error %v", name, err)
		return "", errors.New(fmt.Sprintf("cannot load account %v. Error %v", name, err))
	}

	privateKeyBytes, err := StateM.WalletManager.SafeStore.DecryptPrivateKey(&acc.Crypto, passphrase)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("Cannot decrypt privatekey. Error %v", err))
		return "", errors.New(fmt.Sprintf("Cannot decrypt privatekey. Error %v", err))
	}

	am.Account = acc
	am.AccountID = acc.PublicKey
	if err := StateM.SaveState(); err != nil {
		return "", errors.New("Cannnot update State from SwitchAccount")
	}
	return string(privateKeyBytes), nil
}

func (am *AccountManager) GetKeyWallet(passphrase string) (*hdwallet.KeyWallet, error) {
	walletEncrypted := new(models.Wallet)
	if err := database.Wallet.Find(bson.M{"walletid": StateM.WalletManager.WalletID}).One(&walletEncrypted); err != nil {
		return nil, errors.New("Cannot load Wallet from database in InitTransaction")
	}

	masterKey, err := StateM.WalletManager.SafeStore.DecryptMasterKey(walletEncrypted, passphrase)
	if err != nil {
		return nil, errors.New("Cannot decrypt Wallet from database in InitTransaction")
	}

	keyWallet, err := hdwallet.DeriveWithIndex(StateM.AccountManage.Account.Index, nil, masterKey)
	if err != nil {
		return nil, errors.New("Cannot regenerate keyset from masterkey in InitTransaction")
	}
	return keyWallet, nil
}

func (am *AccountManager) GetBalance(publicKey, tokenID string) (map[string]uint64, error) {
	if publicKey == "" {
		publicKey = StateM.AccountManage.AccountID
	}

	mapBalance := make(map[string]uint64)
	if len(tokenID) == 0 {
		listTokenID := make([]string, 0)
		if err := database.Coins.Find(bson.M{"publickey": publicKey}).Distinct("tokenid", &listTokenID); err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot get token id from wallet. Error: %v", err))
		}
		for _, id := range listTokenID {
			balance, err := getBalanceByTokenID(publicKey, id)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Cannot get balance by token id. Error: %v", err))
			}
			mapBalance[id] = balance
		}
	} else {
		balance, err := getBalanceByTokenID(publicKey, tokenID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cannot get balance by token id. Error: %v", err))
		}
		mapBalance[tokenID] = balance
	}

	return mapBalance, nil
}

func (am *AccountManager) SyncAccount(publicKey, passphrase string) error {
	if publicKey == "" {
		publicKey = StateM.AccountManage.AccountID
	}
	acc := &models.Account{}
	if err := database.Accounts.Find(bson.M{"publickey": publicKey}).One(&acc); err != nil {
		revel.AppLog.Error(fmt.Sprintf("cannot get current account. Error %v", err))
		return errors.New(fmt.Sprintf("cannot get current account. Error %v", err))
	}
	privateKeyBytes, err := StateM.WalletManager.SafeStore.DecryptPrivateKey(&acc.Crypto, passphrase)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("cannot decrypt privatekey. Error %v", err))
		return errors.New(fmt.Sprintf("cannot decrypt privatekey. Error %v", err))
	}
	if publicKey == StateM.AccountManage.AccountID || publicKey == ""{
		if err := JobSyncAccountFromRemote(string(privateKeyBytes)); err != nil {
			revel.AppLog.Error(fmt.Sprintf("cannot sync from actived account. Error %v", err))
			return errors.New(fmt.Sprintf("cannot sync from actived account. Error %v", err))
		}
	} else {
		synckerManager := InitSynckerManager(string(privateKeyBytes))
		if err := synckerManager.SyncAccountJob(); err != nil {
			revel.AppLog.Errorf("cannot sync from account. %v-%v. Error %v", acc.Name, acc.PaymentAddress, err)
			return errors.New(fmt.Sprintf("cannot sync from account. %v-%v. Error %v", acc.Name, acc.PaymentAddress, err))
		}
	}
	return nil
}

func (am *AccountManager) SyncAllAccounts(accounts []*models.Account, passphrase string) []error {
	listError := make([]error, 0)
	errChan := make(chan error)
	for _, account := range accounts {
		go func(acc *models.Account, errChan chan error) {
			privateKeyBytes, err := StateM.WalletManager.SafeStore.DecryptPrivateKey(&acc.Crypto, passphrase)
			if err != nil {
				revel.AppLog.Errorf("cannot decrypt private key for account %v -%v. Error %v", acc.Name, acc.PaymentAddress, err)
				errChan <- errors.New(fmt.Sprintf("cannot decrypt private key for account %v -%v. Error %v", acc.Name, acc.PaymentAddress, err))
				return
			}

			synckerManager := InitSynckerManager(string(privateKeyBytes))
			if err := synckerManager.SyncAccountJob(); err != nil {
				revel.AppLog.Errorf("cannot sync from account. %v-%v. Error %v", acc.Name, acc.PaymentAddress, err)
				errChan <- errors.New(fmt.Sprintf("cannot sync from account. %v-%v. Error %v", acc.Name, acc.PaymentAddress, err))
				return
			}

			errChan <- nil
		}(account, errChan)
	}

	for range accounts {
		err := <- errChan
		if err != nil {
			listError = append(listError, err)
		}
	}
	return listError
}

func (am *AccountManager) GetAllOutputCoinR(tokenID string) ([]*models.Coins, error) {
	if am == nil {
		return nil, errors.New("cannot get output coins. Account manager is nil")
	}
	keyWallet, err := hdwallet.Base58CheckDeserialize(am.Account.ViewKey)
	if err != nil {
		return nil, errors.New("cannot get viewing key from account view key")
	}
	byteData, err := StateM.RpcCaller.GetListOutputCoinsInBytes(am.Account.PaymentAddress, keyWallet.KeySet.ReadonlyKey.Rk[:], tokenID)

	if err != nil {
		return nil, errors.New("Cannot request list output coin")
	}
	var listCoins []*models.Coins
	if err := json.Unmarshal(byteData, &listCoins); err != nil {
		return nil, err
	}
	return listCoins, nil
}

func (am *AccountManager) GetUnspentOutputCoinR(keyWallet *hdwallet.KeyWallet, tokenID string) ([]*models.Coins, error) {
	listCoin, err := am.GetAllOutputCoinR(tokenID)
	if err != nil {
		return nil, err
	}

	tmpSerialNumbers := make([]string, 0)
	for index, coin := range listCoin {
		snd, _, _ := base58.Base58Check{}.Decode(coin.SNDerivator)
		sn := crypto.GenerateSerialNumber(keyWallet.KeySet.PrivateKey, snd)
		listCoin[index].SerialNumber = base58.Base58Check{}.Encode(sn, common.ZeroByte)
		tmpSerialNumbers = append(tmpSerialNumbers, listCoin[index].SerialNumber)
	}

	mapSerialNumber, err := StateM.RpcCaller.HasSerialNumbers(StateM.AccountManage.Account.PaymentAddress, tmpSerialNumbers, tokenID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot check HasSerialNUmber. Error: %v", err))
	}
	unspentCoins := make([]*models.Coins, 0)
	for index := range listCoin {
		isSpent, ok := mapSerialNumber[listCoin[index].SerialNumber]
		if ok && isSpent == false {
			unspentCoins = append(unspentCoins, listCoin[index])
		}
	}
	return unspentCoins, nil
}

func (am *AccountManager) GetRandomCommitments(inputCoinsStr, tokenIDStr string) ([]uint64, []uint64, []string, error) {
	paymentAddressStr := StateM.AccountManage.Account.PaymentAddress
	return StateM.RpcCaller.GetRandomCommitments(paymentAddressStr, inputCoinsStr, tokenIDStr)
}

func (am *AccountManager) ChooseBestCoinsToSpend(coins []*models.Coins, amount uint64) ([]*models.Coins, []*models.Coins, uint64, error) {
	resultCoins := make([]*models.Coins, 0)
	remainCoins := make([]*models.Coins, 0)

	// either take the smallest coins, or a single largest one
	overAmount := uint64(0)
	overAmountCoin := new(models.Coins)
	lowerAmountCoins := make([]*models.Coins, 0)
	for _, coin := range coins {
		value, err := strconv.ParseUint(coin.Value, 10, 64)
		if err != nil {
			continue
		}
		if value < amount {
			lowerAmountCoins = append(lowerAmountCoins, coin)
		} else {
			if overAmount == 0 || value < overAmount {
				overAmount = value
				overAmountCoin = coin
			}
			remainCoins = append(remainCoins, coin)
		}
	}

	sort.Slice(lowerAmountCoins, func(i, j int) bool {
		valuei, _ := strconv.ParseUint(lowerAmountCoins[i].Value, 10, 64)
		valuej, _ := strconv.ParseUint(lowerAmountCoins[j].Value, 10, 64)
		return valuei < valuej
	})

	finalAmount := uint64(0)
	if overAmount > amount {
		resultCoins = append(resultCoins, overAmountCoin)
		v, _ := strconv.ParseUint(overAmountCoin.Value, 10, 64)
		finalAmount += v
		if len(coins)-2 > common.MinUnspentCoins && len(lowerAmountCoins) > 0 {
			resultCoins = append(resultCoins, lowerAmountCoins[0])
			v, _ := strconv.ParseUint(lowerAmountCoins[0].Value, 10, 64)
			finalAmount += v
		}
	} else {
		l := len(lowerAmountCoins)
		for i := range lowerAmountCoins {
			resultCoins = append(resultCoins, lowerAmountCoins[l-1-i])
			v, _ := strconv.ParseUint(lowerAmountCoins[l-1-i].Value, 10, 64)
			finalAmount += v
			if finalAmount > amount {
				break
			}
		}

		if finalAmount < amount {
			return nil, nil, 0, errors.New(fmt.Sprintf("Not enough coin %v to spend", overAmountCoin.TokenID))
		}
		if len(resultCoins) > common.MaxTxInput {
			return nil, nil, 0, errors.New("Number of Input Coins is larger than MaxTxInput")
		}
	}

	return resultCoins, remainCoins, finalAmount, nil
}
