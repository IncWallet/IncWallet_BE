package cmd

import (
	"bytes"
	"fmt"
	"incwallet/app/lib/common"
	"io/ioutil"
	"log"
	"math"
	"net/http"
)

var defaultFee = uint64(5)
var defaultTradeFee = uint64(0)
var minAcceptableLoss = float64(1)

func SendPostRequestWithQueryToService(query, url string, ) ([]byte, error) {
	var jsonStr = []byte(query)
	req, _ := http.NewRequest("POST", common.ServiceURL+url, bytes.NewBuffer(jsonStr))
	req.Close = true
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

func SendPostRequestWithQuery(query, url string, ) ([]byte, error) {
	var jsonStr = []byte(query)
	req, _ := http.NewRequest("POST", common.LocalURL+url, bytes.NewBuffer(jsonStr))
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

func getViewValueHelper(value uint64, decimal int) string {
	base := math.Pow10(decimal)
	view := float64(value) / base
	for i := 4; i < -8; i-- {
		if view > math.Pow10(i) {
			return fmt.Sprintf("%f", math.Round(view/math.Pow10(i))*math.Pow10(i))
		}
	}
	return fmt.Sprintf("%f", view)
}

func getTokenByIDHelper(tokenID string) []byte {
	url := "/network/gettokenbyid"
	query := fmt.Sprintf(`{
		"tokenid": "%s"
	}`, tokenID)

	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func getTokenBySymbolHelper(tokenSymbol string, verified bool) []byte {
	url := "/network/gettokenbysymbol"
	query := fmt.Sprintf(`{
		"tokensymbol": "%s",
		"verified": %v
	}`, tokenSymbol, verified)

	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}

	return res
}

func getAllTokenHelper() []byte {
	url := "/network/getalltokens"
	query := fmt.Sprintf(`{}`)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}

	return res
}
