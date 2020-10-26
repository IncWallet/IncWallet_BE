package controllers

import (
	"errors"
	"fmt"
	"github.com/revel/cron"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/common"
	"incwallet/app/lib/rpccaller"
	"incwallet/app/models"
	"net/http"
)

var StateM *StateManager

type StateManager struct {
	NetworkManager  *NetworkManager
	RpcCaller       *rpccaller.RPCService
	AccountManage   *AccountManager
	WalletManager   *WalletManager
	PdeManager      *PdeManager
	CommitteManager *CommitteeManager
	JobManager      *cron.Cron
}

func IsStateFull() (bool, int) {
	if StateM.WalletManager.WalletID == "" {
		return false, 0
	}
	if StateM.AccountManage.AccountID == "" {
		return false, 1
	}
	if StateM.NetworkManager.NetworkName == "" {
		return false, 2
	}
	return true, 0
}

func InitState() {
	StateM = new(StateManager)
	StateM.AccountManage = new(AccountManager)
	StateM.WalletManager = new(WalletManager)
	StateM.RpcCaller = new(rpccaller.RPCService)
	StateM.NetworkManager = new(NetworkManager)
	StateM.PdeManager = new(PdeManager)
	StateM.CommitteManager = new(CommitteeManager)
}

func LoadState() error {
	InitState()
	state := &models.State{}
	if err := database.State.Find(bson.M{}).One(&state); err != nil {
		if err := database.State.Insert(StateM); err != nil {
			return errors.New(fmt.Sprintf("Cannot create empty state. Error %v", err))
		}
		revel.AppLog.Warnf("State is empty. Create new State")
		return nil
	} else {
		if err := StateM.WalletManager.Init(state.WalletID); err != nil {
			revel.AppLog.Warnf("Cannot load state from Init WalletManage. Error %v", err)
		}
		if err := StateM.AccountManage.Init(state.AccountID); err != nil {
			revel.AppLog.Warnf("Cannot load state from Init WalletManage. Error %v", err)
		}
		if err := StateM.NetworkManager.Init(state.Network, common.GetNetworkURL(state.Network)); err != nil {
			revel.AppLog.Warnf("Cannot load network from Init WalletManage. Error %v", err)
		}
		StateM.RpcCaller.Init(state.Network)
		return nil
	}
}

func (sm *StateManager) SaveState() error {
	currentState := &models.State{}
	if err := database.State.Find(bson.M{}).One(&currentState); err != nil {
		return errors.New(fmt.Sprintf("Cannot load state from database. Error %v", err))
	}
	currentState.WalletID = sm.WalletManager.WalletID
	currentState.AccountID = sm.AccountManage.AccountID
	currentState.Network = sm.NetworkManager.NetworkName

	if err := database.State.UpdateId(currentState.ID, currentState); err != nil {
		return errors.New(fmt.Sprintf("Cannot update State to database. Error %v", err))
	}
	return nil
}

/*
State controller
*/
type StateCtrl struct {
	*revel.Controller
}

/*
State info
*/
func (c StateCtrl) GetInfo() revel.Result {
	flag, code := IsStateFull()
	if !flag {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot show info, import or add account first"), "", code))
	}
	c.Response.Status = http.StatusCreated
	return c.RenderJSON(responseJsonBuilder(nil, stateJsonBuilder(), 0))
}
