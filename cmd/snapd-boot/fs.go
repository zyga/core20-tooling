package main

import (
	"os"
	"syscall"
)

func createSkeletonFileSystem() error {
	if err := mountProc(); err != nil {
		return err
	}
	if err := mountSys(); err != nil {
		return err
	}
	if err := mountDev(); err != nil {
		return err
	}
	if err := mountDevPts(); err != nil {
		return err
	}
	return nil
}

func mountProc() error {
	const source = "proc"
	const target = "/proc"
	const fstype = "proc"
	const flags = syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID
	return mountFundamental(source, target, fstype, flags, "")
}

func mountSys() error {
	const source = "sysfs"
	const target = "/sys"
	const fstype = "sysfs"
	const flags = syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID
	return mountFundamental(source, target, fstype, flags, "")
}

func mountDev() error {
	const source = "udev"
	const target = "/dev"
	const fstype = "devtmpfs"
	const flags = syscall.MS_NOSUID
	return mountFundamental(source, target, fstype, flags, "mode=0755")
}

func mountDevPts() error {
	const source = "devpts"
	const target = "/dev/pts"
	const fstype = "devpts"
	const flags = syscall.MS_NOEXEC | syscall.MS_NOSUID
	// XXX: What is the significance of gid=5?
	// How does it affect base snaps?
	return mountFundamental(source, target, fstype, flags, "gid=5,mode=0620")
}

func mountFundamental(source, target, fstype string, flags uintptr, data string) error {
	// The kernel creates /dev/console so some directories may already exist.
	if err := os.Mkdir(target, 0755); err != nil && !os.IsExist(err) {
		return err
	}
	return syscall.Mount(source, target, fstype, flags, data)
}
