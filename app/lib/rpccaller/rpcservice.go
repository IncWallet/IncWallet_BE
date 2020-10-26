package rpccaller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"incwallet/app/lib/crypto"
	"incwallet/app/lib/transaction"
	"incwallet/app/models"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
)

func (this *RPCService) GetAllToken() ([]*models.Tokens, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("did not init RPCService")
	}

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "listprivacycustomtoken",
		"params": [],
		"id": 1
	}`)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	var queryResult AutoListToken
	if err := json.Unmarshal(byteData, &queryResult); err != nil {
		return nil, err
	}

	if queryResult.Error == nil {
		listToken := make([]*models.Tokens, 0)
		for _, autoToken := range queryResult.Result.ListCustomToken {
			token := &models.Tokens{
				ID:       autoToken.ID,
				Symbol:   autoToken.Symbol,
				Name:     autoToken.Name,
				Decimal:  0,
				EDecimal: 0,
				ESymbol:  "",
				Verified: false,
				Network:  this.Network,
			}
			listToken = append(listToken, token)
		}
		return listToken, nil
	}
	return nil, errors.New(fmt.Sprintf("Error: %v", queryResult.Error))
}

func (this *RPCService) GetPdeHistory(beaconHeight uint64) ([]models.PdeTradeHistory, error) {
	query := fmt.Sprintf(`{
				"jsonrpc": "1.0",
				"method": "extractpdeinstsfrombeaconblock",
				"params": [
					{"BeaconHeight":%v}],
				"id": 1
			}`,beaconHeight)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		revel.AppLog.Warnf("cannot send extracted pde history request. error %v", err)
		return nil, errors.New(fmt.Sprintf("Error from send query : %v", err))
	}
	var queryResult AutoPdeTradeHistory
	if err = json.Unmarshal(byteData, &queryResult); err != nil {
		revel.AppLog.Warnf("cannot unmarshal extracted pde history. error %v", err)
		return nil, errors.New(fmt.Sprintf("Error from unmarshal: %v", err))
	}
	pdeHistory := make([]models.PdeTradeHistory, 0)
	if queryResult.Error == nil{
		if 	queryResult.Result.PDETrades != nil{
			for _,tx := range queryResult.Result.PDETrades{
				//get more trade info
				txTradeRequest, err := this.GetAutoTxByHash(tx.RequestedTxID)
				if err != nil {
					revel.AppLog.Warnf("cannot get trade request detail %v. Error %v", tx.RequestedTxID, err)
					return nil, errors.New(fmt.Sprintf("cannot get trade request detail %v. Error %v", tx.RequestedTxID, err))
				}
				meta := new(transaction.PDETradeRequest)
				if err = json.Unmarshal([]byte(txTradeRequest.Result.Metadata), meta); err != nil {
					revel.AppLog.Warnf("cannot get metadate in trade request detail %v. Error %v", tx.RequestedTxID, err)
					return nil, errors.New(fmt.Sprintf("cannot get metadate in trade request detail %v. Error %v", tx.RequestedTxID, err))
				}
				tx.TradeAmount = meta.SellAmount
				tx.TradeFee = meta.TradingFee
				pdeHistory = append(pdeHistory,tx)
			}
		}
	}
	return pdeHistory, nil
}

func (this *RPCService) GetCommitteeInfo() ([]*models.Committee,error) {
	if len(this.Url)== 0 {
		return nil, errors.New("Did not init RPCService")
	}
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "getbeaconbeststate",
		"params": [],
		"id": 1
	}`)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	var queryResult AutoBeaconBestState
	if err = json.Unmarshal(byteData, &queryResult); err != nil {
		return nil, err
	}
	if queryResult.Error == nil {
		listCommitte := make([]*models.Committee, 0)

		for _, stakingKey := range queryResult.Result.CandidateShardWaitingForNextRandom {
			committe := &models.Committee{
				Epoch: queryResult.Result.Epoch,
				Key:     stakingKey,
				Index:   -1,
				Role:    common.CandidateRole,
				ShardId: -1,
			}
			listCommitte = append(listCommitte, committe)
		}
		for shardId, stakingKeys := range queryResult.Result.ShardPendingValidator {
			for index, key := range stakingKeys {
				id, _ := strconv.Atoi(shardId)
				committe := &models.Committee{
					Epoch: queryResult.Result.Epoch,
					Key:     key,
					Role:    common.PendingRole,
					Index:   index,
					ShardId: id,
				}
				listCommitte = append(listCommitte, committe)
			}
		}
		for shardId, stakingKeys := range queryResult.Result.ShardCommittee {
			for index, key := range stakingKeys {
				id, _ := strconv.Atoi(shardId)
				committe := &models.Committee{
					Epoch: queryResult.Result.Epoch,
					Key:     key,
					Role:    common.ValidatorRole,
					Index:   index,
					ShardId: id,
				}
				listCommitte = append(listCommitte, committe)
			}
		}
		return listCommitte, nil
	}
	return nil, errors.New(fmt.Sprintf("Error: %v", queryResult.Error))
}

