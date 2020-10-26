package controllers

import (
	"fmt"
	"incwallet/app/lib/common"
	"incwallet/app/lib/hdwallet"
	"incwallet/app/models"
)

type InfoJson struct {
	AccountName    string
	PrivateKey     string
	PaymentAddress string
	PublicKey      string
	ViewingKey     string
	MiningKey      string
	Network        string
	ValuePRV       float64
	ValueUSDT      float64
	ValueBTC       float64
}

type AccBalanceJson struct {
	TokenID      string
	TokenName    string
	TokenDecimal int
	TokenSymbol  string
	TokenImage   string
	Amount       uint64
}

type TxHistoryJson struct {
	TxHash       string
	LockTime     string
	Type         string
	TokenID      string
	TokenName    string
	TokenSymbol  string
	TokenDecimal int
	TokenImage   string
	Fee          uint64
	Amount       uint64
}

type CoinJson struct {
	PublicKey      string
	CoinCommitment string
	SNDerivator    string
	SerialNumber   string
	Value          string
	IsSpent        bool
	TokenID        string
	TokenName      string
	TokenSymbol    string
	TokenDecimal   int
	TokenImage     string
}

type AccountJson struct {
	Name           string
	PaymentAddress string
	PublicKey      string
	ViewingKey     string
	MiningKey      string
	ValuePRV       float64
	ValueUSDT      float64
	ValueBTC       float64
}

type ResponseJson struct {
	Error *AppError
	Msg   interface{}
}

type AppError struct {
	Code int
	Msg  string
}

type TxInfoJson struct {
	BlockHash   string
	BlockHeight uint64
	TxSize      int
	Index       int
	ShardID     int
	Hash        string
	Version     int
	Type        string
	LockTime    string
	Fee         uint64
	ProofDetail struct {
		InputCoins  []models.AutoCoin
		OutputCoins []models.AutoCoin
	}
	SigPubKey                     string
	Sig                           string
	Metadata                      string
	PrivacyCustomTokenID          string
	PrivacyCustomTokenName        string
	PrivacyCustomTokenSymbol      string
	PrivacyCustomTokenProofDetail struct {
		InputCoins  []models.AutoCoin
		OutputCoins []models.AutoCoin
	}
	PrivacyCustomTokenIsPrivacy bool
	PrivacyCustomTokenFee       uint64
	IsInMempool                 bool
	IsInBlock                   bool
	Info                        string
}

type PdeHistoryJson struct {
	TraderAddressStr    string
	ReceiveTokenIDStr   string
	ReceiveTokenSymbol  string
	ReceiveTokenName    string
	ReceiveTokenDecimal int
	ReceiveTokenImage   string
	ReceiverAmount      uint64
	SendTokenIDStr      string
	SendTokenSymbol     string
	SendTokenName       string
	SendTokenDecimal    int
	SendTokenImage      string
	SendAmount          uint64
	RequestedTxID       string
	BlockHeight         uint64
	LockTime            string
	ShardID             int
	Status              string
}

type PdePoolPairPriceJson struct {
	FromTokenID        string
	FromTokenName      string
	FromTokenSymbol    string
	FromTokenDecimal   int
	FromTokenImage     string
	FromTokenPoolValue uint64
	ToTokenID          string
	ToTokenName        string
	ToTokenSymbol      string
	ToTokenDecimal     int
	ToTokenImage       string
	ToTokenPoolValue   uint64
	ExchangeAmount     uint64
	ReceiveAmount      uint64
	BeaconHeight       uint64
}

type MinerInfoJson struct {
	AccountName    string
	PaymentAddress string
	MiningKey      string
	BeaconHeight   uint64
	Epoch          uint64
	Reward         uint64
	Status         string
	ShardID        int
	Index          int
}

type StateJson struct {
	WalletID    string
	AccountID   string
	ShardID     byte
	NetworkName string
	NetworkURL  string
	BestBlock   map[int]uint64
}

type Result struct {
	Size int
	Detail interface{}
}

