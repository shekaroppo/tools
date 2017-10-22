package cmd

import (
	"fmt"
	"log"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

// lmfCmd represents the lmf command
var lmfCmd = &cobra.Command{
	Use:   "lmf",
	Short: "List all mutual funds",
	Run: func(cmd *cobra.Command, args []string) {
		output, err := mflib.ListMutualFundHelper()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(output)
	},
}

func init() {
	RootCmd.AddCommand(lmfCmd)
}
