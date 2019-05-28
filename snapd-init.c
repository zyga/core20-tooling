#define _GNU_SOURCE

#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mount.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

int main(int argc, char **argv) {
    printf("snapd-init started...\n");
    fflush(stdout);

    mkdir("/dev", 0755);
    mkdir("/proc", 0755);
    mkdir("/root", 0700);
    mkdir("/run", 0755);
    mkdir("/sys", 0755);
    mkdir("/tmp", 0755);
    mkdir("/var", 0755);
    mkdir("/var/lock", 0755);

    if (mount("sysfs", "/sys", "sysfs", MS_NODEV | MS_NOEXEC | MS_NOSUID, NULL) < 0) {
        perror("cannot mount /sys");
        return 1;
    }
    if (mount("proc", "/proc", "proc", MS_NODEV | MS_NOEXEC | MS_NOSUID, NULL) < 0) {
        perror("cannot mount /proc");
        return 1;
    }
    if (mount("udev", "/dev", "devtmpfs", MS_NOSUID, "mode=0755") < 0) {
        perror("cannot mount /dev");
        return 1;
    }
    mkdir("/dev/pts", 0755);
    if (mount("devpts", "/dev/pts", "devpts", MS_NOEXEC | MS_NOSUID, "gid=5,mode=0620") < 0) {
        perror("cannot mount /dev/pts");
        return 1;
    }
    if (mount("tmpfs", "/run", "tmpfs", MS_NOEXEC | MS_NOSUID, "size=10%,mode=0755") < 0) {
        perror("cannot mount /run");
        return 1;
    }
    mkdir("/run/initramfs", 0755);
    mkdir("/bin", 0755);
    mkdir("/etc", 0755);
    symlink("/proc/mounts", "/etc/mtab");

    printf("essential file-systems mounted\n");
    fflush(stdout);

    pid_t child = fork();
    if (child < 0) {
        perror("cannot fork helper process");
        return 1;
    }
    if (child == 0) {
        execl("/busybox-static", "sh", "-c",
              "for cmd in $(/busybox-static --list); do /busybox-static ln -s /busybox-static /bin/$cmd; done;", NULL);
        perror("cannot exec busybox boostrap script");
        return 1;
    } else {
        int wstatus = 0;
        if (waitpid(child, &wstatus, 0) < 0) {
            perror("cannot wait for helper process");
            return 1;
        }
        if (WIFEXITED(wstatus)) {
            int exit_code = WEXITSTATUS(wstatus);
            if (exit_code != 0) {
                perror("helper process failed");
            }
        }
        if (WIFSIGNALED(wstatus)) {
            perror("helper process killed by signal");
        }
    }
    printf("/bin populated with busybox symlinks\n");
    fflush(stdout);

    execl("/busybox-static", "sh", NULL);

    // TODO:
    // - determine system boot mode, either RUN or RECOVERY.
    // - determine current BOOT-SET (should be "factory" in this demo)
    // - mount the snapd-systems partition
    // - mount the current kernel snap from the snapd-systems partition
    // - extract the initrd from the kernel snap
    // - chain load the initrd from the kernel snap

    pause();
    return 0;
}
