package controllers

import (
	"errors"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/common"
	"incwallet/app/lib/hdwallet"
	"incwallet/app/lib/transaction"
	"incwallet/app/models"
	"net/http"
	"strconv"
)

type TransactionsCtrl struct {
	*revel.Controller
}

type TxParam struct {
	Receivers  map[string]uint64    `json:"receivers"`
	Fee        uint64               `json:"fee"`
	Info       string               `json:"info"`
	Passphrase string               `json:"passphrase"`
	TokenID    string               `json:"tokenid"`
	Metadata   transaction.Metadata `json:"metadata"`
	TxHash     string               `json:"txhash"`
}

type TxTradeParam struct {
	FromTokenID      string `json:"fromtokenid"`
	ToTokenID        string `json:"totokenid"`
	SendAmount       uint64 `json:"sendamount"`
	MinReceiveAmount uint64 `json:"minreceiveamount"`
	TradingFee       uint64 `json:"tradingfee"`
	TraderAddressStr string `json:"traderaddressstr"`
	TxFee            uint64 `json:"txfee"`
	Passphrase       string `json:"passphrase"`
}

type TxMineParam struct {
	PublicKey      string `json:"publickey"`
	MiningKey      string `json:"miningkey"`
	PaymentAddress string `json:"paymentaddress"`
	TxFee          uint64 `json:"txfee"`
	Passphrase     string `json:"passphrase"`
}

/* TRADE TRANSACTION */
/*
Init Trade PRV Tx
- from token
- to token
- send amount
- min receive amount
- trading fee
- trader address str
- passphrase
*/
func (c TransactionsCtrl) InitTradePRVTransaction() revel.Result {
	txTradeParam := &TxTradeParam{}
	if err := c.Params.BindJSON(&txTradeParam); err != nil {
		return c.RenderJSON("Error: bad request")
	}
	if flag, _ := IsStateFull(); !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction, import or add account first"), "", 0))
	}

	txManager := &TransactionManager{
		am:        StateM.AccountManage,
		rpcCaller: StateM.RpcCaller,
	}
	receiver := make(map[string]uint64)
	receiver[common.BurnAddress2] = txTradeParam.SendAmount + txTradeParam.TradingFee
	txFee := txTradeParam.TxFee
	meta := transaction.NewPDETradeRequestMetadata(
		txTradeParam.ToTokenID,
		txTradeParam.FromTokenID,
		txTradeParam.SendAmount,
		txTradeParam.MinReceiveAmount,
		txTradeParam.TradingFee,
		txTradeParam.TraderAddressStr,
	)

	txHash, err := txManager.InitTransaction(
		receiver,
		txFee,
		"",
		meta,
		common.PRVID,
		false,
		txTradeParam.Passphrase)
	if err != nil {
		revel.AppLog.Errorf("Cannot Init transaction. Error %v", err)
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction"), err.Error(), 0))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, txHash, 0))
}

/*
Init Trade Token Tx
- from token
- to token
- send amount
- min receive amount
- trading fee
- trader address str
- token id
- passphrase
*/
func (c TransactionsCtrl) InitTradeTokenTransaction() revel.Result {
	txTradeParam := &TxTradeParam{}
	if err := c.Params.BindJSON(&txTradeParam); err != nil {
		return c.RenderJSON("Error: bad request")
	}
	if flag, _ := IsStateFull(); !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction, import or add account first"), "", 0))
	}

	txManager := &TransactionManager{
		am:        StateM.AccountManage,
		rpcCaller: StateM.RpcCaller,
	}

	receiver := make(map[string]uint64)
	receiver[common.BurnAddress2] = txTradeParam.SendAmount + txTradeParam.TradingFee
	txFee := txTradeParam.TxFee
	meta := transaction.NewPDETradeRequestMetadata(
		txTradeParam.ToTokenID,
		txTradeParam.FromTokenID,
		txTradeParam.SendAmount,
		txTradeParam.MinReceiveAmount,
		txTradeParam.TradingFee,
		txTradeParam.TraderAddressStr,
	)

	txHash, err := txManager.InitTokenTransaction(
		nil,
		receiver,
		txFee,
		"",
		meta,
		txTradeParam.FromTokenID,
		false,
		txTradeParam.Passphrase)
	if err != nil {
		revel.AppLog.Errorf("Cannot Init transaction. Error %v", err)
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction"), err.Error(), 0))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, txHash, 0))
}

