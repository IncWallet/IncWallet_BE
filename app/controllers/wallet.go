package controllers

import (
	"errors"
	"github.com/revel/revel"
	"net/http"
)

/*
Books controller
*/
type WalletCtrl struct {
	*revel.Controller
}

type WalletParam struct {
	Security   int    `json:"security"`
	Passphrase string `json:"passphrase"`
	Mnemonic   string `json:"mnemonic"`
	Network    string `json:"network"`
}

/*
Import Wallet
*/
func (c WalletCtrl) ImportWallet() revel.Result {
	walletParam := &WalletParam{}
	if err := c.Params.BindJSON(&walletParam); err != nil {
		return c.RenderJSON("Error: bad request")
	} else {
		err := StateM.WalletManager.ImportWallet(walletParam.Mnemonic, walletParam.Passphrase, walletParam.Network)
		if err != nil {
			revel.AppLog.Errorf("Cannot create and save Wallet to database. Error %v", err)
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot create message"), err.Error(), 0))
		}
		c.Response.Status = http.StatusCreated
		return c.RenderJSON(responseJsonBuilder(nil, "Done", 0))
	}
}

/*
Create Wallet
*/
func (c WalletCtrl) CreateWallet() revel.Result {
	walletParam := &WalletParam{}
	if err := c.Params.BindJSON(&walletParam); err != nil {
		return c.RenderJSON("Error: bad request")
	} else {
		mnemonic, err := StateM.WalletManager.CreateNewWallet(walletParam.Security, walletParam.Passphrase, walletParam.Network)
		if err != nil {
			revel.AppLog.Errorf("Cannot create and save Wallet to database. Error %v", err)
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJSON(responseJsonBuilder(errors.New("Cannot create message"), err.Error(), 0))
		}
		c.Response.Status = http.StatusCreated
		return c.RenderJSON(responseJsonBuilder(nil, mnemonic, 0))
	}
}
