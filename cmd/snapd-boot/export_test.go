package main

import (
	"io"
	"io/ioutil"

	. "gopkg.in/check.v1"
)

var (
	KernelArgs = kernelArgs
)

func MockProcCmdLine(c *C, text string) (restore func()) {
	old := procCmdLine
	f, err := ioutil.TempFile(c.MkDir(), "cmdline")
	c.Assert(err, IsNil)
	defer f.Close()
	_, err = io.WriteString(f, text)
	c.Assert(err, IsNil)
	procCmdLine = f.Name()
	return func() {
		procCmdLine = old
	}
}
