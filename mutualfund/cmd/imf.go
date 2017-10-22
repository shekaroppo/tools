package cmd

import (
	"log"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

// imfCmd represents the imf command
var imfCmd = &cobra.Command{
	Use:   "imf",
	Short: "Create a new mutual fund",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 5 {
			log.Fatal("Invalid command line arguments")
		}
		mflib.InsertMutualFundHelper(
			args[1], args[2], args[3], args[4], args[5])
	},
}

func init() {
	RootCmd.AddCommand(imfCmd)
}