func getTokenByID(tokenID string) *models.Tokens {
	if token, err := StateM.NetworkManager.GetTokenByID(tokenID); err != nil {
		return &models.Tokens{
			ID:       tokenID,
			Symbol:   tokenID[len(tokenID)-4:] + "*",
			Name:     tokenID[len(tokenID)-4:] + "*",
			Decimal:  9,
			EDecimal: 9,
			ESymbol:  "",
			Verified: false,
			Image:    "",
		}
	} else {
		if token.Verified == false {
			token.Symbol = token.Symbol + "*"
		}
		return token
	}

}

func responseJsonBuilder(err error, msg interface{}, code int) *ResponseJson {
	if err == nil {
		return &ResponseJson{
			Error: nil,
			Msg:   msg,
		}
	}
	return &ResponseJson{
		Error: &AppError{
			Code: code,
			Msg:  err.Error(),
		},
		Msg: msg,
	}
}

func accountJsonBuilder(listAccounts []models.Account, listTotalPRV, listTotalUSDT, listTotalBTC []float64) []AccountJson {
	listAccountJson := make([]AccountJson, 0)
	for i, account := range listAccounts {
		tmp := &AccountJson{
			Name:           account.Name,
			PaymentAddress: account.PaymentAddress,
			PublicKey:      account.PublicKey,
			ViewingKey:     account.ViewKey,
			MiningKey:      account.MiningKey,
			ValuePRV:       listTotalPRV[i],
			ValueBTC:       listTotalBTC[i],
			ValueUSDT:      listTotalUSDT[i],
		}
		listAccountJson = append(listAccountJson, *tmp)
	}
	return listAccountJson
}

func balanceJsonBuilder(listBalance map[string]uint64) []AccBalanceJson {
	listAccountBalanceJson := make([]AccBalanceJson, 0)
	for tokenID, balance := range listBalance {
		token := getTokenByID(tokenID)
		accountBalanceJson := &AccBalanceJson{
			TokenID:      tokenID,
			TokenName:    token.Name,
			TokenSymbol:  token.Symbol,
			TokenDecimal: token.Decimal,
			TokenImage:   token.Image,
			Amount:       balance,
		}
		listAccountBalanceJson = append(listAccountBalanceJson, *accountBalanceJson)
	}
	return listAccountBalanceJson
}

func txHistoryJsonBuilder(listTxHistory []models.TxHistory, size int) *Result {
	listTxHistoryJson := make([]*TxHistoryJson, 0)
	for _, tx := range listTxHistory {
		txJson := new(TxHistoryJson)
		txJson.LockTime = tx.LockTime
		txJson.TxHash = tx.TxHash
		txJson.Fee = 0
		if tx.Type == common.TransferTokenType {
			txJson.TokenID = tx.TokenID
			if tx.VOutTokens+tx.TokenFee > tx.VInTokens {
				txJson.Type = common.ReceiveStr
				txJson.Amount = tx.VOutTokens
			}
			if tx.VOutTokens+tx.TokenFee < tx.VInTokens {
				txJson.Fee = tx.TokenFee
				txJson.Type = common.SendStr
				txJson.Amount = tx.VInTokens - tx.VOutTokens - tx.TokenFee
			}

		} else {
			txJson.TokenID = common.PRVID
			if tx.VOutPRVs+tx.Fee > tx.VInPRVs {
				txJson.Type = common.ReceiveStr
				txJson.Amount = tx.VOutPRVs
			}
			if tx.VOutPRVs+tx.Fee <= tx.VInPRVs {
				txJson.Fee = tx.Fee
				txJson.Type = common.SendStr
				txJson.Amount = tx.VInPRVs - tx.VOutPRVs - tx.Fee
			}
		}
		token := getTokenByID(txJson.TokenID)
		txJson.TokenName = token.Name
		txJson.TokenSymbol = token.Symbol
		txJson.TokenDecimal = token.Decimal
		txJson.TokenImage = token.Image

		listTxHistoryJson = append(listTxHistoryJson, txJson)
	}
	return &Result{
		Size:   size,
		Detail: listTxHistoryJson,
	}
}

