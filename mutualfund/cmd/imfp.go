package cmd

import (
	"log"
	"strconv"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

// imfpCmd represents the imfp command
var imfpCmd = &cobra.Command{
	Use:   "imfp",
	Short: "Insert the record of mutual fund purchase",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 4 {
			log.Fatal("Usage: mutualfund imfp <mfid> <amount> <nav> <date>")
		}
		mfid, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		nav, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		amount, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			log.Fatal(err)
		}
		dateStr := args[3]
		mflib.InsertMutualFundPurchaseHelper(mfid, nav, amount, dateStr)
	},
}

func init() {
	RootCmd.AddCommand(imfpCmd)
}
