package controllers

import (
	"errors"
	"fmt"
	"incwallet/app/database"
	"incwallet/app/lib/common"
	"incwallet/app/models"
)

type CommitteeManager struct {
}

func (cm *CommitteeManager) UpdateCommitteeFromChain() error{
	if flag, _ := IsStateFull() ; !flag{
		return nil
	}

	//if StateM.NetworkManager.Network.Name == common.Mainnet {
	//	return nil
	//}

	listMiner, err := StateM.RpcCaller.GetCommitteeInfo()
	if err != nil{
		return errors.New(fmt.Sprintf("cannot get committee from chain. Error %v", err))
	}

	listCommittee := make([]interface{},0)
	for _, committee := range listMiner{
		c := &models.Committee{
			Epoch:   committee.Epoch,
			BeaconHeight: StateM.NetworkManager.BeaconState.Height,
			Key:     committee.Key,
			ShardId: committee.ShardId,
			Role:    committee.Role,
			Index:   committee.Index,
		}
		listCommittee = append(listCommittee,c)
	}
	if len(listCommittee) > 0 {
		database.Committee.DropCollection()
	}

	cc := database.Committee.Bulk()
	cc.Insert(listCommittee...)
	_,err = cc.Run()
	if err!= nil{
		return errors.New(fmt.Sprintf("cannot insert bulk of commitee from chain. Error %v", err))
	}
	return nil
}

func (cm *CommitteeManager) UpdateMinerRewardFromChain() error{
	if flag, _ := IsStateFull() ; !flag{
		return nil
	}

	//if StateM.NetworkManager.Network.Name == common.Mainnet {
	//	return nil
	//}

	rpcReward,err := StateM.RpcCaller.GetRewardAmount()
	if err != nil{
		return errors.New(fmt.Sprintf("cannot get reward from chain. Error %v", err))
	}

	listReward := make([]interface{},0)
	for key,_ := range rpcReward{
		r := &models.CommitteReward{
			BeaconHeight: StateM.NetworkManager.BeaconState.Height,
			PublicKey: key,
			Amount: rpcReward[key][common.PRVID],
		}
		listReward = append(listReward,r)
	}

	if len(listReward) > 0 {
		database.Reward.DropCollection()
	}

	cc := database.Reward.Bulk()
	cc.Insert(listReward...)
	_,err = cc.Run()
	if err!= nil{
		return errors.New(fmt.Sprintf("cannot insert bulk of reward from chain. Error %v", err))
	}
	return nil
}