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

// minerCmd
var minerCmd = &cobra.Command{
	Use:   "miner",
	Short: "miner command for committe info, reward",
}
// minerRewardCmd
var minerRewardCmd = &cobra.Command{
	Use:   "reward",
	Short: "miner reward",
	Run: func(cmd *cobra.Command, args []string) {
		resBytes :=  getCommitteeReward()
		result, appError := getResponse(resBytes)

		if appError != nil{
			fmt.Println("Status error - ",appError.Msg)
		} else {
			fmt.Println("Status: done")
			var listMinerInfo []*controllers.MinerInfoJson
			data,_ := json.Marshal(result)
			err := json.Unmarshal(data,&listMinerInfo)

			if err != nil{
				fmt.Println("Error from Unmarshal : ",err)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"PublicKey","Amount"})
			total := uint64(0)

			for _, minerInfo := range listMinerInfo{
				table.Append([]string{minerInfo.PaymentAddress,fmt.Sprintf("%v",minerInfo.Reward)})
				total += minerInfo.Reward
			}

			table.SetFooter([]string{ "TOTAL", strconv.FormatUint(total,10)})
			table.SetAutoMergeCells(true)
			table.SetRowLine(false)
			table.Render()
		}
	},
}

func getCommitteeReward() [] byte{
	url := "/miner/reward"
	query := fmt.Sprintf("{}")
	res, err := SendPostRequestWithQuery(query, url)
	if err != nil {
		log.Fatalln("Error send post request query")
	}
	return res
}
