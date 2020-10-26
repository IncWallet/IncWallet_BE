package rpccaller

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/stretchr/testify/assert"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"incwallet/app/lib/crypto"
	"incwallet/app/lib/hdwallet"
	"incwallet/app/lib/transaction"
	"incwallet/app/models"
	"testing"
)
// temp account
var privateKeyStr = "112t8rnbCjhDpBQjjNJQ2bABVAbkZC2GaFvbTw7kCaN7RyqLC9Pwh7v7bNoLh5PqDcj2SYDk1HNqoKDwcFdeDRhdcB2mjGuRdnykkqka5HV1"
var paymentAddStr = "12S4CvTpc5wHbNXyvdzPktYMGCZcj3pSTCUS73DCShRz3kTmsjorQHdhjmhTMMqLfthMMFTpwVDBbY7kyHYeJcB3wMiK56mum1TaBkU"
var readKey = "13hfo5qjMt1gc6CTYiAkbaXHBvbow4oxZQgFN1snXeZDy8QBDBY7VrFPhH9feVg1pYkvRd8tkb9ui6C2b6vBGjnDVnumgkuE5dhzpyy"
var tokenID = "0000000000000000000000000000000000000000000000000000000000000004"

func TestRPCService_GetAutoTxByHash(t *testing.T) {
	hash := ""
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	tx, err := rpcCaller.GetAutoTxByHash(hash)
	fmt.Println(err)
	metaDataStr := tx.Result.Metadata
	meta := new(transaction.PDETradeRequest)
	json.Unmarshal([]byte(metaDataStr), meta)
	fmt.Println(meta.TraderAddressStr)
	b, err := rpcCaller.GetTransactionByHash(hash)
	var tmp models.AutoTxByHash
	err = json.Unmarshal(b, &tmp)
	fmt.Println(err)
}

func TestParsePrivateKey (t *testing.T) {
	keyWallet, _ := hdwallet.Base58CheckDeserialize(privateKeyStr)
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	paymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)
	viewingKeyStr := keyWallet.Base58CheckSerialize(wallet.ReadonlyKeyType)
	miningKeyStr := hdwallet.GenerateMiningKey(keyWallet.KeySet.PrivateKey)

	fmt.Println(base58.Base58Check{}.Encode(keyWallet.KeySet.PaymentAddress.Pk, common.ZeroByte))
	fmt.Println(paymentAddStr)
	fmt.Println(viewingKeyStr)
	fmt.Println(miningKeyStr)
	fmt.Println()
}

func TestRPCService_GetListOutputCoins(t *testing.T) {
	rpcCaller := new(RPCService)

	rpcCaller.InitTestnet(common.URLTestnet)
	keyWallet, _ := hdwallet.Base58CheckDeserialize(readKey)
	coinListBytes, err := rpcCaller.GetListOutputCoinsInBytes(paymentAddStr, keyWallet.KeySet.ReadonlyKey.Rk[:], tokenID)
	if err != nil {
		panic(err)
	}

	var listCoins []*models.Coins
	if err := json.Unmarshal(coinListBytes, &listCoins); err != nil {
		panic(err)
	}
	lsSn := make([]string, 0)
	for i, coin := range listCoins {
		kw, _ := wallet.Base58CheckDeserialize(privateKeyStr)
		snd, _, _ := base58.Base58Check{}.Decode(coin.SNDerivator)
		sn := crypto.GenerateSerialNumber(kw.KeySet.PrivateKey, snd)
		snStr := string(base58.Base58Check{}.Encode(sn, common.ZeroByte))
		listCoins[i].SerialNumber = snStr
		lsSn = append(lsSn,  snStr)
	}
	mapSerialNumbers, _ := rpcCaller.HasSerialNumbers(paymentAddStr, lsSn, tokenID)

	inputCoins := make([]*models.Coins, 0)
	for _, coin := range listCoins {
		if mapSerialNumbers[coin.SerialNumber] == false {
			inputCoins = append(inputCoins, coin)
		}
	}
	safeInputCoins := make([]*models.Coins, len(inputCoins))
	for index, coin := range inputCoins {
		safeInputCoins[index] = coin
		safeInputCoins[index].SerialNumber = ""
		safeInputCoins[index].Value = string(0)
	}

	safeInputCoinsData, _ := json.Marshal(inputCoins)
	indexes, myIndexes, commitments, _ := rpcCaller.GetRandomCommitments(paymentAddStr, string(safeInputCoinsData), tokenID)
	fmt.Println(indexes)
	fmt.Println(myIndexes)
	fmt.Println(commitments)

}

