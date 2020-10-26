package controllers

import (
	"errors"
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/common"
	"incwallet/app/models"
	"math"
	"net/http"
)

/*
Accounts controller
*/
type AccountsCtrl struct {
	*revel.Controller
}

type AccountParam struct {
	Name       string `json:"name"`
	Passphrase string `json:"passphrase"`
	TokenID    string `json:"tokenid"`
	Limit      int    `json:"limit"`
	PrivateKey string `json:"privatekey"`
	PublicKey  string `json:"publickey"`
}

/*
import Account
- account name
- private key
- passphrase
*/
func (c AccountsCtrl) ImportAccount() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	} else {
		wallet := &models.Wallet{}
		if err := database.Wallet.Find(bson.M{"walletid": StateM.WalletManager.WalletID}).One(&wallet); err != nil {
			revel.AppLog.Errorf("Does not exist any Wallet in database. Error %v", err)
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot import account, create or import wallet first"), err.Error(), 0))
		}

		err := StateM.AccountManage.ImportAccount(accountParam.Name, accountParam.PrivateKey, accountParam.Passphrase)
		if err != nil {
			revel.AppLog.Errorf("Cannot add account to database. Error %v", err)
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot import account"), err.Error(), 0))
		}

		// start sync data
		errChan := make(chan error)
		go func(errChan chan error) {
			err := JobSyncAccountFromRemote(accountParam.PrivateKey)
			errChan <- err
		}(errChan)

		for {
			if err := <- errChan; err != nil {
				return c.RenderJSON(responseJsonBuilder(errors.New("cannot sync from import account"), err.Error(), 0))
			}
			break
		}

		c.Response.Status = http.StatusCreated
		return c.RenderJSON(responseJsonBuilder(nil, infoJsonBuilder(StateM.AccountManage.Account, "", 0, 0 ,0), 0))
	}
}

/*
Add Account
- account name
- passphrase
*/
func (c AccountsCtrl) AddAccount() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	} else {
		wallet := &models.Wallet{}
		if err := database.Wallet.Find(bson.M{"walletid": StateM.WalletManager.WalletID}).One(&wallet); err != nil {
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot add account, create or import wallet first"), err.Error(), 0))
		}

		masterKey, err := StateM.WalletManager.SafeStore.DecryptMasterKey(wallet, accountParam.Passphrase)
		if err != nil {
			revel.AppLog.Errorf("Cannot Decrypt Master Key. Error %v", err)
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot decrypt master key"), err.Error(), 0))
		}

		privateKeyStr, err := StateM.AccountManage.AddAccount(accountParam.Name, masterKey, wallet.WalletId, wallet.ShardID, accountParam.Passphrase)
		if err != nil {
			revel.AppLog.Errorf("Cannot add account to database. Error %v", err)
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot add account"), err.Error(), 0))
		}

		// start sync data
		errChan := make(chan error)
		go func(errChan chan error) {
			err := JobSyncAccountFromRemote(privateKeyStr)
			errChan <- err
		}(errChan)

		for {
			if err := <- errChan; err != nil {
				return c.RenderJSON(responseJsonBuilder(errors.New("cannot sync from add account"), err.Error(), 0))
			}
			break
		}

		c.Response.Status = http.StatusCreated
		return c.RenderJSON(responseJsonBuilder(nil, infoJsonBuilder(StateM.AccountManage.Account, "", 0 ,0 ,0), 0))
	}
}

/*
Switch Account
- account name
- passphrase
*/
func (c AccountsCtrl) SwitchAccount() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	} else {
		if flag, _ := IsStateFull(); !flag {
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot switch account, import or add account first"), "", 0))
		}
		// Switch Account
		privateKeyStr, err := StateM.AccountManage.SwitchAccount(accountParam.Name, accountParam.Passphrase)
		if err != nil {
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot switch account"), err.Error(), 0))
		}

		// start sync data
		errChan := make(chan error)
		go func(errChan chan error) {
			err := JobSyncAccountFromRemote(privateKeyStr)
			errChan <- err
		}(errChan)

		for {
			if err := <- errChan; err != nil {
				return c.RenderJSON(responseJsonBuilder(errors.New("cannot sync from switched account"), err.Error(), 0))
			}
			break
		}

		c.Response.Status = http.StatusCreated
		return c.RenderJSON(responseJsonBuilder(nil, infoJsonBuilder(StateM.AccountManage.Account, "", 0 ,0 ,0), 0))
	}
}

/*
Balance Account
- token id
*/
func (c AccountsCtrl) GetBalance() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	} else {
		if flag, _ := IsStateFull(); !flag {
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot get balance, import or add account first"), "", 0))
		}
		mapBalance := make(map[string]uint64)
		if len(accountParam.TokenID) == 0 {
			mapBalance, err = StateM.AccountManage.GetBalance("","")
		} else {
			mapBalance, err = StateM.AccountManage.GetBalance("", accountParam.TokenID)
		}

		if err != nil {
			revel.AppLog.Errorf("cannot retrieve balance account. Error %v", err)
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot retrieve balance"), err.Error(), 0))
		}

		c.Response.Status = http.StatusCreated
		return c.RenderJSON(responseJsonBuilder(nil, balanceJsonBuilder(mapBalance), 0))
	}
}

