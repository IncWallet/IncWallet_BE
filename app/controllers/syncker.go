package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/cron"
	"incwallet/app/database"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"incwallet/app/lib/crypto"
	"incwallet/app/lib/hdwallet"
	"incwallet/app/models"

	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type SynckerManager struct {
	privateKeyStr string
}

func InitSynckerManager(privateKeyStr string) *SynckerManager {
	return &SynckerManager{
		privateKeyStr: privateKeyStr,
	}
}

func GetAmountFromInputCoin(coins []models.AutoCoin) (uint64, error) {
	amount := uint64(0)
	if coins == nil || len(coins) == 0 {
		return amount, nil
	}
	for _, coin := range coins {
		tmpCoin := new(models.Coins)
		err := database.Coins.Find(bson.M{
			"serialnumber": coin.CoinDetails.SerialNumber,
			"publickey":    StateM.AccountManage.AccountID}).One(&tmpCoin)
		if err != nil {
			return uint64(0), nil
		}
		coinValue, err := strconv.ParseUint(tmpCoin.Value, 10, 64)
		if err != nil {
			return uint64(0), err
		}
		amount += coinValue
	}
	return amount, nil
}

func GetAmountFromOutputCoin(coins []models.AutoCoin, publicKey string) (uint64, error) {
	amount := uint64(0)
	if coins == nil || len(coins) == 0 {
		return amount, nil
	}

	for _, coin := range coins {
		if publicKey != coin.CoinDetails.PublicKey {
			continue
		}

		tmpCoin := new(models.Coins)
		err := database.Coins.Find(bson.M{"coincommitment": coin.CoinDetails.CoinCommitment}).One(&tmpCoin)
		if err != nil {
			return uint64(0), errors.New(fmt.Sprintf("not found coin commitment %v", coin.CoinDetails.CoinCommitment))
		}
		coinValue, err := strconv.ParseUint(tmpCoin.Value, 10, 64)
		if err != nil {
			return uint64(0), err
		}
		amount += coinValue
	}
	return amount, nil
}

func GetTxHistory(autoTxHash *models.AutoTxByHash, publickeyStr string) (*models.TxHistory, error) {
	inAmountPRV, err := GetAmountFromInputCoin(autoTxHash.Result.ProofDetail.InputCoins)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("Cannot sum input amount. Error %v", err))
		return nil, errors.New(fmt.Sprintf("cannot sum input amount. Error %v", err))
	}
	outAmountPRV, err := GetAmountFromOutputCoin(autoTxHash.Result.ProofDetail.OutputCoins, publickeyStr)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("Cannot sum output amount. Error %v", err))
		return nil, errors.New(fmt.Sprintf("cannot sum output amount. Error %v", err))
	}
	tokenID := common.PRVID
	if autoTxHash.Result.Type == "tp" {
		tokenID = autoTxHash.Result.PrivacyCustomTokenID
	}
	inAmountToken, err := GetAmountFromInputCoin(autoTxHash.Result.PrivacyCustomTokenProofDetail.InputCoins)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("Cannot sum input token amount. Error %v", err))
		return nil, errors.New(fmt.Sprintf("cannot sum input token amount. Error %v", err))
	}
	outAmountToken, err := GetAmountFromOutputCoin(autoTxHash.Result.PrivacyCustomTokenProofDetail.OutputCoins, publickeyStr)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("Cannot sum output token amount. Error %v", err))
		return nil, errors.New(fmt.Sprintf("cannot sum output token amount. Error %v", err))
	}

	txHistory := &models.TxHistory{
		TxHash:     autoTxHash.Result.Hash,
		PublicKey:  publickeyStr,
		LockTime:   autoTxHash.Result.LockTime,
		Type:       autoTxHash.Result.Type,
		Fee:        autoTxHash.Result.Fee,
		VInPRVs:    inAmountPRV,
		VOutPRVs:   outAmountPRV,
		TokenID:    tokenID,
		TokenFee:   autoTxHash.Result.PrivacyCustomTokenFee,
		VInTokens:  inAmountToken,
		VOutTokens: outAmountToken,
	}
	return txHistory, nil
}

func GetTokenIDFromTxHash(hash string) (string, error) {
	dataByte, _ := StateM.RpcCaller.GetTransactionByHash(hash)
	autoTxHash := new(models.AutoTxByHash)
	if err := json.Unmarshal(dataByte, autoTxHash); err != nil {
		return "", errors.New(fmt.Sprintf("cannot unmarshal tx hash %v. Error %v", hash, err))
	}
	if autoTxHash.Error != nil {
		return "", errors.New(fmt.Sprintf("cannot get tx hash %v. Error %v", hash, autoTxHash.Error))
	}
	if autoTxHash.Result.PrivacyCustomTokenID != "" {
		return autoTxHash.Result.PrivacyCustomTokenID, nil
	} else {
		return common.PRVID, nil
	}
	return "", nil
}