func (this *RPCService) GetRewardAmount() (map[string]map[string]uint64, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("Did not init RPCService")
	}
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "listrewardamount",
		"params": [],
		"id": 1
	}`)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	var queryResult AutoRewardAmount
	if err = json.Unmarshal(byteData, &queryResult); err != nil {
		return nil, err
	}
	if queryResult.Error == nil {
		return queryResult.Result, nil
	}
	return nil, errors.New(fmt.Sprintf("Error: %v", queryResult.Error))
}

func (this *RPCService) SendRawTx(txBase58Check string) ([]byte, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("Did not init RPCService")
	}

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0", 
		"method": "sendtransaction", 
		"params": ["%s"], 
		"id": 1}
	`, txBase58Check)

	resBytes, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot send request. Error %v", err))
	}
	var result AutoSendRawTxResult
	err = json.Unmarshal(resBytes, &result)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot send request. Error %v", err))
	}
	if result.Error != nil {
		return nil, errors.New(fmt.Sprintf("cannot send request. Error %v", result.Error))
	}
	return resBytes, err
}

func (this *RPCService) SendRawTxToken(txBase58Check string) ([]byte, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("Did not init RPCService")
	}

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0", 
		"method": "sendrawprivacycustomtokentransaction", 
		"params": ["%s"], 
		"id": 1}
	`, txBase58Check)

	resBytes, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot send request. Error %v", err))
	}
	var result AutoSendRawTokenTxResult
	err = json.Unmarshal(resBytes, &result)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot send request. Error %v", err))
	}
	if result.Error != nil {
		return nil, errors.New(fmt.Sprintf("cannot send request. Error %v", result.Error))
	}
	return resBytes, err
}

func (this *RPCService) GetRandomCommitments(paymentAddStr string, inputCoinsStr string, tokenIDStr string) ([]uint64, []uint64, []string, error) {
	if len(this.Url) == 0 {
		return nil, nil, nil, errors.New("Did not init RPCService")
	}

	query := fmt.Sprintf(`{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "randomcommitments",
		"params": [
			"%s",
			%s,
			"%s"
		]
	}`, paymentAddStr, inputCoinsStr, tokenIDStr)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, nil, nil, errors.New(fmt.Sprintf("Cannot send post request. Error: %v", err))
	}
	result := new(AutoRandomCommitments)
	if err := json.Unmarshal(byteData, result); err != nil {
		return nil, nil, nil, errors.New(fmt.Sprintf("Cannot unmarshal result. Error: %v", err))
	}
	return result.Result.CommitmentIndices, result.Result.MyCommitmentIndexs, result.Result.Commitments, nil
}

func (this *RPCService) SendPostRequestWithQuery(query string) ([]byte, error) {
	if len(this.Url) == 0 {
		return []byte{}, errors.New("Debugtool has not set mainnet or testnet")
	}
	var jsonStr = []byte(query)
	req, _ := http.NewRequest("POST", this.Url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, err
		}
		return body, nil
	}
}

// ==============TRANSACTION ===============
// Parse from byte to AutoTxByHash
func ParseAutoTxHashFromBytes(b []byte) (*models.AutoTxByHash, error) {
	data := new(models.AutoTxByHash)
	err := json.Unmarshal(b, data)
	if err != nil {
		return nil, err
	}

	if data.Error != nil {
		return nil, errors.New(fmt.Sprintf("%v", data.Error))
	}
	return data, nil
}

func (this *RPCService) GetTransactionByHash(txHash string) ([]byte, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("did not init RPCService")
	}
	query := fmt.Sprintf(`{
		"jsonrpc":"1.0",
		"method":"gettransactionbyhash",
		"params":["%s"],
		"id":1
	}`, txHash)
	return this.SendPostRequestWithQuery(query)
}

// Query the RPC server then return the AutoTxByHash
func (this *RPCService) GetAutoTxByHash(txHash string) (*models.AutoTxByHash, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("did not init RPCService")
	}
	b, err := this.GetTransactionByHash(txHash)
	if err != nil {
		return nil, err
	}
	autoTx, txError := ParseAutoTxHashFromBytes(b)
	if txError != nil {
		return nil, txError
	}
	return autoTx, nil
}

// Get metadate from transaction
func (this *RPCService) GetMetadataFromTxByHash(txHash string) (*models.AutoTxByHash, string, error) {
	autoTx, err := this.GetAutoTxByHash(txHash)
	if err != nil {
		return nil, "", err
	}
	return autoTx, autoTx.Result.Metadata, nil
}

// Get only the proof of transaction requiring the txHash
func (this *RPCService) GetProofTransactionByHash(txHash string) (string, error) {
	tx, err := this.GetAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.Proof, nil
}

// Get only the Sig of transaction requiring the txHash
func (this *RPCService) GetSigTransactionByHash(txHash string) (string, error) {
	tx, err := this.GetAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.Sig, nil
}

// Get only the BlockHash of transaction requiring the txHash
func (this *RPCService) GetBlockHashTransactionByHash(txHash string) (string, error) {
	tx, err := this.GetAutoTxByHash(txHash)
	if err != nil {
		return "", err
	}
	return tx.Result.BlockHash, nil
}

// Get only the BlockHeight of transaction requiring the txHash
func (this *RPCService) GetBlockHeightTransactionByHash(txHash string) (uint64, error) {
	tx, err := this.GetAutoTxByHash(txHash)
	if err != nil {
		return uint64(0), err
	}
	return tx.Result.BlockHeight, nil
}

// ==============ACCOUNT PRV ===============

func (this *RPCService) GetListReceiveTxHash(paymentAddStr string) (map[string]string, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("Did not init RPCService")
	}

	query := fmt.Sprintf(`{  
	   "jsonrpc":"1.0",
	   "method":"gettransactionhashbyreceiver",
	   "params":["%s"],
	   "id":1
	}`, paymentAddStr)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	txHashHistory := new(AutoTxHashHistory)
	if err := json.Unmarshal(byteData, txHashHistory); err != nil {
		return nil, err
	}
	if txHashHistory.Error == nil {
		txsHash := make(map[string]string)
		for _, txsHashByShard := range txHashHistory.Result {
			for _, item := range txsHashByShard {
				txsHash[item] = item
			}
		}
		return txsHash, nil
	}
	return nil, errors.New(fmt.Sprintf("Error: %v", txHashHistory.Error))
}

func (this *RPCService) GetListOutputCoins(paymentAddStr string, readOnlyKey []byte, tokenID string) ([]models.Coins, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("Did not init RPCService")
	}

	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "listoutputcoins",
		"params": [
			0,
			999999,
			[
				{
			  "PaymentAddress": "%s",
			  "StartHeight": 0
				}
			],
			"%s"
		  ],
		"id": 1
	}`, paymentAddStr, tokenID)

	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	listOutCoinResult := new(AutoListOutputPRV)
	if err := json.Unmarshal(byteData, listOutCoinResult); err != nil {
		return nil, err
	}
	if listOutCoinResult.Error == nil {
		coinList := make([]models.Coins, len(listOutCoinResult.Result.Outputs[paymentAddStr]))
		for i, coin := range listOutCoinResult.Result.Outputs[paymentAddStr] {
			coin.TokenID = tokenID
			ciphertextBytes, _, err := base58.Base58Check{}.Decode(coin.CoinDetailsEncrypted)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("cannot base58 decode coin detail encrypted. Error %v", err))
			}

			ciphertext := new(crypto.HybridCipherText)
			err = ciphertext.SetBytes(ciphertextBytes)
			if err != nil {
				//revel.AppLog.Warnf("cannot unmarshal coin detail encrypted. Error %v", err)
				coinList[i] = coin
			} else {
				msg, err := crypto.HybridDecrypt(ciphertext, new(crypto.Scalar).FromBytesS(readOnlyKey))
				if err != nil {
					return nil, errors.New(fmt.Sprintf("cannot decrypt coin info. Error %v", err))
				}
				coin.Randomness = base58.Base58Check{}.Encode(new(crypto.Scalar).FromBytesS(msg[0:crypto.Ed25519KeySize]).ToBytesS(), common.ZeroByte)
				coin.Value = fmt.Sprintf("%v",new(big.Int).SetBytes(msg[crypto.Ed25519KeySize:]).Uint64())
				coinList[i] = coin
			}
		}
		return coinList, nil
	}
	return nil, errors.New(fmt.Sprintf("Error: %v", listOutCoinResult.Error))
}

