package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"incwallet/app/controllers"
	"incwallet/app/lib/common"
	"log"
	"math"
	"os"
	"strconv"
	"syscall"
)

func getTxHistory(token string, limit int) []byte {
	url := "/transactions/history"
	query := fmt.Sprintf(`{
    	"tokenid" : "%s",
		"limit" : %v
  		}`, token, limit)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func getTxInfo(txHash string) []byte {
	url := "/transactions/info"
	query := fmt.Sprintf(`{
		"txhash": "%s"
	}`, txHash)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func buildSendTxQuery(receiver string, amount uint64, info string, fee uint64, tokenID string) string {
	mapReceiver := make(map[string]uint64)
	mapReceiver[receiver] = amount
	receiverByte, _ := json.Marshal(mapReceiver)
	query := fmt.Sprintf(`
	{
		"receivers": %s,
		"fee": %v,
		"info": "%s",
		"tokenid": "%s"
	}`, string(receiverByte), fee, info, tokenID)
	return query
}

func sendTx(receiver string, amount uint64, info string, fee uint64, tokenID string, passphrase string) []byte {
	url := "/transactions/create"
	if tokenID != "" && tokenID != common.PRVID {
		url = "/transactions/createtoken"
	}
	mapReceiver := make(map[string]uint64)
	mapReceiver[receiver] = amount
	receiverByte, _ := json.Marshal(mapReceiver)
	query := fmt.Sprintf(`{
		"receivers": %s,
		"fee": %v,
		"info": "%s",
		"tokenid": "%s",
		"passphrase": "%s"
	}`, string(receiverByte), fee, info, tokenID, passphrase)

	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

// txCmd represent the transaction command
var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "tx commands for history, send, and receive",
}

// Tx history cmd
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show send/receive transactions history",
	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		limit, _ := cmd.Flags().GetInt("limit")

		tokenID := token
		if tokenID != "" {
			if tokenInfo, err := getTokenBySymbol(token); err == nil {
				tokenID = tokenInfo.ID
			}
		}

		resBytes := getTxHistory(tokenID, limit)
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			var listTx []controllers.TxHistoryJson
			data, _ := json.Marshal(result)
			json.Unmarshal(data, &listTx)

			mapToken, err := getAllToken()
			if err != nil {
				fmt.Println("Status: error - ", err)
			} else {
				fmt.Println("Status: done")
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"LockTime", "Type", "Token", "Amount", "TxHash"})
				for _, tx := range listTx {
					tokenInfo, found := mapToken[tx.TokenID]
					if !found {
						table.Append([]string{tx.LockTime, tx.Type, tx.TokenSymbol, strconv.FormatUint(tx.Amount, 10), tx.TxHash})
					} else
					{
						amountStr := getViewValueHelper(tx.Amount, tokenInfo.Decimal)
						table.Append([]string{tx.LockTime, tx.Type, tx.TokenSymbol, amountStr, tx.TxHash})
					}
				}
				table.Render()
			}
		}
	},
}

// Tx send cmd
var sendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "create PRV/pToken transaction",
	Run: func(cmd *cobra.Command, args []string) {
		receiver, _ := cmd.Flags().GetString("receiver")
		amount, _ := cmd.Flags().GetFloat64("amount")
		info, _ := cmd.Flags().GetString("info")
		token, _ := cmd.Flags().GetString("token")
		tokenID := token
		rawAmount := uint64(0)
		if tokenID != "" && tokenID != common.PRVID {
			if tokenInfo, err := getTokenBySymbol(token); err != nil {
				fmt.Println("Status: error ", "cannot get token detail")
			} else {
				tokenID = tokenInfo.ID
				rawAmount = uint64(amount * math.Pow10(tokenInfo.Decimal))
			}
		} else {
				tokenID = common.PRVID
				rawAmount = uint64(amount * math.Pow10(9))
		}

		fmt.Println("Query:", buildSendTxQuery(receiver, rawAmount, info, defaultFee, tokenID))
		fmt.Print("Password:")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}

		resBytes := sendTx(receiver, rawAmount, info, defaultFee, tokenID, string(password))
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			fmt.Println("Status: done")
			fmt.Println("Tx Hash:", result)
		}
	},
}

// Tx receive cmd
var receiveTxCmd = &cobra.Command{
	Use:   "receive",
	Short: "generate QR code to receive PRV or pToken",
	Run: func(cmd *cobra.Command, args []string) {
		accInfo, err := getInfo("")
		if err != nil {
			fmt.Println("Status: error - ", err.Error())
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Payment Address"})
			table.Append([]string{accInfo.PaymentAddress})
			table.Render()
		}
	},
}

// Tx info cmd
var txInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "get detail transaction info",
	Run: func(cmd *cobra.Command, args []string) {
		txHash, _ := cmd.Flags().GetString("hash")

		resBytes := getTxInfo(txHash)
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			txInfo := new(controllers.TxInfoJson)
			data, _ := json.Marshal(result)
			json.Unmarshal(data, txInfo)
			tmp, _ := json.MarshalIndent(txInfo, "", "   ")
			fmt.Println(string(tmp))
		}
	},
}
