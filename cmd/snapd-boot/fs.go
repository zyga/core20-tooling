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
	if err := mountRun(); err != nil {
		return err
	}
	paths := []string{"/bin", "/etc", "/tmp", "/var", "/var/lock", "/run/initramfs"}
	for _, path := range paths {
		if err := mkdir(path, 0755); err != nil {
			return err
		}
	}
	if err := syscall.Symlink("/proc/mounts", "/etc/mtab"); err != nil {
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

func mountRun() error {
	const source = "tmpfs"
	const target = "/run"
	const fstype = "tmpfs"
	const flags = syscall.MS_NOEXEC | syscall.MS_NOSUID
	return mountFundamental(source, target, fstype, flags, "size=10%,mode=0755")
}

func mountFundamental(source, target, fstype string, flags uintptr, data string) error {
	if err := mkdir(target, 0755); err != nil {
		return err
	}
	return syscall.Mount(source, target, fstype, flags, data)
}

// The kernel creates /dev/console so some directories may already exist.
func mkdir(path string, mode os.FileMode) error {
	if err := os.Mkdir(path, mode); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
