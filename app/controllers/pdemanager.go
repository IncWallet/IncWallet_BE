package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/common"
	"incwallet/app/lib/transaction"
	"incwallet/app/models"
	"math"
)

type PdeManager struct {
}

func parsePdeTradeHistoryFromTxHash(txHash string) (*models.PdeTradeHistory, error) {
	//get metadata
	autoTx, metaStr, err := StateM.RpcCaller.GetMetadataFromTxByHash(txHash)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot get metadata from tx %v. Error %v", txHash, err))
	}
	//unmarshal to trade response metadata
	metaTradeResp := new(transaction.PDETradeResponse)
	if err := json.Unmarshal([]byte(metaStr), metaTradeResp); err == nil {
		if metaTradeResp.Type != common.PDETradeResponse1 && metaTradeResp.Type != common.PDETradeResponse2 {
			return nil, nil
		}
		txTradeRequest, err := StateM.RpcCaller.GetAutoTxByHash(metaTradeResp.RequestedTxID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("cannot get trade request detail %v for mint tx %v. Error %v", metaTradeResp.RequestedTxID, txHash, err))
		}
		metaTradeReq := new(transaction.PDETradeRequest)
		if err = json.Unmarshal([]byte(txTradeRequest.Result.Metadata), metaTradeReq); err != nil {
			return nil, errors.New(fmt.Sprintf("cannot get trade request detail %v for mint tx %v. Error %v", metaTradeResp.RequestedTxID, txHash, err))
		}
		receiveAmount := uint64(0)
		if autoTx.Result.Type == common.SalaryPRVType {
			receiveAmount = autoTx.Result.ProofDetail.OutputCoins[0].CoinDetails.Value
		}
		if autoTx.Result.Type == common.TransferTokenType {
			receiveAmount = autoTx.Result.PrivacyCustomTokenProofDetail.OutputCoins[0].CoinDetails.Value
		}

		pdeTxHistory := &models.PdeTradeHistory{
			FromTokenIDStr:   metaTradeReq.TokenIDToSellStr,
			ToTokenIDStr:     metaTradeReq.TokenIDToBuyStr,
			TraderAddressStr: metaTradeReq.TraderAddressStr,
			TradeAmount:      metaTradeReq.SellAmount,
			TradeFee:         metaTradeReq.TradingFee,
			ReceiveAmount:    receiveAmount,
			ShardID:          autoTx.Result.ShardID,
			RequestedTxID:    metaTradeResp.RequestedTxID,
			ResponseTxID:     txHash,
			Type:             metaTradeReq.Type,
			Status:           metaTradeResp.TradeStatus,
			BlockHeight:      autoTx.Result.BlockHeight,
			LockTime:         autoTx.Result.LockTime,
		}
		return pdeTxHistory, nil
	}
	return nil, nil
}

func (pm *PdeManager) GetPool(tokenID1, tokenID2 string) (*models.PdePoolPairs, error) {
	pool := new(models.PdePoolPairs)
	query1 := bson.M{
		"token1idstr": tokenID1,
		"token2idstr": tokenID2,
	}
	query2 := bson.M{
		"token1idstr": tokenID2,
		"token2idstr": tokenID1,
	}
	query := bson.M{"$or": []bson.M{query1, query2}}
	if err := database.PoolPairs.Find(query).Sort("-beaconheight").Limit(1).One(&pool); err != nil {
		return nil, err
	}
	return pool, nil
}

func (pm *PdeManager) GetRate(amount, fee uint64, fromTokenID, toTokenID string, poolPair *models.PdePoolPairs) (uint64, error) {
	fromToken, err := StateM.NetworkManager.GetTokenByID(fromTokenID)
	if err != nil {
		return 0, errors.New("cannot get token info from fromtokenid")
	}
	toToken, err := StateM.NetworkManager.GetTokenByID(toTokenID)
	if err != nil {
		return 0, errors.New("cannot get token info from totokenid")
	}

	poolFromValue := poolPair.Token1PoolValue
	poolToValue := poolPair.Token2PoolValue
	if poolPair.Token1IDStr == toTokenID {
		poolFromValue = poolPair.Token2PoolValue
		poolToValue = poolPair.Token1PoolValue
	}
	realPoolFromValue := float64(poolFromValue) / math.Pow10(fromToken.Decimal)
	realPoolToValue := float64(poolToValue) / math.Pow10(toToken.Decimal)
	realAmount := float64(amount - fee) / math.Pow10(fromToken.Decimal)
	newPoolFromValue := realPoolFromValue + realAmount
	newPoolToValue := realPoolFromValue * realPoolToValue / newPoolFromValue
	getAmount := realPoolToValue - newPoolToValue
	return uint64(getAmount * math.Pow10(toToken.Decimal)), nil
}

func (pm *PdeManager) UpdatePdePoolPairsFromChain() error {
	if flag, _ := IsStateFull() ; !flag{
		return nil
	}

	//if StateM.NetworkManager.Network.Name == common.Mainnet {
	//	return nil
	//}

	tmpBeaconHeight := uint64(0)
	tmp := new(models.PdePoolPairs)
	if err := database.PoolPairs.Find(bson.M{}).Sort("-beaconheight").One(&tmp); err == nil {
		tmpBeaconHeight = tmp.BeaconHeight
	}
	if tmpBeaconHeight >= StateM.NetworkManager.BeaconState.Height {
		return nil
	}

	mapPoolPairs, err := StateM.RpcCaller.GetPdePoolParis(StateM.NetworkManager.BeaconState.Height)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot get pde pool pair from chain. Error %v", err))
	}

	listPoolPairs := make([]interface{}, 0)
	for _, pair := range mapPoolPairs {
		poolPairs := &models.PdePoolPairs{
			BeaconHeight:    StateM.NetworkManager.BeaconState.Height,
			Token1IDStr:     pair.Token1IDStr,
			Token2IDStr:     pair.Token2IDStr,
			Token1PoolValue: pair.Token1PoolValue,
			Token2PoolValue: pair.Token2PoolValue,
		}
		listPoolPairs = append(listPoolPairs, *poolPairs)
	}

	pc := database.PoolPairs.Bulk()
	pc.Insert(listPoolPairs...)
	_, err = pc.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("cannot insert bulk of pde pool pair from chain. Error %v", err))
	}
	return nil
}

