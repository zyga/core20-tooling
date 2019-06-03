package main

import (
	"fmt"
	"io/ioutil"

	"github.com/google/shlex"
)

// kernelArgs returns the kernel command line, as read from /proc/cmdline
func kernelArgs() ([]string, error) {
	bytes, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return nil, err
	}
	return shlex.Split(string(bytes))
}

func main() {
	fmt.Println("snapd-boot")
	args, err := kernelArgs()
	if err != nil {
		fmt.Printf("cannot read kernel command line: %s", err)
		return
	}
	fmt.Printf("kernel command line: %#v\n", args)
	// TODO:
	// - determine system boot mode, either RUN or RECOVERY.
	// - determine current BOOT-SET (should be "factory" in this demo)
	// - mount the snapd-systems partition
	// - mount the current kernel snap from the snapd-systems partition
	// - extract the initrd from the kernel snap
	// - chain load the initrd from the kernel snap
}