/*
Init Trade Token Tx
- from token
- to token
- send amount
- min receive amount
- trading fee
- trader address str
- token id
- iscross
- passphrase
*/
func (c TransactionsCtrl) InitTradeCrossTokenTransaction() revel.Result {
	txTradeParam := &TxTradeParam{}
	if err := c.Params.BindJSON(&txTradeParam); err != nil {
		return c.RenderJSON("Error: bad request")
	}
	if flag, _ := IsStateFull(); !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction, import or add account first"), "", 0))
	}

	if txTradeParam.FromTokenID == common.PRVID || txTradeParam.ToTokenID == common.PRVID {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction"), "one of from or to token id is PRV", 0))
	}

	txManager := &TransactionManager{
		am:        StateM.AccountManage,
		rpcCaller: StateM.RpcCaller,
	}

	receiverPrv := make(map[string]uint64)
	receiverPrv[common.BurnAddress2] = txTradeParam.TradingFee

	receiverToken := make(map[string]uint64)
	receiverToken[common.BurnAddress2] = txTradeParam.SendAmount
	txFee := txTradeParam.TxFee
	meta := transaction.NewPDETradeCrossRequestMetadata(
		txTradeParam.ToTokenID,
		txTradeParam.FromTokenID,
		txTradeParam.SendAmount,
		txTradeParam.MinReceiveAmount,
		txTradeParam.TradingFee,
		txTradeParam.TraderAddressStr,
	)

	txHash, err := txManager.InitTokenTransaction(
		receiverPrv,
		receiverToken,
		txFee,
		"",
		meta,
		txTradeParam.FromTokenID,
		false,
		txTradeParam.Passphrase)
	if err != nil {
		revel.AppLog.Errorf("Cannot Init transaction. Error %v", err)
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction"), err.Error(), 0))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, txHash, 0))
}

/* TRANSFER TRANSACTION */
/*
Init PRVTx
- Receivers
- Fee
- Info
- Passphrase
*/
func (c TransactionsCtrl) InitTransaction() revel.Result {
	txParam := &TxParam{}
	if err := c.Params.BindJSON(&txParam); err != nil {
		return c.RenderJSON("Error: bad request")
	}

	if flag, _ := IsStateFull(); !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction, import or add account first"), "", 0))
	}

	txManager := &TransactionManager{
		am:        StateM.AccountManage,
		rpcCaller: StateM.RpcCaller,
	}

	txHash, err := txManager.InitTransaction(
		txParam.Receivers,
		txParam.Fee,
		txParam.Info,
		txParam.Metadata,
		common.PRVID,
		true,
		txParam.Passphrase)
	if err != nil {
		revel.AppLog.Errorf("Cannot Init transaction. Error %v", err)
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction"), err.Error(), 0))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, txHash, 0))
}

/*
Init PRVTx Token
- Receivers
- Fee
- Info
- TokenID
- Passphrase
*/
func (c TransactionsCtrl) InitTokenTransaction() revel.Result {
	txParam := &TxParam{}
	if err := c.Params.BindJSON(&txParam); err != nil {
		return c.RenderJSON("Error: bad request")
	}
	if flag, _ := IsStateFull(); !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction, import or add account first"), "", 0))
	}

	txManager := &TransactionManager{
		am:        StateM.AccountManage,
		rpcCaller: StateM.RpcCaller,
	}

	if txParam.TokenID == "" || txParam.TokenID == common.PRVID {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), "Only accept pToken", 0))
	}

	txHash, err := txManager.InitTokenTransaction(
		nil,
		txParam.Receivers,
		txParam.Fee,
		txParam.Info,
		txParam.Metadata,
		txParam.TokenID,
		true,
		txParam.Passphrase)

	if err != nil {
		revel.AppLog.Errorf("Cannot Init token transfer transaction. Error %v", err)
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction"), err.Error(), 0))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, txHash, 0))
}