func (pm *PdeManager) UpdatePdeTradeHistoryFromChain() error {
	if flag, _ := IsStateFull() ; !flag{
		return nil
	}

	if StateM.NetworkManager.Network.Name == common.Mainnet {
		return nil
	}

	mapBestBlock, err := StateM.RpcCaller.GetBestBlockInfo()
	if err != nil {
		return errors.New(fmt.Sprintf("cannot get best block. Error %v", err))
	}
	txsChan := make(chan []string)
	for i:= 0; i < common.ShardNumber; i ++ {
		if StateM.NetworkManager.BestBlock[i] < mapBestBlock[i].Height {
			go func(blockHash string, shardID int, txsChan chan []string) {
				txs, err := StateM.RpcCaller.GetTxsFromRetrieveBlock(blockHash)
				if err != nil {
					revel.AppLog.Warnf("cannot get tx from retrieve block. error %v", err)
					txsChan <- []string{}
				} else {
					txsChan <- txs
				}
			}(mapBestBlock[i].Hash, i, txsChan)
		}
	}
	listTx := make([]string, 0)
	for i:= 0; i < common.ShardNumber; i ++ {
		if StateM.NetworkManager.BestBlock[i] < mapBestBlock[i].Height {
			txs := <-txsChan
			if len(txs) > 0 {
				listTx = append(listTx, txs...)
			}
		}
	}
	pdeHistoryChan := make(chan *models.PdeTradeHistory)
	for _, tx := range listTx {
		go func(txHash string, pdeTxChan chan *models.PdeTradeHistory) {
			pdeTx, err := parsePdeTradeHistoryFromTxHash(txHash)
			if err != nil {
				revel.AppLog.Warnf("cannot parse Pde Trade from tx %v. Error %v", txHash, err)
			}
			pdeTxChan <- pdeTx
		}(tx, pdeHistoryChan)

	}
	listPdeTxHistory := make([]interface{}, 0)
	for range listTx {
		tx := <- pdeHistoryChan
		if tx != nil {
			revel.AppLog.Infof("detected trade tx %v", tx.RequestedTxID)
			listPdeTxHistory = append(listPdeTxHistory, tx)
		}
	}
	if len(listPdeTxHistory) > 0 {
		pc := database.PdeHistory.Bulk()
		pc.Insert(listPdeTxHistory...)
		_, err = pc.Run()
		if err != nil {
			return errors.New(fmt.Sprintf("cannot insert bulk of list tx history from chain. Error %v", err))
		}
	}
	for i:= 0; i < common.ShardNumber; i ++ {
		StateM.NetworkManager.BestBlock[i] = mapBestBlock[i].Height
	}
	return nil
}

//func (pm *PdeManager) UpdatePdeTradeHistoryFromChainOld(limit int) error {
//	if flag, _ := IsStateFull() ; !flag{
//		return nil
//	}
//	if StateM.NetworkManager.Network.Name == common.Mainnet {
//		return nil
//	}
//
//	lastBeaconHeight := StateM.NetworkManager.BeaconState.Height
//	fromBeaconHeight := StateM.NetworkManager.BestBlock[-1]
//	tmp := new(models.PdeTradeHistory)
//	if err := database.PdeHistory.Find(bson.M{}).Sort("-beaconheight").One(&tmp); err == nil {
//		if tmp.BeaconHeight > fromBeaconHeight {
//			fromBeaconHeight = tmp.BeaconHeight
//		}
//	}
//
//	if uint64(limit) < lastBeaconHeight- fromBeaconHeight{
//		fromBeaconHeight = lastBeaconHeight - uint64(limit)
//	}
//	errorChan := make(chan error)
//	for index := fromBeaconHeight + 1; index <= lastBeaconHeight; index ++ {
//		go func(beaconHeight uint64, errorChan chan error) {
//			tmpHistory, err := StateM.RpcCaller.GetPdeHistoryFromBeaconIns(beaconHeight)
//			if err != nil {
//				errorChan <- err
//			} else {
//				pc := database.PdeHistory.Bulk()
//				bulkPdeHistory := make([]interface{}, 0)
//				for i := range tmpHistory {
//					bulkPdeHistory = append(bulkPdeHistory, &tmpHistory[i])
//				}
//				if len(tmpHistory) > 0 {
//					pc.Insert(bulkPdeHistory...)
//					_, err := pc.Run()
//					if err != nil {
//						errorChan <- err
//					} else {
//						errorChan <- nil
//					}
//				 } else {
//					errorChan <- nil
//				}
//			}
//		}(index, errorChan)
//	}
//
//	for index := fromBeaconHeight + 1; index <= lastBeaconHeight; index ++ {
//		err := <- errorChan
//		if err != nil {
//			revel.AppLog.Warnf("cannot insert bulk of pde history %v", err)
//		}
//	}
//	StateM.NetworkManager.BestBlock[-1] = lastBeaconHeight
//	return nil
//}
