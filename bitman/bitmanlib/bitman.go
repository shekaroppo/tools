package bitmanlib

import (
	"errors"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli"
)

type CommandArgumentError int

func (e CommandArgumentError) Error() string {
	return "This command exactly requires " + strconv.Itoa(int(e)) + " arguments"
}

func GetBitField(numStr string, startBit, stopBit int) (string, error) {
	var base int
	numBits := stopBit - startBit + 1
	one := new(big.Int)
	one.SetInt64(1)

	if strings.HasPrefix(numStr, "0x") {
		numStr = numStr[2:]
		base = 16
	} else {
		base = 10
	}

	bigNum := new(big.Int)
	_, ok := bigNum.SetString(numStr, base)
	if !ok {
		return "", errors.New("Unable to parse '" + numStr + "' with base=" + strconv.Itoa(base))
	}

	bigNum.Rsh(bigNum, uint(startBit))

	mask := new(big.Int)
	mask.SetBit(mask, numBits, 1)
	mask.Sub(mask, one)

	bigNum.And(bigNum, mask)
	return bigNum.String(), nil
}

func SetBitField(numStr string, startBit, stopBit int, valueStr string) (string, error) {
	var numBase, valueBase int
	if strings.HasPrefix(numStr, "0x") {
		numStr = numStr[2:]
		numBase = 16
	} else {
		numBase = 10
	}
	if strings.HasPrefix(valueStr, "0x") {
		valueStr = valueStr[2:]
		valueBase = 16
	} else {
		valueBase = 10
	}

	one := new(big.Int)
	one.SetInt64(1)

	maxValue := new(big.Int)
	maxValue.SetBit(maxValue, stopBit-startBit+1, 1)
	maxValue.Sub(maxValue, one)
	valueBigNum := new(big.Int)
	_, ok := valueBigNum.SetString(valueStr, valueBase)
	if !ok {
		return "", errors.New("Unable to parse '" + valueStr + "' with base=" + strconv.Itoa(numBase))
	}

	bigNum := new(big.Int)
	_, ok = bigNum.SetString(numStr, numBase)
	if !ok {
		return "", errors.New("Unable to parse '" + numStr + "' with base=" + strconv.Itoa(numBase))
	}
	bitLen := bigNum.BitLen()

	if valueBigNum.Cmp(maxValue) == 1 {
		// value is greater than maxValue.
		return "", errors.New("Value '" + valueStr + "' is greater than allowed value=" + maxValue.Text(10))
	}

	upperOnesLen := bitLen - stopBit
	upperOnesMask := new(big.Int)
	upperOnesMask.SetBit(upperOnesMask, upperOnesLen, 1)
	upperOnesMask.Sub(upperOnesMask, one)
	upperOnesMask.Lsh(upperOnesMask, uint(stopBit))
	lowerOnesMask := new(big.Int)
	lowerOnesMask.SetBit(lowerOnesMask, startBit, 1)
	lowerOnesMask.Sub(lowerOnesMask, one)
	mask := upperOnesMask.Or(upperOnesMask, lowerOnesMask)

	bigNum.And(bigNum, mask)
	valueBigNum.Lsh(valueBigNum, uint(startBit))
	bigNum.Or(bigNum, valueBigNum)
	return "0x" + bigNum.Text(16), nil
}

func PrintCliHelper(c *cli.Context) error {
	var msg string
	if c.NArg() != 3 {
		err := CommandArgumentError(3)
		log.Println(err)
		return err
	}
	numStr := c.Args().Get(0)
	startBitStr := c.Args().Get(1)
	startBit, err := strconv.Atoi(startBitStr)
	if err != nil {
		msg = "Invalid start bit '" + startBitStr + "'"
		log.Println(msg)
		return errors.New(msg)
	}
	stopBitStr := c.Args().Get(2)
	stopBit, err := strconv.Atoi(stopBitStr)
	if err != nil {
		msg = "Invalid stop bit '" + stopBitStr + "'"
		log.Println(msg)
		return errors.New(msg)
	}
	value, err := GetBitField(numStr, startBit, stopBit)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(value)
	return nil
}

func ModifyCliHelper(c *cli.Context) error {
	var msg string
	if c.NArg() != 4 {
		err := CommandArgumentError(4)
		log.Println(err)
		return err
	}
	numStr := c.Args().Get(0)
	startBitStr := c.Args().Get(1)
	startBit, err := strconv.Atoi(startBitStr)
	if err != nil {
		msg = "Invalid start bit '" + startBitStr + "'"
		log.Println(msg)
		return errors.New(msg)
	}
	stopBitStr := c.Args().Get(2)
	stopBit, err := strconv.Atoi(stopBitStr)
	if err != nil {
		msg = "Invalid stop bit '" + stopBitStr + "'"
		log.Println(msg)
		return errors.New(msg)
	}
	newValueStr := c.Args().Get(3)
	value, err := SetBitField(numStr, startBit, stopBit, newValueStr)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(value)
	return nil
}

func GetCliApp() *cli.App {
	app := cli.NewApp()
	log.SetFlags(0)
	app.Name = "bitman"
	app.Usage = "Bit manipulation tool for big numbers"
	app.Version = "1.0.0"
	app.Commands = append(
		app.Commands,
		cli.Command{
			Name:    "print",
			Aliases: []string{"p"},
			Usage:   "Print the selected bits",
			UsageText: `
			This command is used for printing the value contained
			in between the <start> and <stop> bits of a (big) number.
			The bit position starts at 0.

			Usage:
			bitman p <big number> <start bit> <stop bit>

			Example:
			The below command prints 17187 (value contained inbetween
			bit positions 77 and 92).
			bitman p 0x86460000000000000000000 77 92`,
			Action: PrintCliHelper,
		},
		cli.Command{
			Name:    "modify",
			Aliases: []string{"m"},
			Usage:   "Modify the selected bits of a value",
			UsageText: `
			This command is used for setting the value contained
			in between the <start> and <stop> bits of a (big) number.
			The bit position starts at 0.

			Usage:
			bitman m <big number> <start bit> <stop bit> <new value>

			Example:
			The value contained in between bit positions 77 and 92 for
			the below number is 17187. The make it 16000, the below
			command may be used.

			bitman m 0x86460000000000000000000 77 92 16000`,
			Action: ModifyCliHelper,
		},
	)
	return app
}

func main() {
	app := GetCliApp()
	app.Run(os.Args)
}
