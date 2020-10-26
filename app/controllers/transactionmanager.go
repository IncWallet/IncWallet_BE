package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"incwallet/app/lib/hdwallet"
	"incwallet/app/lib/rpccaller"
	"incwallet/app/lib/transaction"
	"incwallet/app/models"
	"strconv"
)

type TransactionManager struct {
	am *AccountManager
	rpcCaller *rpccaller.RPCService
}

func ParseString2Point (data string) (*privacy.Point, error){
	pByte, _, err := base58.Base58Check{}.Decode(data)
	if err != nil {
		return nil, err
	}
	p, err := new(privacy.Point).FromBytesS(pByte)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ParseString2Scalar (data string) (*privacy.Scalar, error){
	sByte, _, err := base58.Base58Check{}.Decode(data)
	if err != nil {
		return nil, err
	}
	s := new(privacy.Scalar).FromBytesS(sByte)
	return s, nil
}

func (txManager *TransactionManager) buildTxParam(keyWallet *hdwallet.KeyWallet,
	allUnspentCoins []*models.Coins,
	receivers map[string]uint64,
	fee uint64,
	info string,
	tokenID string,
	hasPrivacy bool,
	meta transaction.Metadata) (*transaction.TxPrivacyInitParams, *transaction.Commitment, error) {

	paymentInfos := make([]*privacy.PaymentInfo, 0)
	totalTransferAmount := fee
	for paymentAddressStr, amount := range receivers {
		paymentAddressStr, err := strconv.Unquote(paymentAddressStr)
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("cannot set receiver key in BuildTxParam. Quote .... Err %v", err))
		}
		receiverKeyWallet, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("cannot set receiver key in BuildTxParam. Err %v", err))
		}
		paymentInfos = append(paymentInfos, &privacy.PaymentInfo{
			PaymentAddress: receiverKeyWallet.KeySet.PaymentAddress,
			Amount:         amount,
		})
		totalTransferAmount += amount
	}

	inputCoins, _, _, err  := txManager.am.ChooseBestCoinsToSpend(allUnspentCoins, totalTransferAmount)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("cannot choose best coins to spend in BuildTxParam. Error %v", err))
	}
	senderPaymentAddStr := keyWallet.Base58CheckSerialize(wallet.PaymentAddressType)

	inputCoinsBytes, err := json.Marshal(inputCoins)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("cannot marshal input coin in BuildTxParam. Error %v", err))
	}
	cmIndexes, myCmIndexes, commitments, err := txManager.rpcCaller.GetRandomCommitments(senderPaymentAddStr, string(inputCoinsBytes), tokenID)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("cannot get random commitments in BuildTxParam. Error %v", err))
	}

	inputCommitment := &transaction.Commitment{
		CmIndexes:   cmIndexes,
		MyCmIndexes: myCmIndexes,
		Commitments: commitments,
	}

	incInputCoins := make([]*privacy.InputCoin, len(inputCoins))
	for i := range inputCoins {
		incInputCoins[i] = new(privacy.InputCoin).Init()
		pubKey, err := ParseString2Point(inputCoins[i].PublicKey)
		if err != nil {
			return nil, nil, err
		}
		incInputCoins[i].CoinDetails.SetPublicKey(pubKey)

		value, err := strconv.ParseUint(inputCoins[i].Value,10, 64)
		if err != nil {
			return nil, nil, errors.New(fmt.Sprintf("cannot parse input coin value. Error %v", err))
		}
		incInputCoins[i].CoinDetails.SetValue(value)

		cm, err := ParseString2Point(inputCoins[i].CoinCommitment)
		if err != nil {
			return nil, nil, err
		}
		incInputCoins[i].CoinDetails.SetCoinCommitment(cm)

		snd, err := ParseString2Scalar(inputCoins[i].SNDerivator)
		if err != nil {
			return nil, nil, err
		}
		incInputCoins[i].CoinDetails.SetSNDerivator(snd)

		sn, err := ParseString2Point(inputCoins[i].SerialNumber)
		if err != nil {
			return nil, nil, err
		}
		incInputCoins[i].CoinDetails.SetSerialNumber(sn)

		r, err := ParseString2Scalar(inputCoins[i].Randomness)
		if err != nil {
			return nil, nil, err
		}
		incInputCoins[i].CoinDetails.SetRandomness(r)

		incInputCoins[i].CoinDetails.SetInfo([]byte(inputCoins[i].Info))
		b, _ := json.Marshal(incInputCoins[i])
		fmt.Println(string(b))
	}
	privateKey := privacy.PrivateKey(keyWallet.KeySet.PrivateKey)
	param := transaction.InitTxPrivacyParams(&privateKey,paymentInfos, incInputCoins, fee, hasPrivacy, tokenID, meta,[]byte(info) )
	return param, inputCommitment, nil
}

