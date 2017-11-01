package main

import (
	"os"

	"github.com/rameshg87/tools/todo/todolib"
)

func main() {
	app := todolib.GetApp()
	app.Run(os.Args)
}
