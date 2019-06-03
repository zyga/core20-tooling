package main

import "fmt"

func main() {
	fmt.Println("snapd-boot")
	// TODO:
	// - determine system boot mode, either RUN or RECOVERY.
	// - determine current BOOT-SET (should be "factory" in this demo)
	// - mount the snapd-systems partition
	// - mount the current kernel snap from the snapd-systems partition
	// - extract the initrd from the kernel snap
	// - chain load the initrd from the kernel snap
}