func (this *RPCService) GetListOutputCoinsInBytes(paymentAddStr string, readOnlyKey []byte, tokenID string) ([]byte, error) {
	coins, err := this.GetListOutputCoins(paymentAddStr, readOnlyKey, tokenID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(coins)
}

func (this *RPCService) HasSerialNumbers(paymentAddStr string, lsSerialNumbers []string, tokenID string) (map[string]bool, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("Did not init RPCService")
	}
	lsSN, err := json.Marshal(lsSerialNumbers)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Cannot Marshal lsSerialNumber. Error %v", err))
	}

	query := fmt.Sprintf(`{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "hasserialnumbers",
		"params": [
			"%s",
			%s,
			"%s"
		]
	}`, paymentAddStr, string(lsSN), tokenID)

	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	result := new(AutoHasSerialNumber)
	if err := json.Unmarshal(byteData, result); err != nil {
		return nil, err
	}

	if len(result.Result) != len(lsSerialNumbers) {
		return nil, errors.New("Invalid HasSerialNumber query result")
	}
	snMap := make(map[string]bool, len(lsSerialNumbers))
	for index, sn := range lsSerialNumbers {
		snMap[sn] = result.Result[index]
	}
	return snMap, nil
}

// ==============PDE  ===============

func (this *RPCService) GetBeaconHeight() (*models.BeaconState, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("Did not init RPCService")
	}
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "getblockchaininfo",
		"params": [],
		"id": 1
	}`)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	var queryResult AutoBeaconHeight
	if err = json.Unmarshal(byteData, &queryResult); err != nil {
		return nil, err
	}
	if queryResult.Error == nil {
		beaconState := &models.BeaconState{
			Height:              queryResult.Result.BestBlock["-1"].Height,
			Epoch:               queryResult.Result.BestBlock["-1"].Epoch,
			RemainingBlockEpoch: queryResult.Result.BestBlock["-1"].RemainingBlockEpoch,
		}
		return beaconState, nil
	}
	return nil, errors.New(fmt.Sprintf("Error: %v", queryResult.Error))
}

func (this *RPCService) GetPdePoolParis(beaconHeight uint64) (map[string]models.PdePoolPairs, error) {
	if len(this.Url) == 0 {
		return nil, errors.New("did not init RPCService")
	}
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "getpdestate",
		"params": [
			{"BeaconHeight":%v}],
		"id": 1
	}`, beaconHeight)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		return nil, err
	}
	var queryResult AutoPdeState
	if err = json.Unmarshal(byteData, &queryResult); err != nil {
		return nil, err
	}
	if queryResult.Error == nil {
		return queryResult.Result.PDEPoolPairs, nil
	}
	return nil, errors.New(fmt.Sprintf("Error: %v", queryResult.Error))

}

