package main

import (
	"os/exec"
	"syscall"
)

func spawnBusyBox() error {
	const initScript = `
	for cmd in $(/busybox-static --list); do
	    /busybox-static ln -fs /busybox-static /bin/$cmd;
	done
	`
	cmd := exec.Command("/busybox-static", "sh", "-c", initScript)
	if err := cmd.Run(); err != nil {
		return err
	}
	return syscall.Exec("/bin/sh", []string{"sh"}, []string{"PATH=/bin"})
}
