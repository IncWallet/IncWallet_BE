/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"incwallet/app/controllers"
	"log"
	"os"
	"strconv"
)

func getBalance(token string) []byte {
	url := "/accounts/balance"
	query := fmt.Sprintf(`{
    	"tokenid" : "%s"
  		}`, token)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func getUnspent(token string) []byte {
	url := "/accounts/unspent"
	query := fmt.Sprintf(`{
    	"tokenid" : "%s"
  		}`, token)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

// balanceCmd
var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "show current balance",
	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		tokenID := token
		if tokenID != "" {
			if tokenInfo, err := getTokenBySymbol(token); err == nil {
				tokenID = tokenInfo.ID
			}
		}

		resBytes := getBalance(tokenID)
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			var listBalance []controllers.AccBalanceJson
			data, _ := json.Marshal(result)
			json.Unmarshal(data, &listBalance)

			mapToken, err := getAllToken()
			if err != nil {
				fmt.Println("Status: error - ", err)
			} else {
				fmt.Println("Status: done")
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Token ID", "Token", "Amount"})
				for _, v := range listBalance {
					tokenInfo, found := mapToken[v.TokenID]
					if !found {
						table.Append([]string{v.TokenID, v.TokenSymbol, strconv.FormatUint(v.Amount, 10)})
					} else
					{
						amountStr := getViewValueHelper(v.Amount, tokenInfo.Decimal)
						table.Append([]string{v.TokenID, v.TokenSymbol, amountStr})
					}
				}
				table.Render()
			}
		}
	},
}

// unspendCmd
var utxoCmd = &cobra.Command{
	Use:   "utxo",
	Short: "show current unspent coins",
	Run: func(cmd *cobra.Command, args []string) {
		token, _ := cmd.Flags().GetString("token")
		tokenID := token
		if tokenID != "" {
			if tokenInfo, err := getTokenBySymbol(token); err == nil {
				tokenID = tokenInfo.ID
			}
		}

		resBytes := getUnspent(tokenID)
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			var listBalance []controllers.CoinJson
			data, _ := json.Marshal(result)
			json.Unmarshal(data, &listBalance)

			mapToken, err := getAllToken()
			if err != nil {
				fmt.Println("Status: error - ", err)
			} else {
				fmt.Println("Status: done")

				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Coin Commitment", "Token", "Value"})
				for _, v := range listBalance {
					tokenInfo, found := mapToken[v.TokenID]
					if !found {
						table.Append([]string{
							v.CoinCommitment,
							v.TokenSymbol,
							v.Value,
						})
					} else
					{
						amount, _ := strconv.ParseUint(v.Value, 10, 64)
						amountStr := getViewValueHelper(amount, tokenInfo.Decimal)
						table.Append([]string{
							v.CoinCommitment,
							v.TokenSymbol,
							amountStr,
						})
					}
				}
				table.Render()
			}
		}
	},
}
