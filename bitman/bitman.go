package bitman

import (
	"math/big"
	"strings"
)

func GetBitField(numStr string, startBit, stopBit int) string {
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
	bigNum.SetString(numStr, base)
	bigNum.Rsh(bigNum, uint(startBit))

	mask := new(big.Int)
	mask.SetBit(mask, numBits, 1)
	mask.Sub(mask, one)

	bigNum.And(bigNum, mask)
	return bigNum.String()
}

func SetBitField(numStr string, startBit, stopBit int, valueStr string) string {
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
	valueBigNum.SetString(valueStr, valueBase)
	if valueBigNum.Cmp(maxValue) == 1 {
		// value is greater than maxValue.
		return ""
	}

	bigNum := new(big.Int)
	bigNum.SetString(numStr, numBase)
	bitLen := bigNum.BitLen()

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
	return "0x" + bigNum.Text(16)
}
