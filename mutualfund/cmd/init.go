package cmd

import (
	"log"

	"github.com/rameshg87/tools/mutualfund/mflib"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		err := mflib.InitDb()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
