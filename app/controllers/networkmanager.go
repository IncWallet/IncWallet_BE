package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/common"
	"incwallet/app/lib/rpccaller"
	"incwallet/app/models"
	"io/ioutil"
	"net/http"
	"strings"
)

type  NetworkManager struct {
	NetworkName string
	Network   *models.Network
	BeaconState *models.BeaconState
	BestBlock map[int]uint64
}

func (nm *NetworkManager) Init(name, url string) error {
	network := &models.Network{}

	if name != common.Testnet && name != common.Mainnet && name != common.Local{
		return errors.New("network name is not correct")
	}
	network.Name = name
	if url == "" {
		network.Url = common.GetNetworkURL(name)
	} else {
		network.Url = url
	}
	nm.Network = network
	nm.NetworkName = name
	nm.BeaconState = new(models.BeaconState)
	nm.BestBlock = make(map[int]uint64)
	return nil
}

func (nm *NetworkManager) UpdateBeaconState() error {
	if flag, _ := IsStateFull() ; !flag{
		return nil
	}
	//if StateM.NetworkManager.Network.Name == common.Mainnet{
	//	return nil
	//}

	beaconState, err := StateM.RpcCaller.GetBeaconHeight()
	if err != nil {
		return errors.New(fmt.Sprintf("cannot get beacon beststate. Error %v", err))
	}
	if nm.BeaconState == nil || nm.BeaconState.Height < beaconState.Height {
		nm.BeaconState = beaconState
	}
	return nil
}

func (nm *NetworkManager) GetDecimalByID(tokenID string) (int, error) {
	token, err := nm.GetTokenByID(tokenID)
	if err != nil {
		return 0, err
	}
	return token.Decimal, nil
}

func (nm *NetworkManager) GetAllTokens(pageSize, pageIndex int) (map[string]*models.Tokens, error) {
	prv := &models.Tokens{
			ID:       common.PRVID,
			Symbol:   "PRV",
			Name:     "Privacy Coin",
			Decimal:  9,
			EDecimal: 9,
			ESymbol:  "PRV",
			Verified: true,
			Image: fmt.Sprintf("%v%v%v",common.ImageURLPrefix, "prv", common.ImageURLSubfix),
			Network:  nm.NetworkName}

	var tokens []*models.Tokens
	if err := database.Tokens.Find(bson.M{
		"network" : nm.NetworkName,
	}).Skip((pageIndex-1) * pageSize).Limit(pageSize).All(&tokens); err != nil {
		return nil, errors.New(fmt.Sprintf("cannot get all tokens. Error %v", err))
	}
	mapToken := make(map[string]*models.Tokens)
	mapToken[common.PRVID] = prv
	for _, token := range tokens {
		mapToken[token.ID] = token
	}
	return mapToken, nil
}

func (nm *NetworkManager) GetTokenByID(tokenID string) (*models.Tokens, error) {
	if tokenID == common.PRVID {
		return &models.Tokens{
			ID:       common.PRVID,
			Symbol:   "PRV",
			Name:     "Privacy Coin",
			Decimal:  9,
			EDecimal: 9,
			ESymbol:  "PRV",
			Verified: true,
			Network:  nm.NetworkName,
			Image: fmt.Sprintf("%v%v%v",common.ImageURLPrefix, "prv", common.ImageURLSubfix),
		}, nil
	}
	var token models.Tokens
	if err := database.Tokens.Find(bson.M{
		"id" : tokenID,
		"network" : nm.NetworkName,
	}).One(&token); err != nil {
		return nil, errors.New(fmt.Sprintf("Token ID %v is not exist", tokenID))
	}
	return &token, nil
}

