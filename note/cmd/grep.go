package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/rameshg87/tools/note/lib"
	"github.com/spf13/cobra"
)

// grepCmd represents the grep command
var grepCmd = &cobra.Command{
	Use:   "grep",
	Short: "List notes having a particular string",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("note: No filename provided")
		}
		files, err := lib.Grep(args[0])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(strings.Join(lib.BaseNames(files), "\n"))
	},
}

func init() {
	RootCmd.AddCommand(grepCmd)
}
