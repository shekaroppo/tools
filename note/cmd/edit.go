package cmd

import (
	"log"

	"github.com/rameshg87/tools/note/lib"
	"github.com/spf13/cobra"
)

var create bool

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a note",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("note: No filename provided")
		}
		err := lib.Edit(args[0], create)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
	editCmd.Flags().BoolVarP(&create, "create", "c", false, "create")
}