func (nm *NetworkManager) GetTokenBySymbol(symbol string) (*models.Tokens, error) {
	if strings.EqualFold(symbol, common.PRVSymbol) {
		return &models.Tokens{
			ID:       common.PRVID,
			Symbol:   "PRV",
			Name:     "Privacy Coin",
			Decimal:  9,
			EDecimal: 9,
			ESymbol:  "PRV",
			Verified: true,
			Network:  nm.NetworkName,
			Image: fmt.Sprintf("%v%v%v",common.ImageURLPrefix, "prv", common.ImageURLSubfix),
		}, nil
	}
	var tokens []models.Tokens
	if err := database.Tokens.Find(bson.M{
		"symbol" :bson.RegEx{
			Pattern: "^" + symbol,
			Options: "i",
		},
		"network" : nm.NetworkName,
	}).All(&tokens); err != nil {
		return nil, errors.New(fmt.Sprintf("Token symbol %v is not exist", symbol))
	}
	if len(tokens) > 0 {
		for i := range tokens {
			if tokens[i].Verified == true {
				return &tokens[i], nil
			}
		}
		return &tokens[0], nil
	}
	return nil, errors.New("requested token does not exist")
}

func (nm *NetworkManager) UpdateAllToken() (int, error){
	var listToken []models.Tokens
	mapToken := make(map[string]bool)
	if err := database.Tokens.Find(bson.M{}).All(&listToken); err == nil {
		revel.AppLog.Infof("Current number of tokens local: %v", len(listToken))
		for _, token := range listToken {
			mapToken[token.ID] = true
		}
	}

	listAutoAppTokenInfo, err := nm.GetListAppTokenInfo()
	if err != nil {
		revel.AppLog.Errorf("cannot get list token info from app. Error %v", err)
	}

	listAutoToken, err := StateM.RpcCaller.GetAllToken()
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Cannot request list tokens. Error %v", err))
	}
	revel.AppLog.Infof("Current number of tokens remote: %v", len(listAutoToken))
	newTokens := make([]*models.Tokens, 0)
	for _, autoToken := range listAutoToken {
		if _, found := mapToken[autoToken.ID]; !found {
			//Check token info from app
			if listAutoAppTokenInfo != nil {
				for _, appToken := range listAutoAppTokenInfo.Result {
					if appToken.TokenID == autoToken.ID {
						autoToken.Decimal = appToken.PDecimals
						autoToken.EDecimal = appToken.Decimals
						autoToken.Symbol = appToken.PSymbol
						autoToken.ESymbol = appToken.Symbol
						autoToken.Name = appToken.Name
						autoToken.Verified = appToken.Verified
						autoToken.Network = nm.NetworkName
						autoToken.Image = fmt.Sprintf("%v%v%v",common.ImageURLPrefix, strings.ToLower(appToken.Symbol[0:]), common.ImageURLSubfix)
					}
				}
			}
			//End check token info from app
			newTokens = append(newTokens, autoToken)
			if err := database.Tokens.Insert(autoToken); err != nil {
				return 0, errors.New(fmt.Sprintf("Cannot insert new token. Error %v", err))
			}
		}
	}

	revel.AppLog.Infof("Update %v new tokens from remote", len(newTokens))
	return len(newTokens), nil
}

func (nm *NetworkManager) GetListAppTokenInfo() (*rpccaller.AutoListAppTokenInfo, error){
	var url string
	if nm.NetworkName == common.Testnet {
		url = "https://test-api2.incognito.org/ptoken/list"
	}
	if nm.NetworkName == common.Mainnet {
		url = "https://api2.incognito.org/ptoken/list"
	}

	resp, err := http.Get(url)
	if err != nil {
		revel.AppLog.Errorf("cannot get list token info from app api. Error %v", err)
		return nil, errors.New(fmt.Sprintf("cannot get list token info from app api. Error %v", err))
	}
	defer resp.Body.Close()

	resultBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		revel.AppLog.Errorf("cannot read body from response. Error %v", err)
		return nil, errors.New(fmt.Sprintf("cannot read body from response. Error %v", err))
	}

	var listTokenInfo rpccaller.AutoListAppTokenInfo
	err = json.Unmarshal(resultBytes, &listTokenInfo)
	if err != nil {
		revel.AppLog.Errorf("cannot unmarshal result. Error %v", err)
		return nil, errors.New(fmt.Sprintf("cannot unmarshal result. Error %v", err))
	}
	return &listTokenInfo, nil
}