#define _GNU_SOURCE

#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mount.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <unistd.h>

static int do_mkdir(const char *path, mode_t mode) {
    if (mkdir(path, mode) < 0) {
        switch (errno) {
            case EEXIST:
                fprintf(stderr, "kernel already created: %s\n", path);
                break;
            default:
                fprintf(stderr, "cannot create directory %s: %m\n", path);
                return -1;
        }
    }
    return 0;
}

static int do_mkdir_many(const char **dnames, mode_t mode) {
    for (const char **dname = dnames; *dname != NULL; ++dname) {
        if (do_mkdir(*dname, mode) < 0) {
            return -1;
        }
    }
    return 0;
}

static int construct_skeleton_fs(void) {
    const char *dnames[] = {
		/* NOTE: kernel creates /dev and /root internally */
		/* NOTE: /proc is mounted by snapd-boot */
		/* NOTE: /sys is mounted by snapd-boot */
        "/bin", "/etc", "/run", "/tmp", "/var", "/var/lock", NULL,
    };
    if (do_mkdir_many(dnames, 0755) < 0) {
        return -1;
    }
    if (mount("udev", "/dev", "devtmpfs", MS_NOSUID, "mode=0755") < 0) {
        fprintf(stderr, "cannot mount /dev: %m\n");
        return -1;
    }
    if (do_mkdir("/dev/pts", 0755) < 0) {
        return -1;
    }
    if (mount("devpts", "/dev/pts", "devpts", MS_NOEXEC | MS_NOSUID, "gid=5,mode=0620") < 0) {
        fprintf(stderr, "cannot mount /dev/pts: %m\n");
        return -1;
    }
    if (mount("tmpfs", "/run", "tmpfs", MS_NOEXEC | MS_NOSUID, "size=10%,mode=0755") < 0) {
        fprintf(stderr, "cannot mount /run: %m\n");
        return -1;
    }
    if (do_mkdir("/run/initramfs", 0755) < 0) {
        return -1;
    }
    if (symlink("/proc/mounts", "/etc/mtab") < 0) {
        fprintf(stderr, "cannot symlink /proc/mounts as /etc/mtab: %m\n");
        return -1;
    }
    return 0;
}

static int construct_busybox_farm(void) {
    if (access("/busybox-static", X_OK) != 0) {
        return -1;
    }
    pid_t child = fork();
    if (child < 0) {
        fprintf(stderr, "cannot fork process for busybox setup: %m\n");
        return -1;
    }
    if (child == 0) {
        execl("/busybox-static", "sh", "-c",
              "for cmd in $(/busybox-static --list); do"
              "    /busybox-static ln -fs /busybox-static /bin/$cmd;"
              "done",
              NULL);
        fprintf(stderr, "cannot exec busybox setup script: %m\n");
        return -1;
    } else {
        int wstatus = 0;
        if (waitpid(child, &wstatus, 0) < 0) {
            fprintf(stderr, "cannot wait for busybox setup process: %m\n");
            return -1;
        }
        if (WIFEXITED(wstatus)) {
            int exit_code = WEXITSTATUS(wstatus);
            if (exit_code != 0) {
                fprintf(stderr, "busybox setup process failed with code %d\n", exit_code);
                return -1;
            }
        }
        if (WIFSIGNALED(wstatus)) {
            int sig_num = WTERMSIG(wstatus);
            fprintf(stderr, "busybox setup process killed by signal %d\n", sig_num);
            return -1;
        }
    }
    return 0;
}

int call_snapd_boot(void) {
    pid_t child = fork();
    if (child < 0) {
        fprintf(stderr, "cannot fork process for snapd-boot: %m\n");
        return -1;
    }
    if (child == 0) {
        execl("/snapd-boot", "snapd-boot", NULL);
        fprintf(stderr, "cannot exec snapd-boot: %m\n");
        return -1;
    } else {
        int wstatus = 0;
        if (waitpid(child, &wstatus, 0) < 0) {
            fprintf(stderr, "cannot wait for snapd-boot process: %m\n");
            return -1;
        }
        if (WIFEXITED(wstatus)) {
            int exit_code = WEXITSTATUS(wstatus);
            if (exit_code != 0) {
                fprintf(stderr, "snapd-boot failed with code %d\n", exit_code);
                return -1;
            }
        }
        if (WIFSIGNALED(wstatus)) {
            int sig_num = WTERMSIG(wstatus);
            fprintf(stderr, "snapd-boot killed by signal %d\n", sig_num);
            return -1;
        }
    }
    return 0;
}

int main(int argc, char **argv) {
    setlinebuf(stdout);
    printf("snapd-init\n");
    if (call_snapd_boot() < 0) {
        return 1;
    }
    if (construct_skeleton_fs() < 0) {
        return 1;
    }
    if (construct_busybox_farm() == 0) {
        execl("/busybox-static", "sh", NULL);
    }
    pause();
    return 0;
}
