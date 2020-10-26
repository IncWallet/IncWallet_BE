package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"incwallet/app/controllers"
	"incwallet/app/lib/common"
	"log"
	"math"
	"os"
	"strings"
	"syscall"
)

// pdeCmd
var pdeCmd = &cobra.Command{
	Use:   "pde",
	Short: "pde command for trade history, check price, and trade",
}

func getState() (*controllers.StateJson, error) {
	url := "/state"
	query := fmt.Sprintf(`{}`)
	resBytes, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	result, appError := getResponse(resBytes)
	if appError != nil {
		fmt.Println("Status: error - ", appError.Msg)
		return nil, errors.New(appError.Msg)
	}
	var state *controllers.StateJson
	data, _ := json.Marshal(result)
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, errors.New(err.Error())
	}
	return state, nil
}

func getPrice(fromToken, toToken string, amount uint64, fee uint64) (*controllers.PdePoolPairPriceJson, error) {
	url := "/pde/price"
	query := fmt.Sprintf(`{
    	"fromtokenidstr" : "%s",
    	"totokenidstr" : "%s",
		"exchangeamount": %v,
		"exchangefee": %v
  		}`, fromToken, toToken, amount,fee)
	state, err := getState()
	if err != nil{
		log.Fatalln("Error send post request query to get state")
	}
	resBytes := make([]byte,0)
	if state.NetworkName == common.Mainnet {
		resBytes, err = SendPostRequestWithQueryToService(query,url)
		if err!= nil{
			log.Fatalln("Error send post request query")
		}
	} else {
		resBytes, err = SendPostRequestWithQuery(query,url)
		if err!= nil{
			log.Fatalln("Error send post request query")
		}
	}

	result, appError := getResponse(resBytes)
	if appError != nil {
		fmt.Println("Status: error - ", appError.Msg)
		return nil, errors.New(appError.Msg)
	}
	var poolPairPrice *controllers.PdePoolPairPriceJson
	data, _ := json.Marshal(result)
	if err := json.Unmarshal(data, &poolPairPrice); err != nil {
		return nil, errors.New(err.Error())
	}
	return poolPairPrice, nil
}

func getPdeTxHistory(pair, symbol string, limit int)  []byte{
	url := "/pde/txhistory"
	token1s := ""
	token2s := ""
	if pair != "" {
		symbols := strings.Split(pair, "-")
		if len(symbols) > 1 {
			token1s = symbols[0]
			token2s = symbols[1]
		} else {
			token1s = symbols[0]
		}
	} else {
		token1s = symbol
	}
	token1id := ""
	token2id := ""
	if token1s != ""{
		if token1, _ := getTokenBySymbol(token1s); token1 != nil {
			token1id = token1.ID
		}
	}
	if token2s != "" {
		if token2, _ := getTokenBySymbol(token2s); token2 != nil {
			token2id = token2.ID
		}
	}
	query := fmt.Sprintf(`{
		"limit": %d,
		"token1id": "%v",
		"token2id": "%v"
		}`, limit, token1id, token2id)
	state, err := getState()
	if err!= nil{
		log.Fatalln("Error send post request query to get state")
	}
	if state.NetworkName == common.Mainnet {
		res, err := SendPostRequestWithQueryToService(query,url)
		if err!= nil{
			log.Fatalln("Error send post request query")
		}
		return res
	} else {
		res, err := SendPostRequestWithQuery(query,url)
		if err!= nil{
			log.Fatalln("Error send post request query")
		}
		return res
	}
}

func sendTradePRVRequest(fromTokenId, toTokenId string,
	sendAmount uint64, minReceiveAmount uint64,
	tradingFee uint64, traderAddressStr, passphrase string, isCheck bool) []byte {

	query := fmt.Sprintf(`{
		"fromtokenid": "%s",
		"totokenid": "%s",
		"sendamount": %v,
		"minreceiveamount": %v,
		"tradingfee": %v,
		"traderaddressstr":"%v",
		"txfee": %v,
		"passphrase": "%s"
		}`, fromTokenId, toTokenId, sendAmount, minReceiveAmount, tradingFee, traderAddressStr, defaultFee, passphrase)
	if isCheck == true {
		fromToken, _ := getTokenByID(fromTokenId)
		toToken, _ := getTokenByID(toTokenId)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Trade Request", "View", "Raw"})
		table.Append([]string{"From Token", fromToken.Symbol, fromToken.ID})
		table.Append([]string{"Send Amount", getViewValueHelper(sendAmount, fromToken.Decimal), fmt.Sprintf("%v",sendAmount)})
		table.Append([]string{"To Token", toToken.Symbol, toToken.ID})
		table.Append([]string{"Min Accepted Amount", getViewValueHelper(minReceiveAmount, toToken.Decimal), fmt.Sprintf("%v", minReceiveAmount) })
		table.Render()
		return nil
	} else {
		url := ""
		if fromTokenId == common.PRVID {
			url = "/pde/tradeprv"
		} else {
			url = "/pde/tradetoken"
		}
		res, err := SendPostRequestWithQuery(query,url)
		if err!= nil{
			log.Fatalln("Error send post request query")
		}
		return res
	}
}

