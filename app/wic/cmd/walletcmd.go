package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"incwallet/app/controllers"
	"log"
	"os"
	"strings"
	"syscall"
)

func createNewWallet(security int, passphrase, network string) []byte {
	url := "/wallet/create"
	query := fmt.Sprintf(`{
    	"security" : %v,
    	"passphrase" : "%s",
		"network" : "%s"
  		}`, security, passphrase, network)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

func importWallet(mnemonic, passphrase, network string) []byte {
	url := "/wallet/import"
	query := fmt.Sprintf(`{
    	"mnemonic" : "%s",
    	"passphrase" : "%s",
		"network" : "%s"
  		}`, mnemonic, passphrase, network)
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}

// walletCmd represent the wallet command
var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "wallet commands for import and create",
}

// importWalletCmd subcommand of wallet
var importWalletCmd = &cobra.Command{
	Use:   "import",
	Short: "import wallet by seed phrase",
	Run: func(cmd *cobra.Command, args []string) {
		mnemonic, _ := cmd.Flags().GetString("mnemonic")
		network, _ := cmd.Flags().GetString("network")
		fmt.Print("Enter Password:")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}
		fmt.Println()
		fmt.Print("Confirm Password:")
		confirmPassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}
		fmt.Println()
		if string(password) != string(confirmPassword) {
			fmt.Println("Status: error password is not correct")
			return
		}
		resBytes := importWallet(mnemonic, string(password), network)
		responseStatus := new(controllers.ResponseJson)
		json.Unmarshal(resBytes, responseStatus)
		if responseStatus.Error != nil {
			fmt.Println("Status: error - ", responseStatus.Error.Msg)
		} else {
			fmt.Println("Status: done")
		}
	},
}
// createWalletCmd subcommand of wallet
var createWalletCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new wallet",
	Run: func(cmd *cobra.Command, args []string) {
		security, _ := cmd.Flags().GetInt("security")
		network, _ := cmd.Flags().GetString("network")
		fmt.Print("Enter Password:")
		password, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}
		fmt.Println()
		fmt.Print("Confirm Password:")
		confirmPassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println()
			fmt.Println("Status: error " + err.Error())
		}
		fmt.Println()
		if string(password) != string(confirmPassword) {
			fmt.Println("Status: error password is not correct")
			return
		}
		resBytes := createNewWallet(security,string(password), network)
		responseStatus := new(controllers.ResponseJson)
		json.Unmarshal(resBytes, responseStatus)
		if responseStatus.Error != nil {
			fmt.Println("Status: error - ", responseStatus.Error.Msg)
		} else {
			fmt.Println("Status: done")
			fmt.Println("Please remember passphrase and backup mnemonic words")
			table := tablewriter.NewWriter(os.Stdout)
			words := strings.Fields(strings.Trim(fmt.Sprintf("%v",responseStatus.Msg), `"`))
			for i:= 0; i < len(words) / 4; i++ {
				table.Append([]string{
					fmt.Sprintf("%v %s", i*4 + 0, words[i*4 + 0]),
					fmt.Sprintf("%v %s", i*4 + 1, words[i*4 + 1]),
					fmt.Sprintf("%v %s", i*4 + 2, words[i*4 + 2]),
					fmt.Sprintf("%v %s", i*4 + 3, words[i*4 + 3]),
				})
			}
			table.Render()
		}
	},
}
