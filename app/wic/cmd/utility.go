package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"incwallet/app/controllers"
	"incwallet/app/models"
)

func getAllToken() (map[string]*models.Tokens, error) {
	resBytesToken := getAllTokenHelper()
	result, appError := getResponse(resBytesToken)
	if appError != nil {
		return nil, errors.New(fmt.Sprintf("cannot get all token. Error %v", appError))
	} else {
		var mapToken map[string]*models.Tokens
		data, err := json.Marshal(result)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("cannot get all token. Error %v", err))
		}
		if err := json.Unmarshal(data, &mapToken); err != nil {
			return nil, errors.New(fmt.Sprintf("cannot get all token. Error %v", err))
		}
		return mapToken, nil
	}
}

func getTokenBySymbol(symbol string) (*models.Tokens, error) {
	resBytesToken := getTokenBySymbolHelper(symbol, true)
	result, appError := getResponse(resBytesToken)
	if appError != nil {
		return nil, errors.New(fmt.Sprintf("cannot get token %v. Error %v", symbol, appError))
	} else {
		var token *models.Tokens
		data, err := json.Marshal(result)
		if err != nil {
			errors.New(fmt.Sprintf("cannot get token %v. Error %v", symbol, err))
		}
		if err := json.Unmarshal(data, &token); err != nil {
			return nil, errors.New(fmt.Sprintf("cannot get token %v. Error %v", symbol, err))
		}
		return token, nil
	}
}

func getTokenByID(tokenID string) (*models.Tokens, error) {
	resBytesToken := getTokenByIDHelper(tokenID)
	result, appError := getResponse(resBytesToken)
	if appError != nil {
		return nil, errors.New(fmt.Sprintf("cannot get token %v. Error %v", tokenID, appError))
	} else {
		var token *models.Tokens
		data, err := json.Marshal(result)
		if err != nil {
			errors.New(fmt.Sprintf("cannot get token %v. Error %v", tokenID, err))
		}
		if err := json.Unmarshal(data, &token); err != nil {
			return nil, errors.New(fmt.Sprintf("cannot get token %v. Error %v", tokenID, err))
		}
		return token, nil
	}
}

func getResponse(resBytes []byte) (interface{}, *controllers.AppError) {
	responseStatus := new(controllers.ResponseJson)
	json.Unmarshal(resBytes, responseStatus)
	if responseStatus.Error != nil {
		return nil, responseStatus.Error
	} else {
		return responseStatus.Msg, nil
	}
}