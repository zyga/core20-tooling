package main

import (
	"io/ioutil"

	"github.com/google/shlex"
)

var procCmdLine = "/proc/cmdline"

// kernelArgs returns the kernel command line, as read from /proc/cmdline
func kernelArgs() ([]string, error) {
	bytes, err := ioutil.ReadFile(procCmdLine)
	if err != nil {
		return nil, err
	}
	return shlex.Split(string(bytes))
}
