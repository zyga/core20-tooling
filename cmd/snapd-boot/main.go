package main

import (
	"fmt"
	"os"
	"time"
)

func run() error {
	if err := createSkeletonFileSystem(); err != nil {
		return err
	}
	args, err := kernelArgs()
	if err != nil {
		return fmt.Errorf("cannot read kernel command line: %s", err)
	}
	fmt.Printf("kernel command line: %#v\n", args)
	if err := spawnBusyBox(); err != nil {
		return err
	}
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
		time.Sleep(10 * time.Second)
		os.Exit(1)
	}
}
