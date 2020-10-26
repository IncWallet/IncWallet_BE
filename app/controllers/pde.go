package controllers

import (
	"errors"
	"github.com/revel/revel"
	"gopkg.in/mgo.v2/bson"
	"incwallet/app/database"
	"incwallet/app/lib/common"
	"incwallet/app/models"
	"math"
	"strconv"
)

type PdeCtrl struct {
	*revel.Controller
}

type PdeParams struct {
	Token1IDStr    string `json:"token1id"`
	Token2IDStr    string `json:"token2id"`
	ExchangeAmount uint64 `json:"exchangeamount"`
	ExchangeFee    uint64 `json:"exchangefee"`
	FromTokenIDStr string `json:"fromtokenidstr"`
	ToTokenIDStr   string `json:"totokenidstr"`
	Limit          int    `json:"limit"`
}

func getExchangeRate(fromToken, toToken  string, amount, fee uint64, isEstimate bool) (uint64, *models.PdePoolPairs, error) {
	pool, err := new(PdeManager).GetPool(fromToken, toToken)
	if err != nil {
		return uint64(0), nil, err
	}
	if isEstimate == true {
		token := getTokenByID(fromToken)
		unitAmount := uint64(math.Pow10(token.Decimal))
		unitReceiveAmount, err := StateM.PdeManager.GetRate(unitAmount, fee, fromToken, toToken, pool)
		if err != nil {
			return 0, nil, err
		}
		receiveAmount := uint64(float64(amount) * float64(unitReceiveAmount) / float64(unitAmount))

		return receiveAmount, pool, err
	}
	receiveAmount, err := StateM.PdeManager.GetRate(amount, fee, fromToken, toToken, pool)
	return receiveAmount, pool, err
}

func (p *PdeCtrl) GetPdePoolPairPrice() revel.Result {
	pdeParams := &PdeParams{}
	if err := p.Params.BindJSON(&pdeParams); err != nil {
		return p.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}

	receiveAmount,requestPoolPair, err := getExchangeRate(pdeParams.FromTokenIDStr, pdeParams.ToTokenIDStr, pdeParams.ExchangeAmount, pdeParams.ExchangeFee, false)
	if err != nil {
		return p.RenderJSON(responseJsonBuilder(errors.New("cannot get exchange rate from pool pair data"), err.Error(), 0))
	}
	return p.RenderJSON(responseJsonBuilder(nil, pdePoolPairPriceJsonBuilder(requestPoolPair,
		pdeParams.FromTokenIDStr, pdeParams.ExchangeAmount, receiveAmount), 0))
}

func (p *PdeCtrl) GetPdeCrossPoolPairPrice() revel.Result {
	pdeParams := &PdeParams{}
	if err := p.Params.BindJSON(&pdeParams); err != nil {
		return p.RenderJSON(responseJsonBuilder(errors.New("bad request"), err.Error(), 0))
	}
	if pdeParams.FromTokenIDStr == common.PRVID || pdeParams.ToTokenIDStr == common.PRVID {
		return p.RenderJSON(responseJsonBuilder(errors.New("bad request"),"one of from or to token id is prv", 0))
	}
	receiveAmount, requestPoolPair, err := getExchangeRate(pdeParams.FromTokenIDStr, common.PRVID, pdeParams.ExchangeAmount, uint64(0), false)
	if err != nil {
		return p.RenderJSON(responseJsonBuilder(errors.New("cannot get exchange rate from first pool pair data"), err.Error(), 0))
	}

	finalReceiveAmount, finalRequestPoolPair, err := getExchangeRate(common.PRVID, pdeParams.ToTokenIDStr, receiveAmount, uint64(0), false)
	if err != nil {
		return p.RenderJSON(responseJsonBuilder(errors.New("cannot get exchange rate from second pool pair data"), err.Error(), 0))
	}
	pdePoolPairPrices := make([]*PdePoolPairPriceJson, 2)
	pdePoolPairPrices[0] = pdePoolPairPriceJsonBuilder(requestPoolPair, pdeParams.FromTokenIDStr, pdeParams.ExchangeAmount, receiveAmount)
	pdePoolPairPrices[1] = pdePoolPairPriceJsonBuilder(finalRequestPoolPair, common.PRVID, receiveAmount, finalReceiveAmount)
	return p.RenderJSON(responseJsonBuilder(nil, pdePoolPairPrices, 0))
}

func (p *PdeCtrl) GetPdeTradeHistory() revel.Result {
	var pageIndex, pageSize int
	var err error
	if p.Params.Get("pageindex") == "" && p.Params.Get("pagesize") == "" {
		pageIndex = 1
		pageSize = 1000000000
	} else {
		pageIndex, err = strconv.Atoi(p.Params.Get("pageindex"))
		if err != nil {
			return p.RenderJSON(responseJsonBuilder(errors.New("cannot get trade history, pageindex is invalid"), err.Error(), 0))
		}
		pageSize, err = strconv.Atoi(p.Params.Get("pagesize"))
		if err != nil {
			return p.RenderJSON(responseJsonBuilder(errors.New("cannot get trade history, pageSize is invalid"), err.Error(), 0))
		}
	}

	tokenID1 := p.Params.Get("tokenid1")
	tokenID2 := p.Params.Get("tokenid2")

	var pdeHistory []*models.PdeTradeHistory
	query := bson.M{}

	if tokenID1 != "" && tokenID2 != "" {
		query1 := bson.M{
			"fromtokenidstr": tokenID1,
			"totokenidstr": tokenID2,
		}
		query2 := bson.M{
			"fromtokenidstr": tokenID2,
			"totokenidstr": tokenID1,
		}
		query = bson.M{"$or": []bson.M{query1, query2}}
	} else {
		if tokenID1 != "" {
			query = bson.M{"$or": []bson.M{bson.M{"fromtokenidstr": tokenID1}, bson.M{"totokenidstr": tokenID1}}}
		}
		if tokenID2 != "" {
			query = bson.M{"$or": []bson.M{bson.M{"fromtokenidstr": tokenID2}, bson.M{"totokenidstr": tokenID2}}}
		}
	}

	size, err := database.PdeHistory.Find(query).Count()
	if err != nil {
		return p.RenderJSON(responseJsonBuilder(errors.New("cannot get total pde history"), err.Error(), 0))
	}

	err = database.PdeHistory.Find(query).Sort("-locktime").Skip((pageIndex-1) * pageSize).Limit(pageSize).All(&pdeHistory)
	if err != nil {
		return p.RenderJSON(responseJsonBuilder(errors.New("cannot get pde history"), err.Error(), 0))
	}

	return p.RenderJSON(responseJsonBuilder(nil, pdeTradeHistoryJsonBuilder(pdeHistory, size), 0))
}