/* GET INFORMATION */
/*
/transaction/history
TxHistory
- token id
*/
func (c TransactionsCtrl) GetTxHistory() revel.Result {
	if flag, _ := IsStateFull(); !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get tx history, import or add account first"), "", 0))
	}

	var pageIndex, pageSize int
	var err error
	if c.Params.Get("pageindex") == "" && c.Params.Get("pagesize") == "" {
		pageIndex = 1
		pageSize = 1000000000
	} else {
		pageIndex, err = strconv.Atoi(c.Params.Get("pageindex"))
		if err != nil {
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot get all token, pageindex is invalid"), err.Error(), 0))
		}
		pageSize, err = strconv.Atoi(c.Params.Get("pagesize"))
		if err != nil {
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot get all token, pageSize is invalid"), err.Error(), 0))
		}
	}

	tokenID := c.Params.Get("tokenid")
	var query bson.M
	if len(tokenID) != 0 {
		query = bson.M{
			"publickey": StateM.AccountManage.AccountID,
			"tokenid":   tokenID,
		}
	} else {
		query = bson.M{
			"publickey": StateM.AccountManage.AccountID,
		}
	}

	var listTxHistory []models.TxHistory

	size, err := database.TxHistory.Find(query).Count()
	if err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot load tx history"), err.Error(), 0))
	}

	if err := database.TxHistory.Find(query).Sort("-locktime").Skip((pageIndex - 1) * pageSize).Limit(pageSize).All(&listTxHistory); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot load tx history"), err.Error(), 0))
	}

	return c.RenderJSON(responseJsonBuilder(nil, txHistoryJsonBuilder(listTxHistory, size), 0))
}

/*
/transactions/info
Tx Info
- txhash
*/
func (c TransactionsCtrl) GetTxInfo() revel.Result {
	txParam := &TxParam{}
	if err := c.Params.BindJSON(&txParam); err != nil {
		return c.RenderJSON("Error: bad request")
	}

	var tx *models.AutoTxByHash
	tx, err := StateM.RpcCaller.GetAutoTxByHash(txParam.TxHash)
	if err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get tx info from txhash"), err.Error(), 0))
	}
	return c.RenderJSON(responseJsonBuilder(nil, txInfoJsonBuilder(tx), 0))
}

/* MINER TRANSACTION */
/*
Init Stop Auto Staking Tx
- publickey
- mingingkey
- txfee
- passphrase
*/
func (c TransactionsCtrl) InitStopAutoStakingTransaction() revel.Result {
	txMineParam := &TxMineParam{}
	if err := c.Params.BindJSON(&txMineParam); err != nil {
		return c.RenderJSON("Error: bad request")
	}
	if flag, _ := IsStateFull(); !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction, import or add account first"), "", 0))
	}

	txManager := &TransactionManager{
		am:        StateM.AccountManage,
		rpcCaller: StateM.RpcCaller,
	}
	receiver := make(map[string]uint64)
	receiver[common.BurnAddress2] = uint64(0)
	txFee := txMineParam.TxFee

	committeePublicKey, err := hdwallet.GetMiningPubKey(txMineParam.MiningKey, txMineParam.PaymentAddress)
	if err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction, cannot get mining public key"), err.Error(), 0))
	}

	meta := transaction.NewStopAutoStakingMetadata(
		committeePublicKey,
	)

	txHash, err := txManager.InitTransaction(
		receiver,
		txFee,
		"",
		meta,
		common.PRVID,
		false,
		txMineParam.Passphrase)
	if err != nil {
		revel.AppLog.Errorf("Cannot Init transaction. Error %v", err)
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot create transaction"), err.Error(), 0))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, txHash, 0))
}
