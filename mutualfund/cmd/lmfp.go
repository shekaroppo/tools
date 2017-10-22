package cmd

import (
	"fmt"
	"log"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

// lmfpCmd represents the lmfp command
var lmfpCmd = &cobra.Command{
	Use:   "lmfp",
	Short: "List mutual fund purchases",
	Run: func(cmd *cobra.Command, args []string) {
		output, err := mflib.ListMutualFundPurchaseHelper()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(output)
	},
}

func init() {
	RootCmd.AddCommand(lmfpCmd)
}
