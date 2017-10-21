package cmd

import (
	"log"

	"github.com/rameshg87/tools/note/lib"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean all editor temp files from NOTES_DIR",
	Run: func(cmd *cobra.Command, args []string) {
		err := lib.Clean()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanCmd)
}