func (this *RPCService) GetPdeHistoryFromBeaconIns(beaconHeight uint64) ([]models.PdeTradeHistory, error) {
	query := fmt.Sprintf(`{
				"jsonrpc": "1.0",
				"method": "extractpdeinstsfrombeaconblock",
				"params": [
					{"BeaconHeight":%v}],
				"id": 1
			}`,beaconHeight)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		revel.AppLog.Warnf("cannot send extracted pde history request. error %v", err)
		return nil, errors.New(fmt.Sprintf("Error from send query : %v", err))
	}
	var queryResult AutoPdeTradeHistory
	if err = json.Unmarshal(byteData, &queryResult); err != nil {
		revel.AppLog.Warnf("cannot unmarshal extracted pde history. error %v", err)
		return nil, errors.New(fmt.Sprintf("Error from unmarshal: %v", err))
	}
	pdeHistory := make([]models.PdeTradeHistory, 0)
	if queryResult.Error == nil{
		if 	queryResult.Result.PDETrades != nil{
			for _,tx := range queryResult.Result.PDETrades{
				//get more trade info
				txTradeRequest, err := this.GetAutoTxByHash(tx.RequestedTxID)
				if err != nil {
					revel.AppLog.Warnf("cannot get trade request detail %v. Error %v", tx.RequestedTxID, err)
					return nil, errors.New(fmt.Sprintf("cannot get trade request detail %v. Error %v", tx.RequestedTxID, err))
				}
				meta := new(transaction.PDETradeRequest)
				if err = json.Unmarshal([]byte(txTradeRequest.Result.Metadata), meta); err != nil {
					revel.AppLog.Warnf("cannot get metadate in trade request detail %v. Error %v", tx.RequestedTxID, err)
					return nil, errors.New(fmt.Sprintf("cannot get metadate in trade request detail %v. Error %v", tx.RequestedTxID, err))
				}
				tx.TradeAmount = meta.SellAmount
				tx.TradeFee = meta.TradingFee
				pdeHistory = append(pdeHistory,tx)
			}
		}
	}
	return pdeHistory, nil
}