func (txManager *TransactionManager) InitTxParam(keyWallet *hdwallet.KeyWallet,
	receivers map[string]uint64,
	info string,
	meta transaction.Metadata,
	fee uint64,
	tokenIDStr string,
	hasPrivacy bool) (*transaction.TxPrivacyInitParams, *transaction.Commitment, error) {
	unspentCoins, err := txManager.am.GetUnspentOutputCoinR(keyWallet, tokenIDStr)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("cannot get unspent PRV coin. Error: %v", err))
	}

	return txManager.buildTxParam(keyWallet, unspentCoins, receivers, fee, info, tokenIDStr, hasPrivacy, meta)
}

func (txManager *TransactionManager) InitTransaction(receivers map[string]uint64,
	fee uint64,
	info string,
	metadata transaction.Metadata,
	tokenIDStr string,
	hasPrivacy bool,
	passphrase string) (string, error) {
	privateKeyBytes, err := StateM.WalletManager.SafeStore.DecryptPrivateKey(&StateM.AccountManage.Account.Crypto, passphrase)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot decrypt privatekey. Error %v", err))
	}

	keyWallet, err := hdwallet.Base58CheckDeserialize(string(privateKeyBytes))
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot init key wallet. Error %v", err))
	}
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)

	txParam, commitments, err := txManager.InitTxParam(keyWallet, receivers, info, metadata, fee, tokenIDStr, hasPrivacy)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot build tx param. Error %v", err))
	}

	tx := new(transaction.Tx)
	err = tx.Init(txParam, commitments)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot init transaction in InitTransaction. Error %v", err))
	}
	txByte, err := json.Marshal(tx)

	tmp := tx.Proof.GetInputCoins()
	fmt.Println(len(tmp))
	if len(tmp) > 0 {
		b, _ := json.Marshal(tmp)
		fmt.Println(string(b))
		fmt.Println(tmp[0].CoinDetails.GetCoinCommitment())
	}

	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot marshal tx. Error %v", err))
	}
	base58CheckData := base58.Base58Check{}.Encode(txByte, common.ZeroByte)

	//send raw transaction
	_, err = StateM.RpcCaller.SendRawTx(base58CheckData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot send raw transaction. Error %v", err))
	}
	return tx.Hash().String(), nil
}


func (txManager *TransactionManager) InitTokenTransaction(receiversPrv, receiversToken map[string]uint64,
	fee uint64,
	info string,
	metadata transaction.Metadata,
	tokenIDStr string,
	hasPrivacy bool,
	passphrase string) (string, error) {
	privateKeyBytes, err := StateM.WalletManager.SafeStore.DecryptPrivateKey(&StateM.AccountManage.Account.Crypto, passphrase)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot decrypt privatekey. Error %v", err))
	}

	keyWallet, err := hdwallet.Base58CheckDeserialize(string(privateKeyBytes))
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot init key wallet. Error %v", err))
	}
	keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)

	prvTxParam, prvCommitments, err := txManager.InitTxParam(keyWallet, receiversPrv, info, metadata, fee, common.PRVID, hasPrivacy)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot build tx fee param. Error %v", err))
	}

	tokenTxParam, tokenCommitments, err := txManager.InitTxParam(keyWallet, receiversToken, "", nil, uint64(0), tokenIDStr, hasPrivacy)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot build tx token param. Error %v", err))
	}

	mintable := false
	//if metadata != nil {
	//	mintable = true
	//}
	txToken := new(transaction.TxCustomTokenPrivacy)
	err = txToken.Init(prvTxParam, tokenTxParam, prvCommitments, tokenCommitments, mintable)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot init token transaction. Error %v", err))
	}
	txByte, err := json.Marshal(txToken)
	if err != nil {
		return "", errors.New(fmt.Sprintf("cannot marshal txtoken. Error %v", err))
	}
	base58CheckData := base58.Base58Check{}.Encode(txByte, common.ZeroByte)

	//send raw transaction
	_, err = StateM.RpcCaller.SendRawTxToken(base58CheckData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Cannot send raw transaction. Error %v", err))
	}
	return txToken.Hash().String(), nil
}