// pdePriceCmd
var pdePriceCmd = &cobra.Command{
	Use:   "price",
	Short: "pde price",
	Run: func(cmd *cobra.Command, args []string) {
		pair, _ := cmd.Flags().GetString("pair")
		amount, _ := cmd.Flags().GetFloat64("amount")
		symbols := strings.Split(pair, "-")
		rawAmount := uint64(0)
		fromTokenID := symbols[0]
		if fromTokenID != "" {
			if tokenInfo, err := getTokenBySymbol(fromTokenID); err == nil {
				fromTokenID = tokenInfo.ID
				rawAmount = uint64(amount * math.Pow10(tokenInfo.Decimal))
			}
		}
		toTokenID := symbols[1]
		if toTokenID != "" {
			if tokenInfo, err := getTokenBySymbol(toTokenID); err == nil {
				toTokenID = tokenInfo.ID
			}
		}

		poolPairPrice, err := getPrice(fromTokenID, toTokenID, rawAmount, defaultTradeFee)
		if err != nil {
			fmt.Println("Status: error - ", err.Error())
		} else {
			fmt.Println("Status: done")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"", poolPairPrice.FromTokenName, poolPairPrice.ToTokenName})
			table.Append([]string{
				"Symbol",
				poolPairPrice.FromTokenSymbol,
				poolPairPrice.ToTokenSymbol,
			})
			table.Append([]string{
				"Pool Size",
				getViewValueHelper(poolPairPrice.FromTokenPoolValue, poolPairPrice.FromTokenDecimal),
				getViewValueHelper(poolPairPrice.ToTokenPoolValue, poolPairPrice.ToTokenDecimal),
			})
			table.Append([]string{
				"Exchange Rate",
				getViewValueHelper(poolPairPrice.ExchangeAmount, poolPairPrice.FromTokenDecimal),
				getViewValueHelper(poolPairPrice.ReceiveAmount, poolPairPrice.ToTokenDecimal),
			})
			table.Render()
		}
	},
}

// pdeTxHistoryCmd
var pdeTxHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "pde tx history command",
	Run: func(cmd *cobra.Command, args []string) {
		pair, _ := cmd.Flags().GetString("pair")
		symbol, _ := cmd.Flags().GetString("symbol")
		limit, _ := cmd.Flags().GetInt("limit")

		resBytes := getPdeTxHistory(pair, symbol, limit)
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			fmt.Println("Status: done")

			var listTx []*controllers.PdeHistoryJson
			data, _ := json.Marshal(result)
			err := json.Unmarshal(data, &listTx)

			if err != nil{
				fmt.Println("Error form Unmarshal : ", err)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Lock Time", "Send", "Send Amount","Receive", "Receive Amount", "Status","Tx Trade"})

			for _,tx := range listTx{
				table.Append([]string{
					fmt.Sprintf("%v",tx.LockTime),
					tx.SendTokenSymbol,
					getViewValueHelper(tx.SendAmount, tx.SendTokenDecimal),
					tx.ReceiveTokenSymbol,
					getViewValueHelper(tx.ReceiverAmount, tx.ReceiveTokenDecimal),
					tx.Status,
					tx.RequestedTxID})
			}
			table.Render()
		}
	},
}

// pdeTradeCmd
var pdeTradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "pde trade command",
	Run: func(cmd *cobra.Command, args []string) {
		fromToken, _ := cmd.Flags().GetString("fromtoken")
		toToken, _ := cmd.Flags().GetString("totoken")
		sendAmount, _ := cmd.Flags().GetFloat64("sendamount")
		minAcceptableLoss, _ := cmd.Flags().GetFloat64("minacceptableloss")

		// get from token id and raw amount to send
		fromTokenId := fromToken
		rawSendAmount := uint64(0)
		if token, _ := getTokenBySymbol(fromToken); token != nil {
			fromTokenId = token.ID
			rawSendAmount = uint64(sendAmount * math.Pow10(token.Decimal))
		}

		if rawSendAmount == uint64(0) {
			fmt.Println("Status: error")
			fmt.Println("Msg: invalid trade request - cannot found from token info")
			return
		}

		// get to token id and min raw receive token amount
		toTokenId := toToken
		rawMinAcceptableReceiveAmount := uint64(0)
		if token, err := getTokenBySymbol(toToken); token != nil{
			toTokenId = token.ID
			poolPairPrice, err := getPrice(fromTokenId, toTokenId, rawSendAmount, defaultFee)
			if err != nil {
				fmt.Println("Status: error - ", err.Error())
			} else {
				rawMinAcceptableReceiveAmount = uint64(float64(poolPairPrice.ReceiveAmount) * (100 - minAcceptableLoss)/100)
			}
		} else {
			fmt.Println("Status: error")
			fmt.Println("Msg: invalid trade request - cannot found to token info.")
			fmt.Println(fmt.Sprintf("Detail: %v", err))
			return
		}

		if rawMinAcceptableReceiveAmount == uint64(0) {
			fmt.Println("Status: error")
			fmt.Println("Msg: invalid trade request")
			fmt.Println(fmt.Sprintf("Detail: receive amount is zero"))
			return
		}

		// get trader address from account info
		traderAddressStr := ""
		if accInfo, _ := getInfo(""); accInfo != nil {
			traderAddressStr = accInfo.PaymentAddress
		}
		if traderAddressStr == "" {
			fmt.Println("Status: error")
			fmt.Println("Msg: invalid trade request - trader address is empty")
			return
		}

		// show trade request to check
		sendTradePRVRequest(fromTokenId, toTokenId, rawSendAmount, rawMinAcceptableReceiveAmount,
			defaultTradeFee, traderAddressStr, "", true)

		// get password
		fmt.Print("Password to confirm:")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}

		// send trade request
		resBytes := sendTradePRVRequest(fromTokenId, toTokenId, rawSendAmount, rawMinAcceptableReceiveAmount,
			defaultTradeFee, traderAddressStr, string(password), false)
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println()
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			fmt.Println("Status: done")
			fmt.Println("Tx Hash:", result)
		}
	},
}