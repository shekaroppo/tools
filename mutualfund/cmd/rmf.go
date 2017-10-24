package cmd

import (
	"log"
	"strconv"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

// rmfCmd represents the rmf command
var rmfCmd = &cobra.Command{
	Use:   "rmf",
	Short: "Remove mutual fund",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Usage: mutualfund rmf <mfid>")
		}
		mfid, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}
		mflib.RemoveMutualFundHelper(mfid)
	},
}

func init() {
	RootCmd.AddCommand(rmfCmd)
}
