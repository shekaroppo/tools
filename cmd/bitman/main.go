package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rameshg87/tools/bitman"
)

func main() {
	if len(os.Args) < 4 || len(os.Args) > 5 {
		fmt.Println("Usage: bitman <number> <startbit> <stopbit> [<newvalue>]")
		os.Exit(1)
	}
	numStr := os.Args[1]
	startBit, _ := strconv.Atoi(os.Args[2])
	stopBit, _ := strconv.Atoi(os.Args[3])
	if len(os.Args) == 4 {
		fmt.Println(bitman.GetBitField(numStr, startBit, stopBit))
	} else {
		newValue := os.Args[4]
		fmt.Println(bitman.SetBitField(numStr, startBit, stopBit, newValue))
	}
}
