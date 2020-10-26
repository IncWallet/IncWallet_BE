package controllers

import (
	"errors"
	"github.com/revel/revel"
	"incwallet/app/database"
	"net/http"
	"strconv"
)

/*
Network controller
*/
type NetworkCtrl struct {
	*revel.Controller
}

type NetworkParam struct {
	TokenName   string `json:"tokenname"`
	TokenSymbol string `json:"tokensymbol"`
	Verified    bool   `json:"verified"`
	TokenID     string `json:"tokenid"`
	NetworkName string `json:"networkname"`
	NetworkURL  string `json:"networkurl"`
}

/*
/network/gettokenbyID
Get Token by ID
- token id
*/
func (c NetworkCtrl) GetTokenByID() revel.Result {
	networkParam := &NetworkParam{}
	if err := c.Params.BindJSON(&networkParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	} else {
		token, err := StateM.NetworkManager.GetTokenByID(networkParam.TokenID)
		if err != nil {
			revel.AppLog.Errorf("Cannot get token by id from database. Error %v", err)
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot get token by id"), err.Error(), 0))
		}
		c.Response.Status = http.StatusCreated
		return c.RenderJSON(responseJsonBuilder(nil, token, 0))
	}
}

/*
/network/gettokenbysymbol
Get Token by Symbol
- token symbol
*/
func (c NetworkCtrl) GetTokenBySymbol() revel.Result {
	networkParam := &NetworkParam{}
	if err := c.Params.BindJSON(&networkParam); err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	} else {
		token, err := StateM.NetworkManager.GetTokenBySymbol(networkParam.TokenSymbol)
		if err != nil {
			revel.AppLog.Errorf("Cannot get token by id from database. Error %v", err)
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJSON(responseJsonBuilder(errors.New("cannot get token by symbol"), err.Error(), 0))
		}
		c.Response.Status = http.StatusCreated
		return c.RenderJSON(responseJsonBuilder(nil, token, 0))
	}
}

/*
/network/getalltokens
Get All Token
*/
func (c NetworkCtrl) GetAllToken() revel.Result {
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

	size, err := database.Tokens.Find(nil).Count()
	if err != nil {
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get token size"), err.Error(), 0))
	}

	mapTokens, err := StateM.NetworkManager.GetAllTokens(pageSize, pageIndex)
	if err != nil {
		revel.AppLog.Errorf("Cannot get token by id from database. Error %v", err)
		return c.RenderJSON(responseJsonBuilder(errors.New("cannot get token by id"), err.Error(), 0))
	}
	result := &Result{
		Size:   size,
		Detail: mapTokens,
	}
	return c.RenderJSON(responseJsonBuilder(nil, result, 0))
}
