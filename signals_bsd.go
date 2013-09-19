// +build !linux

package main

import (
	"syscall"
)

func init() {
	passSigs = append(passSigs, syscall.SIGINFO)
}
