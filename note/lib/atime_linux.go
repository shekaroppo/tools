// +build linux

package lib

import (
	"os"
	"syscall"
)

func Atime(filename string) int64 {
	fi, _ := os.Stat(filename)
	stat := fi.Sys().(*syscall.Stat_t)
	return stat.Atim.Sec
}