func GetCurrentTxHashHistory() (map[string]bool, error){
	var listTxHashHistory []*models.TxHistory
	err := database.TxHistory.Find(bson.M{
		"publickey": StateM.AccountManage.AccountID}).All(&listTxHashHistory)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot get list tx history. Error %v", err))
	}

	mapTxHashHistory := make(map[string]bool)
	for _, tx := range listTxHashHistory {
		mapTxHashHistory[tx.TxHash] = true
	}
	return mapTxHashHistory, nil
}

func JobSyncAccountFromRemote(privateKey string) error {
	if StateM.JobManager != nil {
		revel.AppLog.Warnf("Stop current Syncker Job")
		StateM.JobManager.Stop()
	}
	synckerManager := InitSynckerManager(privateKey)

	// Sync token
	revel.AppLog.Info("Sync tokens ...")
	if err := synckerManager.SyncAllToken(); err != nil {
		revel.AppLog.Error(fmt.Sprintf("Sync token error. Error %v", err))
		return errors.New(fmt.Sprintf("Sync token error. Error %v", err))
	}
	revel.AppLog.Info("Sync tokens Done!!!")

	// Sync account once
	if err := synckerManager.SyncAccountJob(); err != nil {
		revel.AppLog.Error(fmt.Sprintf("Sync coin error. Error %v", err))
		return errors.New(fmt.Sprintf("Sync coin error. Error %v", err))
	}

	// Sync account job
	StateM.JobManager = cron.New()
	_ = StateM.JobManager.AddFunc("@every 40s", func() {
		_ = synckerManager.SyncAccountJob()
	})
	StateM.JobManager.Start()
	return nil
}

func (sm *SynckerManager) SyncAllToken() error {
	nNewToken, err := StateM.NetworkManager.UpdateAllToken()
	if err != nil {
		return errors.New(fmt.Sprintf("cannot sync token info. Error %v", err))
	}
	revel.AppLog.Infof("%v new token added", nNewToken)
	return nil
}

func (sm *SynckerManager) UpdateOutputCoins(paymentAddressStr string, privateKey hdwallet.PrivateKey, outputCoins []models.Coins, tokenID string) error {
	listSerialNumber := make([]string, 0)
	listUnSaveCoins := make([]models.Coins, 0)
	for _, coin := range outputCoins {
		tmpCoin := new(models.Coins)
		err := database.Coins.Find(bson.M{"snderivator": coin.SNDerivator}).One(&tmpCoin)
		if err != nil || !tmpCoin.IsSpent {
			snd, _, _ := base58.Base58Check{}.Decode(coin.SNDerivator)
			sn := crypto.GenerateSerialNumber(privateKey, snd)
			coin.SerialNumber = base58.Base58Check{}.Encode(sn, common.ZeroByte)
			listUnSaveCoins = append(listUnSaveCoins, coin)
			listSerialNumber = append(listSerialNumber, coin.SerialNumber)
		}
	}
	mapHasSerialNumber, err := StateM.RpcCaller.HasSerialNumbers(paymentAddressStr, listSerialNumber, tokenID)
	if err != nil {
		revel.AppLog.Error("Cannot check has serial number")
		return errors.New(fmt.Sprintf("cannot check has serial number. Error %v", err))
	}
	for _, coin := range listUnSaveCoins {
		if flag, found := mapHasSerialNumber[coin.SerialNumber]; found {
			coin.IsSpent = flag
		}
		tmpMCoin := new(models.Coins)
		err := database.Coins.Find(bson.M{"snderivator": coin.SNDerivator}).One(&tmpMCoin)
		if err != nil {
			if err1 := database.Coins.Insert(coin); err1 != nil {
				revel.AppLog.Error("Cannot insert new coin")
				return errors.New(fmt.Sprintf("Cannot insert new coin. Error %v", err1))
			} else {
				revel.AppLog.Info(fmt.Sprintf("Insert new coin. SND: %v", coin.SNDerivator))
			}

		} else {
			if tmpMCoin.IsSpent != coin.IsSpent || tmpMCoin.SerialNumber != coin.SerialNumber {
				if err1 := database.Coins.Update(tmpMCoin, coin); err1 != nil {
					revel.AppLog.Error("Cannot update new coin")
					return errors.New(fmt.Sprintf("Cannot update new coin. Error %v", err1))
				} else {
					revel.AppLog.Info(fmt.Sprintf("Update new coin. SND: %v", coin.SNDerivator))
				}
			}
		}
	}
	return nil
}

