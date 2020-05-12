// +build !windows

package proc

import (
	"syscall"
)

func GetSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

func TerminateProc(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}