func TestGetOutputCoins(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	keyWallet, _ := hdwallet.Base58CheckDeserialize(readKey)
	coinListBytes, err := rpcCaller.GetListOutputCoinsInBytes(paymentAddStr, keyWallet.KeySet.ReadonlyKey.Rk[:], tokenID)
	fmt.Println(err)

	var listCoins []*models.Coins
	err = json.Unmarshal(coinListBytes, &listCoins)
	fmt.Println(err)

}


func TestGetUnspentCoins(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	keyWallet, _ := hdwallet.Base58CheckDeserialize(readKey)
	coinListBytes, err := rpcCaller.GetListOutputCoinsInBytes(paymentAddStr, keyWallet.KeySet.ReadonlyKey.Rk[:], tokenID)
	assert.Equal(t, nil, err)

	var listCoins []*models.Coins
	err = json.Unmarshal(coinListBytes, &listCoins)
	assert.Equal(t, nil, err)

	listSerialNumbers := make([]string, 0)
	for i, coin := range listCoins {
		key, _ := wallet.Base58CheckDeserialize(privateKeyStr)
		snd, _, _ := base58.Base58Check{}.Decode(coin.SNDerivator)
		sn := crypto.GenerateSerialNumber(key.KeySet.PrivateKey, snd)
		snStr := string(base58.Base58Check{}.Encode(sn, common.ZeroByte))
		listCoins[i].SerialNumber = snStr
		listSerialNumbers = append(listSerialNumbers,  snStr)
	}
	mapSerialNumbers, err := rpcCaller.HasSerialNumbers(paymentAddStr, listSerialNumbers, tokenID)
	assert.Equal(t, nil, err)

	listInputCoins := make([]*models.Coins, 0)
	for _, coin := range listCoins {
		if mapSerialNumbers[coin.SerialNumber] == false {
			listInputCoins = append(listInputCoins, coin)
		}
	}

	hidenInputCoins := make([]*models.Coins, len(listInputCoins))
	for index, coin := range listInputCoins {
		hidenInputCoins[index] = coin
		hidenInputCoins[index].SerialNumber = ""
		hidenInputCoins[index].Value = string(0)
	}
	b, _ := json.Marshal(hidenInputCoins)
	fmt.Println(string(b))
}

func TestGetUnspentCoinsOnAmount(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	keyWallet, _ := hdwallet.Base58CheckDeserialize(readKey)
	coinListBytes, err := rpcCaller.GetListOutputCoinsInBytes(paymentAddStr, keyWallet.KeySet.ReadonlyKey.Rk[:], tokenID)
	assert.Equal(t, nil, err)

	var listCoins []*models.Coins
	err = json.Unmarshal(coinListBytes, &listCoins)
	assert.Equal(t, nil, err)

	listSerialNumbers := make([]string, 0)
	for i, coin := range listCoins {
		key, _ := wallet.Base58CheckDeserialize(privateKeyStr)
		snd, _, _ := base58.Base58Check{}.Decode(coin.SNDerivator)
		sn := crypto.GenerateSerialNumber(key.KeySet.PrivateKey, snd)
		snStr := string(base58.Base58Check{}.Encode(sn, common.ZeroByte))
		listCoins[i].SerialNumber = snStr
		listSerialNumbers = append(listSerialNumbers,  snStr)
	}
	mapSerialNumbers, err := rpcCaller.HasSerialNumbers(paymentAddStr, listSerialNumbers, tokenID)
	assert.Equal(t, nil, err)

	listInputCoins := make([]*models.Coins, 0)
	for _, coin := range listCoins {
		if mapSerialNumbers[coin.SerialNumber] == false {
			listInputCoins = append(listInputCoins, coin)
		}
	}

	hidenCoins := make([]*models.Coins, len(listInputCoins))
	for index, coin := range listInputCoins {
		hidenCoins[index] = coin
		hidenCoins[index].SerialNumber = ""
		hidenCoins[index].Value = string(0)
	}
	hidenCoinsBytes, _ := json.Marshal(hidenCoins)

	cmIndexes, myCmIndexes, commitments, err := rpcCaller.GetRandomCommitments(paymentAddStr, string(hidenCoinsBytes), tokenID)
	assert.Equal(t, nil, err)
	fmt.Println(myCmIndexes)
	for i := range cmIndexes {
		fmt.Println(commitments[i], cmIndexes[i])
	}
}

