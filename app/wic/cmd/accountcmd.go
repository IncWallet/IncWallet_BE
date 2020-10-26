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
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"incwallet/app/controllers"
	"log"
	"os"
	"syscall"
)

func addAccount(name string, passphrase string) []byte {
	url := "/accounts/add"
	query := fmt.Sprintf(`{
    	"name" : "%s",
    	"passphrase" : "%s"
  		}`, name, passphrase)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func importAccount(name string, privateKey string, passphrase string) []byte {
	url := "/accounts/import"
	query := fmt.Sprintf(`{
    	"name" : "%s",
    	"privateKey" : "%s",
    	"passphrase" : "%s"
  		}`, name, privateKey, passphrase)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func switchAccount(name string, passphrase string) []byte {
	url := "/accounts/switch"
	query := fmt.Sprintf(`{
    	"name" : "%s",
    	"passphrase" : "%s"
  		}`, name, passphrase)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func listAccount() []byte {
	url := "/accounts/list"
	query := "{}"
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func getInfo(passphrase string) (*controllers.InfoJson, error) {
	url := "/accounts/info"
	query := fmt.Sprintf(`{
		"passphrase": "%s"
	}`, passphrase)
	resBytes, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	result, appError := getResponse(resBytes)
	if appError != nil {
		return nil, errors.New(appError.Msg)
	}
	accInfo := new(controllers.InfoJson)
	data, _ := json.Marshal(result)
	json.Unmarshal(data, &accInfo)
	return accInfo, nil
}



// accountCmd represent the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "account commands for add, import, switch, and ls" +
		"" +
		"	",
}

// addCmd subcommand of account
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add new account",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		fmt.Print("Enter Password:")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}
		fmt.Println()
		resBytes := addAccount(name, string(password))
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			fmt.Println("Status: done")

			accInfo := new(controllers.InfoJson)
			data, _ := json.Marshal(result)
			json.Unmarshal(data, accInfo)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Account", accInfo.AccountName})
			table.Append([]string{"Payment Address", accInfo.PaymentAddress})
			table.Append([]string{"Public Key", accInfo.PublicKey})
			table.Append([]string{"Viewing Key", accInfo.ViewingKey})
			table.Append([]string{"Mining Key", accInfo.MiningKey})
			table.Append([]string{"Network", accInfo.Network})
			table.Render()
		}
	},
}

// importAccCmd subcommand of account
var importAccCmd = &cobra.Command{
	Use:   "import",
	Short: "import account by private key",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		privateKey, _ := cmd.Flags().GetString("privatekey")

		fmt.Print("Enter Password:")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}
		fmt.Println()

		resBytes := importAccount(name, privateKey, string(password))
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			fmt.Println("Status: done")

			accInfo := new(controllers.InfoJson)
			data, _ := json.Marshal(result)
			json.Unmarshal(data, accInfo)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Account", accInfo.AccountName})
			table.Append([]string{"Payment Address", accInfo.PaymentAddress})
			table.Append([]string{"Public Key", accInfo.PublicKey})
			table.Append([]string{"Viewing Key", accInfo.ViewingKey})
			table.Append([]string{"Mining Key", accInfo.MiningKey})
			table.Append([]string{"Network", accInfo.Network})
			table.Render()
		}
	},
}

// lsCmd subcommand of account
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "show all imported/added accounts",
	Run: func(cmd *cobra.Command, args []string) {
		resBytes :=listAccount()
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			fmt.Println("Status: done")
			var listAccount []controllers.AccountJson
			data, _ := json.Marshal(result)
			json.Unmarshal(data, &listAccount)

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Payment Address"})
			for _, v := range listAccount {
				table.Append([]string{v.Name, v.PaymentAddress})
			}
			table.Render()
		}
	},
}

// switchCmd subcommand of account
var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "switch between accounts by name",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		fmt.Print("Password:")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}
		fmt.Println()

		resBytes := switchAccount(name,string(password))
		result, appError := getResponse(resBytes)
		if appError != nil {
			fmt.Println("Status: error - ", appError.Msg)
		} else {
			fmt.Println("Status: done")

			accInfo := new(controllers.InfoJson)
			data, _ := json.Marshal(result)
			json.Unmarshal(data, accInfo)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Account", accInfo.AccountName})
			table.Append([]string{"Payment Address", accInfo.PaymentAddress})
			table.Append([]string{"Public Key", accInfo.PublicKey})
			table.Append([]string{"Viewing Key", accInfo.ViewingKey})
			table.Append([]string{"Mining Key", accInfo.MiningKey})
			table.Append([]string{"Network", accInfo.Network})
			table.Render()
		}
	},
}

// infoCmd
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "show wallet info",
	Run: func(cmd *cobra.Command, args []string) {
		accInfo, err := getInfo("")
		if err != nil {
			fmt.Println("Status: error - ", err.Error())
		} else {
			fmt.Println("Status: done")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Account", accInfo.AccountName})
			table.Append([]string{"Private Key", accInfo.PrivateKey})
			table.Append([]string{"Payment Address", accInfo.PaymentAddress})
			table.Append([]string{"Public Key", accInfo.PublicKey})
			table.Append([]string{"Viewing Key", accInfo.ViewingKey})
			table.Append([]string{"Mining Key", accInfo.MiningKey})
			table.Append([]string{"Network", accInfo.Network})
			table.Render()
		}
	},
}

// exportCmd
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export private key & info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Password:")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}
		fmt.Println()
		accInfo, err := getInfo(string(password))
		if err != nil {
			fmt.Println("Status: error - ", err.Error())
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Account", accInfo.AccountName})
			table.Append([]string{"Private Key", accInfo.PrivateKey})
			table.Append([]string{"Payment Address", accInfo.PaymentAddress})
			table.Append([]string{"Public Key", accInfo.PublicKey})
			table.Append([]string{"Viewing Key", accInfo.ViewingKey})
			table.Append([]string{"Mining Key", accInfo.MiningKey})
			table.Append([]string{"Network", accInfo.Network})
			table.Render()
		}
	},
}



