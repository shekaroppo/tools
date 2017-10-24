package cmd

import (
	"log"
	"strconv"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

// rmfpCmd represents the rmfp command
var rmfpCmd = &cobra.Command{
	Use:   "rmfp",
	Short: "Remove a mutual fund purchase",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Usage: mutualfund rmfp <mfpid>")
		}
		mfpid, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		mflib.RemoveMutualFundPurchaseHelper(mfpid)
	},
}

func init() {
	RootCmd.AddCommand(rmfpCmd)
}