/*
List Account
*/
func (c AccountsCtrl) ListAccount() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull(); !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot list account, import or add account first"), "", 0))
	}
	var listAccounts []models.Account
	if err := database.Accounts.Find(bson.M{"wallet": StateM.WalletManager.WalletID}).All(&listAccounts); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get accounts info"), err.Error(), 0))
	}
	listTotalPRV := make([]float64, 0)
	listTotalUSDT := make([]float64, 0)
	listTotalBTC := make([]float64, 0)
	for _, acc := range listAccounts {
		mapBalance, err := StateM.AccountManage.GetBalance(acc.PublicKey,"")
		if err != nil {
			return c.RenderJSON(responseJsonBuilder(errors.New(fmt.Sprintf("cannot get balance for account %v", acc.PublicKey)), err.Error(), 0))
		}
		totalPRV, _ := getTotalValueInPRV(mapBalance)
		listTotalPRV = append(listTotalPRV, float64(totalPRV) / math.Pow10(9))
		totalUSDT, _, _ := getExchangeRate(common.PRVID, common.USDTID, totalPRV, 0, true)
		listTotalUSDT = append(listTotalUSDT, float64(totalUSDT) / math.Pow10(6))
		totalBTC, _, _ := getExchangeRate(common.PRVID, common.BTCID, totalPRV, 0, true)
		listTotalBTC = append(listTotalBTC, float64(totalBTC) / math.Pow10(9))
	}

	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, accountJsonBuilder(listAccounts, listTotalPRV, listTotalUSDT, listTotalBTC), 0))
}

/*
Sync Account
- publickey
- passphrase
*/
func (c AccountsCtrl) SyncAccount() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}
	publicKey := ""
	if accountParam.PublicKey != "" {
		publicKey = accountParam.PublicKey
	}
	err := StateM.AccountManage.SyncAccount(publicKey, accountParam.Passphrase)
	if err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account"), err.Error(), 0))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, "done", 0))
}

/*
Sync All Account
- passphrase
*/
func (c AccountsCtrl) SyncAllAccounts() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}

	var listAccounts []*models.Account
	if err := database.Accounts.Find(bson.M{"wallet": StateM.WalletManager.WalletID}).All(&listAccounts); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get accounts info"), err.Error(), 0))
	}
	listErrors := StateM.AccountManage.SyncAllAccounts(listAccounts, accountParam.Passphrase)
	if len(listErrors) >0  {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot sync all account"), listErrors, 0))
	}
	return c.RenderJSON(responseJsonBuilder(nil, "done", 0))
}

/*
List unspent coins
- tokenid
*/
func (c AccountsCtrl) ListUnspent() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot list utxo, import or add account first"), "", 0))
	}
	var query bson.M
	if accountParam.TokenID == "" {
		query = bson.M{
			"publickey": StateM.AccountManage.AccountID,
			"isspent": false,
		}
	} else {
		query = bson.M{
			"publickey": StateM.AccountManage.AccountID,
			"tokenid": accountParam.TokenID,
			"isspent": false,
		}
	}
	var listUnspent []models.Coins
	if err := database.Coins.Find(query).Sort("tokenid").All(&listUnspent); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot load unpsent coin"), err.Error(), 0))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, coinDetailJsonBuilder(listUnspent), 0))
}

/*
Account info
- passphrase
- publicjey
*/
func (c AccountsCtrl) GetInfo() revel.Result {
	accountParam := &AccountParam{}
	if err := c.Params.BindJSON(&accountParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}

	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot show info, import or add account first"), "", 0))
	}

	account := new(models.Account)
	if accountParam.PublicKey != "" {
		if err := database.Accounts.Find(bson.M{"publickey": accountParam.PublicKey}).One(&account); err != nil {
			return c.RenderJSON(responseJsonBuilder(errors.New(fmt.Sprintf("cannot get info for account %v", accountParam.PublicKey)), err.Error(), 0))
		}
	} else {
		account = StateM.AccountManage.Account
	}

	mapBalance, err := StateM.AccountManage.GetBalance(account.PublicKey,"")
	if err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New(fmt.Sprintf("cannot get balance for account %v", StateM.AccountManage.Account.PublicKey)), err.Error(), 0))
	}
	totalPRV, _ := getTotalValueInPRV(mapBalance)
	totalPRVView := float64(totalPRV) / math.Pow10(9)

	totalUSDT, _, _ := getExchangeRate(common.PRVID, common.USDTID, totalPRV, 0, true)
	totalUSDTView := float64(totalUSDT) / math.Pow10(6)

	totalBTC, _, _ := getExchangeRate(common.PRVID, common.BTCID, totalPRV, 0, true)
	totalBTCView := float64(totalBTC) / math.Pow10(9)

	var privateKeyStr string
	if accountParam.Passphrase != "" {
		privateKey, err := StateM.WalletManager.SafeStore.DecryptPrivateKey(
			&account.Crypto,
			accountParam.Passphrase)
		if err != nil {
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot get account info"), err.Error(), 0))
		}
		privateKeyStr = string(privateKey)
	} else {
		privateKeyStr = "enter passphrase to view ... "
	}

	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, infoJsonBuilder(account, privateKeyStr, totalPRVView, totalUSDTView, totalBTCView), 0))
}