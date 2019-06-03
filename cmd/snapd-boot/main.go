package main

import (
	"fmt"
	"os"
	"syscall"
)

func mountProc() error {
	const path = "/proc"
	err := os.Mkdir(path, 0755)
	if err != nil && os.IsExist(err) {
		return err
	}
	return syscall.Mount("proc", path, "proc",
		syscall.MS_NODEV|syscall.MS_NOEXEC|syscall.MS_NOSUID, "")
}

func run() error {
	err := mountProc()
	if err != nil {
		return err
	}
	args, err := kernelArgs()
	if err != nil {
		return fmt.Errorf("cannot read kernel command line: %s", err)
	}
	fmt.Printf("kernel command line: %#v\n", args)
	// TODO:
	// - determine system boot mode, either RUN or RECOVERY.
	// - determine current BOOT-SET (should be "factory" in this demo)
	// - mount the snapd-systems partition
	// - mount the current kernel snap from the snapd-systems partition
	// - extract the initrd from the kernel snap
	// - chain load the initrd from the kernel snap
	return nil
}

func main() {
	fmt.Println("snapd-boot")
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