func TestChooseBestInputCoin(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	keyWallet, _ := hdwallet.Base58CheckDeserialize(readKey)
	coinListBytes, err := rpcCaller.GetListOutputCoinsInBytes(paymentAddStr, keyWallet.KeySet.ReadonlyKey.Rk[:], tokenID)
	assert.Equal(t, nil, err)

	var listCoins []*models.Coins
	err = json.Unmarshal(coinListBytes, &listCoins)
	assert.Equal(t, nil, err)

	listSerialNumbers := make([]string, 0)
	for i, coin := range listCoins {
		key, _ := wallet.Base58CheckDeserialize(privateKeyStr)
		snd, _, _ := base58.Base58Check{}.Decode(coin.SNDerivator)
		sn := crypto.GenerateSerialNumber(key.KeySet.PrivateKey, snd)
		snStr := string(base58.Base58Check{}.Encode(sn, common.ZeroByte))
		listCoins[i].SerialNumber = snStr
		listSerialNumbers = append(listSerialNumbers,  snStr)
	}
	mapSerialNumbers, err := rpcCaller.HasSerialNumbers(paymentAddStr, listSerialNumbers, tokenID)
	assert.Equal(t, nil, err)

	listInputCoins := make([]*models.Coins, 0)
	for _, coin := range listCoins {
		if mapSerialNumbers[coin.SerialNumber] == false {
			listInputCoins = append(listInputCoins, coin)
		}
	}
}

func TestRPCService_GetListReceiveTxHash(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)

	listTxHash, _ := rpcCaller.GetListReceiveTxHash(paymentAddStr)
	for _, item := range listTxHash {
		dataByte, _ := rpcCaller.GetTransactionByHash(item)
		autoTxHash := new(models.AutoTxByHash)
		json.Unmarshal(dataByte,autoTxHash)
		if autoTxHash.Result.Type == "tp" {
			fmt.Println(item)
			inputCoins := autoTxHash.Result.PrivacyCustomTokenProofDetail.InputCoins
			outputCoins := autoTxHash.Result.PrivacyCustomTokenProofDetail.OutputCoins
			inAmount := uint64(0)
			outAmount := uint64(0)
			for _, coin := range inputCoins {
				inAmount += coin.CoinDetails.Value
			}
			for _, coin := range outputCoins {
				outAmount += coin.CoinDetails.Value
			}
			fmt.Println(inAmount, outAmount)
		}

	}
}

func TestRPCService_GetAllToken(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	listAutoToken, err := rpcCaller.GetAllToken()

	fmt.Println(len(listAutoToken), err)

}

func TestRPCService_GetBeaconHeight(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	beaconHeight, err := rpcCaller.GetBeaconHeight()
	if err != nil {
		fmt.Println("Error : ", err)
	}
	fmt.Println(beaconHeight)
}

func TestRPCService_GetPdeState(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	pdeState, err := rpcCaller.GetPdePoolParis(100)
	if err != nil {
		fmt.Println("Error : ", err)
	}
	fmt.Println(pdeState)
}
func TestRPCService_GetPdeTradeHistory(t *testing.T)  {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	pdeState, err := rpcCaller.GetPdeHistoryFromBeaconIns(0)
	if err != nil {
		fmt.Println("Error : ", err)
	}
	for _,pde := range pdeState{
		fmt.Println(pde)
	}
}

func TestRPCService_GetBestBlockInfo(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	mapBestBlock, err := rpcCaller.GetBestBlockInfo()
	if err != nil {
		fmt.Println(err)
	}
	for i:= 0; i < 8; i++ {
		fmt.Println(i, mapBestBlock[i].Height, mapBestBlock[i].Hash)
	}
}

func TestRPCService_GetTxsFromRetrieveBlock(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLTestnet)
	mapBestBlock, err := rpcCaller.GetBestBlockInfo()
	if err != nil {
		fmt.Println(err)
	}
	for i:= 0; i < 8; i++ {
		txs, err := rpcCaller.GetTxsFromRetrieveBlock(mapBestBlock[i].Hash)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(mapBestBlock[i].Height, txs)
	}
}

func TestRPCService_GetMinerInfo(t *testing.T)  {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLMainnet)
	pdeState, err := rpcCaller.GetCommitteeInfo()
	if err != nil {
		fmt.Println("Error : ", err)
	}
	for _,pde := range pdeState {
		fmt.Println(pde)
	}
}

func TestRPCService_GetRewardAmount(t *testing.T) {
	rpcCaller := new(RPCService)
	rpcCaller.InitTestnet(common.URLMainnet)
	pdeState, err := rpcCaller.GetRewardAmount()
	if err != nil {
		fmt.Println("Error : ", err)
	}
	for pubKey, pde := range pdeState {
		fmt.Println(pubKey, pde)
	}
}