package main_test

import (
	. "gopkg.in/check.v1"

	main "github.com/zyga/core20-tooling/cmd/snapd-boot"
)

type utilsSuite struct{}

var _ = Suite(&utilsSuite{})

func (*utilsSuite) TestSmoke(c *C) {
	restore := main.MockProcCmdLine(c, "BOOT_IMAGE=(kernel_snap)/kernel.img ro quiet console=ttyS0 console=tty panic=-1")
	defer restore()

	args, err := main.KernelArgs()
	c.Assert(err, IsNil)
	c.Check(args, DeepEquals, []string{"BOOT_IMAGE=(kernel_snap)/kernel.img", "ro", "quiet", "console=ttyS0", "console=tty", "panic=-1"})
}
