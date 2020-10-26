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
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "wic",
	Short: "welcome to wic",
	Long: `__      __       .__                                  __            __      __.__        
/  \    /  \ ____ |  |   ____  ____   _____   ____   _/  |_  ____   /  \    /  \__| ____  
\   \/\/   // __ \|  | _/ ___\/  _ \ /     \_/ __ \  \   __\/  _ \  \   \/\/   /  |/ ___\ 
 \        /\  ___/|  |_\  \__(  <_> )  Y Y  \  ___/   |  | (  <_> )  \        /|  \  \___ 
  \__/\  /  \___  >____/\___  >____/|__|_|  /\___  >  |__|  \____/    \__/\  / |__|\___  >
       \/       \/          \/            \/     \/                        \/          \/ `,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("1.0.0")
	},
}

var Name string
var Mnemonic string
var Security int
var Network string
var Token string
var Hash string
var Fee uint64
var Info string
var Receiver string
var Amount float64
var PrvKey string
var Limit int
var Symbols string
var Pair string
var FromToken string
var ToToken string
var SendAmount float64
var MinAcceptable float64

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	//Wic commands
	rootCmd.AddCommand(walletCmd)
	rootCmd.AddCommand(accountCmd)
	rootCmd.AddCommand(txCmd)
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pdeCmd)
	rootCmd.AddCommand(minerCmd)

	//Wallet subcommands
	walletCmd.AddCommand(createWalletCmd)
	walletCmd.AddCommand(importWalletCmd)

	//Account subcommands
	accountCmd.AddCommand(addCmd)
	accountCmd.AddCommand(importAccCmd)
	accountCmd.AddCommand(lsCmd)
	accountCmd.AddCommand(switchCmd)
	accountCmd.AddCommand(infoCmd)
	accountCmd.AddCommand(exportCmd)

	//Tx subcommands
	txCmd.AddCommand(historyCmd)
	txCmd.AddCommand(sendTxCmd)
	txCmd.AddCommand(receiveTxCmd)
	txCmd.AddCommand(txInfoCmd)

	//Balance subcommands
	balanceCmd.AddCommand(utxoCmd)

	//Pde subcommands
	pdeCmd.AddCommand(pdeTxHistoryCmd)
	pdeCmd.AddCommand(pdePriceCmd)

	//Miner subcommands
	minerCmd.AddCommand(minerRewardCmd)


	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wic.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//Flags for create wallet
	createWalletCmd.Flags().IntVarP(&Security,"security","s", 128, "Add your security ")
	createWalletCmd.Flags().StringVarP(&Network,"network","n", "", "Add network info ")
	createWalletCmd.MarkFlagRequired("security")
	createWalletCmd.MarkFlagRequired("network")

	//Flags for import wallet
	importWalletCmd.Flags().StringVarP(&Mnemonic,"mnemonic","m", "", "Add your mnemonic")
	importWalletCmd.Flags().StringVarP(&Network,"network","n", "", "Add network info ")
	importWalletCmd.MarkFlagRequired("mnemonic")
	importWalletCmd.MarkFlagRequired("network")

	//Flags for add account
	addCmd.Flags().StringVarP(&Name,"name","n", "", "Add your name ")
	addCmd.MarkFlagRequired("name")

	//Import account need name, private key and passphrase
	importAccCmd.Flags().StringVarP(&Name,"name","n", "", "Add your name ")
	importAccCmd.Flags().StringVarP(&PrvKey,"privatekey","k", "", "Add your private key")
	importAccCmd.MarkFlagRequired("name")
	importAccCmd.MarkFlagRequired("privatekey")

	//Flags for switch account
	switchCmd.Flags().StringVarP(&Name,"name","n","","Switch account name")
	switchCmd.MarkFlagRequired("name")

	//Flags for get tx history
	historyCmd.Flags().StringVarP(&Token,"token","t","","Add your token")
	historyCmd.Flags().IntVarP(&Limit,"limit","l",20,"Add your limit")

	//Flags for get tx info
	txInfoCmd.Flags().StringVarP(&Hash,"hash","i","","Add your transaction hash")
	txInfoCmd.MarkFlagRequired("hash")

	//Flags for get tx send
	sendTxCmd.Flags().StringVarP(&Receiver, "receiver","r", "","Add receivers info")
	sendTxCmd.Flags().Float64VarP(&Amount, "amount","a", float64(0) ,"Add receiver amount")
	sendTxCmd.Flags().StringVarP(&Token,"token","t","","Add your token")
	sendTxCmd.Flags().StringVarP(&Info,"info","i","","Add your info")
	sendTxCmd.MarkFlagRequired("receiver")
	sendTxCmd.MarkFlagRequired("amount")

	//Flags for get balance
	balanceCmd.Flags().StringVarP(&Token,"token","t","","Add your token")

	//Flags for balance utxo
	utxoCmd.Flags().StringVarP(&Token,"token","t","","Add your token")

	//Flags for pde price cmd
	pdePriceCmd.Flags().StringVarP(&Pair,"pair","p","","Add your pair")
	pdePriceCmd.Flags().Float64VarP(&Amount,"amount","a", float64(0),"Add your pair")
	pdePriceCmd.MarkFlagRequired("pair")
	pdePriceCmd.MarkFlagRequired("amount")

	pdeTxHistoryCmd.Flags().StringVarP(&Pair,"pair","p","","Add your pair")
	pdeTxHistoryCmd.Flags().StringVarP(&Symbols,"symbol","s","","Add your symbol")
	pdeTxHistoryCmd.Flags().IntVarP(&Limit,"limit","l",20,"Add your limit")

	pdeTradeCmd.Flags().StringVarP(&FromToken, "fromtoken","f", "","Add from token")
	pdeTradeCmd.Flags().StringVarP(&ToToken, "totoken","t", "" ,"Add to token")
	pdeTradeCmd.Flags().Float64VarP(&SendAmount,"sendamount","a", float64(0),"Add your amount")
	pdeTradeCmd.Flags().Float64VarP(&MinAcceptable,"minacceptableloss","m", float64(1),"Add your min acceptable amount")
	pdeTradeCmd.MarkFlagRequired("fromtoken")
	pdeTradeCmd.MarkFlagRequired("totoken")
	pdeTradeCmd.MarkFlagRequired("sendamount")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".wic" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".wic")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

