package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func spawnBusyBox() error {
	// Ask busybox about the kind of tools it supports.
	cmd := exec.Command("/busybox-static", "busybox-static", "--list")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(stdout)
	if err := cmd.Start(); err != nil {
		return err
	}
	var hasSh bool
	var tools []string
	for {
		tool, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		tool = strings.TrimSuffix(tool, "\n")
		if tool == "sh" {
			hasSh = true
		}
		tools = append(tools, tool)
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	if !hasSh {
		return fmt.Errorf("busybox-static does not provide sh")
	}
	// Populate /bin with symlinks to busybox-static.
	for _, tool := range tools {
		if err := os.Symlink("/busybox-static", filepath.Join("/bin", tool)); err != nil {
			return err
		}
	}
	// Run /bin/sh
	return syscall.Exec("/bin/sh", []string{"sh"}, []string{"PATH=/bin"})
}
