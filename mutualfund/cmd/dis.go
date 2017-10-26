package cmd

import (
	"fmt"
	"log"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

// disCmd represents the smfs command
var disCmd = &cobra.Command{
	Use:   "dis",
	Short: "Distribution of mutual fund investments",
	Run: func(cmd *cobra.Command, args []string) {
		output, err := mflib.MutualFundDisHelper()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(output)
	},
}

func init() {
	RootCmd.AddCommand(disCmd)
}