func (this *RPCService) GetBestBlockInfo() (map[int]AutoBestBlockDetail, error) {
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "getbestblock",
		"params": "",
		"id": 1
	}`)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		revel.AppLog.Warnf("cannot send get best block request. error %v", err)
		return nil, errors.New(fmt.Sprintf("Error from send query : %v", err))
	}

	var queryResult AutoBestBlock
	if err = json.Unmarshal(byteData, &queryResult); err != nil {
		revel.AppLog.Warnf("cannot unmarshal best block result. error %v", err)
		return nil, errors.New(fmt.Sprintf("Error from unmarshal: %v", err))
	}
	mapBestBlock := make(map[int]AutoBestBlockDetail)
	if queryResult.Error != nil {
		return nil, errors.New(fmt.Sprintf("%v", queryResult.Error))
	}
	for i := 0; i < 8; i ++ {
		index := fmt.Sprintf("%v", i)
		if blockInfo , found := queryResult.Result.BestBlocks[index]; !found {
			return nil, err
		} else {
			mapBestBlock[i] = blockInfo
		}
	}

	return mapBestBlock, nil
}

func (this *RPCService) GetTxsFromRetrieveBlock(blockHash string) ([]string, error) {
	query := fmt.Sprintf(`{
		"jsonrpc": "1.0",
		"method": "retrieveblock",
		"params": ["%s","1"],
		"id": 1
	}`, blockHash)
	byteData, err := this.SendPostRequestWithQuery(query)
	if err != nil {
		revel.AppLog.Warnf("cannot send retrieve block request. error %v", err)
		return nil, errors.New(fmt.Sprintf("Error from send query : %v", err))
	}

	var queryResult AutoRetrieveBlock
	if err = json.Unmarshal(byteData, &queryResult); err != nil {
		revel.AppLog.Warnf("cannot unmarshal retrieve block result. error %v", err)
		return nil, errors.New(fmt.Sprintf("Error from unmarshal: %v", err))
	}
	if queryResult.Error != nil {
		revel.AppLog.Warnf("%v", queryResult.Error)
		return nil, errors.New(fmt.Sprintf("%v", queryResult.Error))
	}
	return queryResult.Result.TxHashes, nil
}