func (sm *SynckerManager) UpdateTxHash(publicKeyStr string, hash string) error {

	dataByte, _ := StateM.RpcCaller.GetTransactionByHash(hash)
	autoTxHash := new(models.AutoTxByHash)
	if err := json.Unmarshal(dataByte, autoTxHash); err != nil {
		revel.AppLog.Errorf("Cannot unmarshal tx hash %v", hash)
		return errors.New(fmt.Sprintf("cannot unmarshal tx hash %v. Error %v", hash, err))
	}

	txHistory, err := GetTxHistory(autoTxHash, publicKeyStr)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("Error %v", err))
		return errors.New(fmt.Sprintf("error %v", err))
	}
	if err := database.TxHistory.Insert(txHistory); err != nil {
		revel.AppLog.Error(fmt.Sprintf("Cannot insert tx history to database. Error %v", err))
		return errors.New(fmt.Sprintf("cannot insert tx history to database .Error %v", err))
	}
	revel.AppLog.Info(fmt.Sprintf("insert TxHash %v", hash))
	return nil
}

func (sm *SynckerManager) UpdateTokenListByAccount(paymentAddressStr string) (map[string]bool, error) {
	listTxHash, err := StateM.RpcCaller.GetListReceiveTxHash(paymentAddressStr)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("Cannot get all tx hash. Error %v", err))
		return nil, errors.New(fmt.Sprintf("Cannot get all tx hash. Error %v", err))
	}
	mapToken := make(map[string]bool)
	errorChan := make(chan error)
	tokenIDChan := make(chan string)

	mapTxHashHistory, err := GetCurrentTxHashHistory()
	if err != nil {
		return nil, errors.New(err.Error())
	}

	for _, hash := range listTxHash {
		go func(hash string, errorChan chan error, tokenIDChan chan string) {
			if _, found := mapTxHashHistory[hash]; found {
				errorChan <- nil
				tokenIDChan <- ""
			} else {
				tokenID, err := GetTokenIDFromTxHash(hash)
				errorChan <- err
				tokenIDChan <- tokenID
			}
		}(hash, errorChan, tokenIDChan)
	}

	for range listTxHash {
		if err := <-errorChan; err != nil {
			revel.AppLog.Warnf("error %v", err)
		}
		tokenID := <-tokenIDChan
		if _, found := mapToken[tokenID]; !found && tokenID != "" {
			mapToken[tokenID] = true
		}
	}
	return mapToken, nil
}

func (sm *SynckerManager) SyncAccountJob() error {
	revel.AppLog.Info("sync account ...")
	keyWallet, publicKeyStr, paymentAddressStr, _, err := hdwallet.GetKeyWalletInfoFromPrivateKey(sm.privateKeyStr)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot init key wallet. Error %v", err))
	}

	// Get token list to update
	mapToken, err := sm.UpdateTokenListByAccount(paymentAddressStr)

	for tokenID, _ := range mapToken {
		outputCoins, err := StateM.RpcCaller.GetListOutputCoins(paymentAddressStr, keyWallet.KeySet.ReadonlyKey.Rk[:], tokenID)
		if err != nil {
			revel.AppLog.Error("Cannot get all output coins from account info")
			return errors.New(fmt.Sprintf("cannot get all output coins from account info. Error %v", err))
		}
		err = sm.UpdateOutputCoins(paymentAddressStr, keyWallet.KeySet.PrivateKey, outputCoins, tokenID)
		if err != nil {
			revel.AppLog.Error(fmt.Sprintf("Cannot update outputcoin. Error %v", err))
			return errors.New(fmt.Sprintf("cannot update outputcoins. Error %v", err))
		}
	}

	//Update list receiver txhash
	listTxHash, err := StateM.RpcCaller.GetListReceiveTxHash(paymentAddressStr)
	if err != nil {
		revel.AppLog.Error(fmt.Sprintf("Cannot get all tx hash. Error %v", err))
		return errors.New(fmt.Sprintf("Cannot get all tx hash. Error %v", err))
	}

	mapTxHashHistory, err := GetCurrentTxHashHistory()
	if err != nil {
		return errors.New(err.Error())
	}

	errorChan := make(chan error)
	for _, item := range listTxHash {
		go func(hash string, errorChan chan error) {
			if _, found := mapTxHashHistory[hash]; found {
				errorChan <- nil
			} else {
				err := sm.UpdateTxHash(publicKeyStr, hash)
				errorChan <- err
			}
		}(item, errorChan)
	}

	for _, hash := range listTxHash {
		if err := <-errorChan; err != nil {
			revel.AppLog.Warnf("cannot update tx hash %v. Error %v", hash, err)
		}
	}
	revel.AppLog.Info("sync account done ...")
	return nil
}