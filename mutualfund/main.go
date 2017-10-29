package main

import (
	"log"
	"os"

	"github.com/rameshg87/tools/mutualfund/mflib"
)

func main() {
	log.SetFlags(0)
	app := mflib.GetCliApp()
	app.Run(os.Args)
}