func coinDetailJsonBuilder(listCoins []models.Coins) []*CoinJson {
	listCoinJson := make([]*CoinJson, 0)
	for _, coin := range listCoins {
		token := getTokenByID(coin.TokenID)
		sn := coin.SerialNumber
		if coin.IsSpent == false {
			sn = ""
		}
		listCoinJson = append(listCoinJson, &CoinJson{
			PublicKey:      coin.PublicKey,
			CoinCommitment: coin.CoinCommitment,
			SNDerivator:    coin.SNDerivator,
			SerialNumber:   sn,
			Value:          coin.Value,
			IsSpent:        coin.IsSpent,
			TokenID:        coin.TokenID,
			TokenSymbol:    token.Symbol,
			TokenName:      token.Name,
			TokenDecimal:   token.Decimal,
			TokenImage:     token.Image,
		})
	}
	return listCoinJson
}

func infoJsonBuilder(account *models.Account, privateKeyStr string, valuePRV, valueUSDT, valueBTC float64) *InfoJson {
	keyWallet, _ := hdwallet.Base58CheckDeserialize(account.PaymentAddress)
	shardID := common.GetShardIDFromPublicKey(keyWallet.KeySet.PaymentAddress.Pk[:])
	infoJson := &InfoJson{
		AccountName:    account.Name,
		PrivateKey:     privateKeyStr,
		PaymentAddress: account.PaymentAddress,
		PublicKey:      account.PublicKey,
		ViewingKey:     account.ViewKey,
		MiningKey:      account.MiningKey,
		Network:        fmt.Sprintf("shard %v %s", shardID, StateM.NetworkManager.NetworkName),
		ValuePRV:       valuePRV,
		ValueUSDT:      valueUSDT,
		ValueBTC:       valueBTC,
	}
	return infoJson
}

func stateJsonBuilder() *StateJson {
	keyWallet, _ := hdwallet.Base58CheckDeserialize(StateM.AccountManage.Account.PaymentAddress)
	shardID := common.GetShardIDFromPublicKey(keyWallet.KeySet.PaymentAddress.Pk[:])
	stateInfo := &StateJson{
		WalletID:    StateM.WalletManager.WalletID,
		AccountID:   StateM.AccountManage.AccountID,
		ShardID:     shardID,
		NetworkName: StateM.NetworkManager.NetworkName,
		NetworkURL:  StateM.NetworkManager.Network.Url,
		BestBlock:   StateM.NetworkManager.BestBlock,
	}
	return stateInfo
}

func txInfoJsonBuilder(tx *models.AutoTxByHash) *TxInfoJson {

	token := &models.Tokens{
		ID:     tx.Result.PrivacyCustomTokenID,
		Symbol: tx.Result.PrivacyCustomTokenSymbol,
		Name:   tx.Result.PrivacyCustomTokenName,
	}

	if tx.Result.Type == common.TransferTokenType {
		token, _ = StateM.NetworkManager.GetTokenByID(tx.Result.PrivacyCustomTokenID)
		if token.Verified == false {
			token.Symbol = token.Symbol + "*"
		}
	}

	txInfo := &TxInfoJson{
		BlockHash:   tx.Result.BlockHash,
		BlockHeight: tx.Result.BlockHeight,
		TxSize:      tx.Result.TxSize,
		Index:       tx.Result.Index,
		ShardID:     tx.Result.ShardID,
		Hash:        tx.Result.Hash,
		Version:     tx.Result.Version,
		Type:        tx.Result.Type,
		LockTime:    tx.Result.LockTime,
		Fee:         tx.Result.Fee,
		ProofDetail: struct {
			InputCoins  []models.AutoCoin
			OutputCoins []models.AutoCoin
		}{tx.Result.ProofDetail.InputCoins, tx.Result.ProofDetail.OutputCoins},
		SigPubKey:                tx.Result.SigPubKey,
		Sig:                      tx.Result.Sig,
		Metadata:                 tx.Result.Metadata,
		PrivacyCustomTokenID:     token.ID,
		PrivacyCustomTokenName:   token.Name,
		PrivacyCustomTokenSymbol: token.Symbol,
		PrivacyCustomTokenProofDetail: struct {
			InputCoins  []models.AutoCoin
			OutputCoins []models.AutoCoin
		}{tx.Result.PrivacyCustomTokenProofDetail.InputCoins,
			tx.Result.PrivacyCustomTokenProofDetail.OutputCoins},
		PrivacyCustomTokenIsPrivacy: tx.Result.PrivacyCustomTokenIsPrivacy,
		PrivacyCustomTokenFee:       tx.Result.PrivacyCustomTokenFee,
		IsInMempool:                 tx.Result.IsInMempool,
		IsInBlock:                   tx.Result.IsInBlock,
		Info:                        tx.Result.Info,
	}
	return txInfo
}

