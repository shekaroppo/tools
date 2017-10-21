package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/rameshg87/tools/note/lib"
	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List notes",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := lib.List()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(strings.Join(lib.BaseNames(files), "\n"))
	},
}

func init() {
	RootCmd.AddCommand(lsCmd)
}
