package cmd

import (
	"fmt"
	"log"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

var sortBy string

// smfsCmd represents the smfs command
var smfsCmd = &cobra.Command{
	Use:   "smfs",
	Short: "Summary of mutual fund investments",
	Run: func(cmd *cobra.Command, args []string) {
		var sortArg string
		if sortBy == "appr" {
			sortArg = "appreciation"
		} else if sortBy == "" {
			sortArg = ""
		} else {
			log.Fatal("Unknown sortBy ", sortBy)
		}
		output, err := mflib.MutualFundSummaryHelper(nil, nil, sortArg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(output)
	},
}

func init() {
	RootCmd.AddCommand(smfsCmd)
	smfsCmd.Flags().StringVarP(&sortBy, "sort", "s", "", "sort output")
}