func pdeTradeHistoryJsonBuilder(pdeHistory []*models.PdeTradeHistory, size int) *Result {
	listJson := make([]*PdeHistoryJson, 0)
	for _, tx := range pdeHistory {
		receiveToken := getTokenByID(tx.ToTokenIDStr)
		sendToken := getTokenByID(tx.FromTokenIDStr)

		historyJson := &PdeHistoryJson{
			TraderAddressStr:    tx.TraderAddressStr,
			ReceiveTokenIDStr:   receiveToken.ID,
			ReceiveTokenSymbol:  receiveToken.Symbol,
			ReceiveTokenName:    receiveToken.Name,
			ReceiveTokenDecimal: receiveToken.Decimal,
			ReceiveTokenImage:   receiveToken.Image,
			ReceiverAmount:      tx.ReceiveAmount,
			SendTokenIDStr:      sendToken.ID,
			SendTokenSymbol:     sendToken.Symbol,
			SendTokenName:       sendToken.Name,
			SendTokenDecimal:    sendToken.Decimal,
			SendTokenImage:      sendToken.Image,
			SendAmount:          tx.TradeAmount,
			RequestedTxID:       tx.RequestedTxID,
			BlockHeight:         tx.BlockHeight,
			LockTime:            tx.LockTime,
			ShardID:             tx.ShardID,
			Status:              tx.Status,
		}
		listJson = append(listJson, historyJson)
	}
	result := &Result{
		Size:   size,
		Detail: listJson,
	}
	return result
}

func pdePoolPairPriceJsonBuilder(poolPair *models.PdePoolPairs, exchangeToken string, exchangeAmount, receiveAmount uint64) *PdePoolPairPriceJson {
	token1 := getTokenByID(poolPair.Token1IDStr)
	token2 := getTokenByID(poolPair.Token2IDStr)
	fromToken := token1
	toToken := token2
	fromTokenPoolValue := poolPair.Token1PoolValue
	toTokenPoolValue := poolPair.Token2PoolValue
	if exchangeToken == token2.ID {
		fromToken = token2
		toToken = token1
		fromTokenPoolValue = poolPair.Token2PoolValue
		toTokenPoolValue = poolPair.Token1PoolValue
	}
	return &PdePoolPairPriceJson{
		FromTokenID:        fromToken.ID,
		FromTokenName:      fromToken.Name,
		FromTokenSymbol:    fromToken.Symbol,
		FromTokenDecimal:   fromToken.Decimal,
		FromTokenImage:     fromToken.Image,
		FromTokenPoolValue: fromTokenPoolValue,
		ToTokenID:          toToken.ID,
		ToTokenName:        toToken.Name,
		ToTokenSymbol:      toToken.Symbol,
		ToTokenDecimal:     toToken.Decimal,
		ToTokenImage:       toToken.Image,
		ToTokenPoolValue:   toTokenPoolValue,
		ExchangeAmount:     exchangeAmount,
		ReceiveAmount:      receiveAmount,
		BeaconHeight:       poolPair.BeaconHeight,
	}

}
