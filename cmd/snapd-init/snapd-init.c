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
    if (construct_busybox_farm() == 0) {
        execl("/busybox-static", "sh", NULL);
    }
    pause();
    return 0;
}
