package controllers

import (
	"errors"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/base58"
	"incwallet/app/lib/common"
	"incwallet/app/lib/hdwallet"
	"incwallet/app/models"
)

type MinerCtrl struct {
	*revel.Controller
}

type MinerParams struct {
	PaymentAddresses []string `json:"paymentaddresses"`
	MiningKeys []string `json:"miningkeys"`
}

func (c *MinerCtrl) GetInfoByAccount() revel.Result {
	minerParam := &MinerParams{}
	if err := c.Params.BindJSON(&minerParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}
	listMinerInfoJson := make([]*MinerInfoJson, 0)
	for i := range minerParam.PaymentAddresses {
		minerInfo := &MinerInfoJson{
			PaymentAddress:    minerParam.PaymentAddresses[i],
			MiningKey:    minerParam.MiningKeys[i],
			BeaconHeight: StateM.NetworkManager.BeaconState.Height,
			Epoch:        StateM.NetworkManager.BeaconState.Epoch,
			Reward: uint64(0),
			Status: "None",
			ShardID: -2,
			Index:  0,
		}
		kw, err := hdwallet.Base58CheckDeserialize(minerInfo.PaymentAddress)
		if err != nil {
			return c.RenderJSON(responseJsonBuilder(errors.New("payment address is invalid"), err.Error(), 0))
		}
		publicKeyStr := base58.Base58Check{}.Encode(kw.KeySet.PaymentAddress.Pk, common.ZeroByte)
		var reward models.CommitteReward
		if err := database.Reward.Find(bson.M{"publickey":publicKeyStr}).One(&reward); err == nil {
			minerInfo.Reward = reward.Amount
		}
		committeePublicKey, err := hdwallet.GetMiningPubKey(minerInfo.MiningKey, minerInfo.PaymentAddress)
		if err != nil {
			revel.AppLog.Errorf("cannot get committee public key from mining key %v. Error %v", minerInfo.MiningKey, err)
			listMinerInfoJson = append(listMinerInfoJson, minerInfo)
			continue
		}
		var committee models.Committee
		if err := database.Committee.Find(bson.M{"key": committeePublicKey}).One(&committee); err == nil {
			minerInfo.Status = committee.Role
			minerInfo.ShardID = committee.ShardId
			minerInfo.Index = committee.Index
		} else {
			revel.AppLog.Errorf("cannot get committee info %v. Error %v", committeePublicKey, err)
		}
		listMinerInfoJson = append(listMinerInfoJson, minerInfo)
	}

	return c.RenderJSON(responseJsonBuilder(nil, listMinerInfoJson, 0))
}

func (c *MinerCtrl) GetInfoByAllAccounts() revel.Result {
	minerParam := &MinerParams{}
	if err := c.Params.BindJSON(&minerParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if flag, _ := IsStateFull() ; !flag{
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot start to sync account, import or add account first"), "", 0))
	}

	var listAccounts []models.Account
	if err := database.Accounts.Find(bson.M{
		"wallet": StateM.WalletManager.WalletID,
	}).All(&listAccounts); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get all accounts"), err.Error(), 0))
	}

	listMinerInfoJson := make([]*MinerInfoJson, 0)
	for i := range listAccounts{
		minerInfo := &MinerInfoJson{
			AccountName: listAccounts[i].Name,
			PaymentAddress:    listAccounts[i].PaymentAddress,
			MiningKey:    listAccounts[i].MiningKey,
			BeaconHeight: StateM.NetworkManager.BeaconState.Height,
			Epoch:        StateM.NetworkManager.BeaconState.Epoch,
			Reward: uint64(0),
			Status: "None",
			ShardID: -2,
			Index:  0,
		}
		kw, err := hdwallet.Base58CheckDeserialize(minerInfo.PaymentAddress)
		if err != nil {
			return c.RenderJSON(responseJsonBuilder(errors.New("payment address is invalid"), err.Error(), 0))
		}
		publicKeyStr := base58.Base58Check{}.Encode(kw.KeySet.PaymentAddress.Pk, common.ZeroByte)
		var reward models.CommitteReward
		if err := database.Reward.Find(bson.M{"publickey":publicKeyStr}).One(&reward); err == nil {
			minerInfo.Reward = reward.Amount
		}
		committeePublicKey, err := hdwallet.GetMiningPubKey(minerInfo.MiningKey, minerInfo.PaymentAddress)
		if err != nil {
			revel.AppLog.Errorf("cannot get committee public key from mining key %v. Error %v", minerInfo.MiningKey, err)
			listMinerInfoJson = append(listMinerInfoJson, minerInfo)
			continue
		}
		var committee models.Committee
		if err := database.Committee.Find(bson.M{"key": committeePublicKey}).One(&committee); err == nil {
			minerInfo.Status = committee.Role
			minerInfo.ShardID = committee.ShardId
			minerInfo.Index = committee.Index
		} else {
			revel.AppLog.Errorf("cannot get committee info %v. Error %v", committeePublicKey, err)
		}
		if minerInfo.Reward > 0 && minerInfo.Status != "None" {
			listMinerInfoJson = append(listMinerInfoJson, minerInfo)
		}
	}

	return c.RenderJSON(responseJsonBuilder(nil, listMinerInfoJson, 0))
}

//func (c *MinerCtrl) GetAll() revel.Result {
//
//}