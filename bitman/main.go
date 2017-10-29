package main

import (
	"os"

	"github.com/rameshg87/tools/bitman/bitmanlib"
)

func main() {
	app := bitmanlib.GetCliApp()
	app.Run(os.Args)
}
