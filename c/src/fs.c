#include "winescape_fs.h"
#include "winescape.h"
#include <errno.h>

int ws_open(const char *path, int flags, unsigned mode) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_OPENAT, (uintptr_t)WS_AT_FDCWD, (uintptr_t)path, (uintptr_t)flags, (uintptr_t)mode, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}

int ws_close(int fd) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_CLOSE, (uintptr_t)fd, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

intptr_t ws_read(int fd, void *buf, size_t count) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_READ, (uintptr_t)fd, (uintptr_t)buf, (uintptr_t)count, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return r;
}

intptr_t ws_write(int fd, const void *buf, size_t count) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_WRITE, (uintptr_t)fd, (uintptr_t)buf, (uintptr_t)count, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return r;
}

int64_t ws_lseek(int fd, int64_t offset, int whence) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_LSEEK, (uintptr_t)fd, (uintptr_t)offset, (uintptr_t)whence, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int64_t)r;
}

int ws_unlink(const char *path) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_UNLINKAT, (uintptr_t)WS_AT_FDCWD, (uintptr_t)path, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_rmdir(const char *path) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_UNLINKAT, (uintptr_t)WS_AT_FDCWD, (uintptr_t)path, WS_AT_REMOVEDIR, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_mkdir(const char *path, unsigned mode) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_MKDIRAT, (uintptr_t)WS_AT_FDCWD, (uintptr_t)path, (uintptr_t)mode, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_rename(const char *oldpath, const char *newpath) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_RENAMEAT, (uintptr_t)WS_AT_FDCWD, (uintptr_t)oldpath, (uintptr_t)WS_AT_FDCWD, (uintptr_t)newpath, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_access(const char *path, unsigned mode) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_FACCESSAT, (uintptr_t)WS_AT_FDCWD, (uintptr_t)path, (uintptr_t)mode, 0, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}
int ws_fchmodat(int dirfd, const char *pathname, unsigned mode, int flags) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_FCHMODAT, (uintptr_t)dirfd, (uintptr_t)pathname, (uintptr_t)mode, (uintptr_t)flags, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_chmod(const char *pathname, unsigned mode) {
    return ws_fchmodat(WS_AT_FDCWD, pathname, mode, 0);
}

int ws_fchownat(int dirfd, const char *pathname, int uid, int gid, int flags) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_FCHOWNAT, (uintptr_t)dirfd, (uintptr_t)pathname, (uintptr_t)uid, (uintptr_t)gid, (uintptr_t)flags, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_chown(const char *pathname, int uid, int gid) {
    return ws_fchownat(WS_AT_FDCWD, pathname, uid, gid, 0);
}

int ws_utimensat(int dirfd, const char *pathname, const void *times, int flags) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_UTIMENSAT, (uintptr_t)dirfd, (uintptr_t)pathname, (uintptr_t)times, (uintptr_t)flags, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_ftruncate(int fd, int64_t length) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_FTRUNCATE, (uintptr_t)fd, (uintptr_t)length, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_symlinkat(const char *target, int newdirfd, const char *linkpath) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_SYMLINKAT, (uintptr_t)target, (uintptr_t)newdirfd, (uintptr_t)linkpath, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_symlink(const char *target, const char *linkpath) {
    return ws_symlinkat(target, WS_AT_FDCWD, linkpath);
}
int ws_flock(int fd, int how) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_FLOCK, (uintptr_t)fd, (uintptr_t)how, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

intptr_t ws_copy_file_range(int fd_in, int64_t *off_in, int fd_out, int64_t *off_out, size_t len, unsigned flags) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_COPY_FILE_RANGE, (uintptr_t)fd_in, (uintptr_t)off_in, (uintptr_t)fd_out, (uintptr_t)off_out, (uintptr_t)len, (uintptr_t)flags, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return r;
}

int ws_statfs(const char *path, void *buf) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_STATFS, (uintptr_t)path, (uintptr_t)buf, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_fstatfs(int fd, void *buf) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_FSTATFS, (uintptr_t)fd, (uintptr_t)buf, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}
int ws_mkdir_all(const char *path, unsigned mode) {
    if (!path || !*path) return 0;
    char buf[1024];
    strncpy(buf, path, sizeof(buf) - 1);
    buf[sizeof(buf) - 1] = '\0';
    for (char *p = buf + 1; *p; p++) {
        if (*p == '/' || *p == '\\') {
            *p = '\0';
            ws_mkdir(buf, mode);
            *p = '/';
        }
    }
    return ws_mkdir(buf, mode);
}

int ws_remove_all(const char *path) {
    if (!path || !*path) return 0;
    if (ws_unlink(path) == 0) return 0;
    return ws_rmdir(path);
}

int ws_copy_file(const char *src, const char *dst) {
    int in = ws_open(src, WS_O_RDONLY | WS_O_CLOEXEC, 0);
    if (in < 0) return -1;
    int out = ws_open(dst, WS_O_WRONLY | WS_O_CREAT | WS_O_TRUNC | WS_O_CLOEXEC, 0644);
    if (out < 0) { ws_close(in); return -1; }
    char buf[65536];
    intptr_t n;
    while ((n = ws_read(in, buf, sizeof(buf))) > 0) {
        if (ws_write(out, buf, (size_t)n) != n) {
            ws_close(in); ws_close(out); return -1;
        }
    }
    ws_close(in);
    ws_close(out);
    return (n < 0) ? -1 : 0;
}
int ws_execve(const char *pathname, char *const argv[], char *const envp[]) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_EXECVE, (uintptr_t)pathname, (uintptr_t)argv, (uintptr_t)envp, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}

int ws_tcgetattr(int fd, void *termios_p) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_IOCTL, (uintptr_t)fd, 0x5401 /* TCGETS */, (uintptr_t)termios_p, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_tcsetattr(int fd, int optional_actions, const void *termios_p) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_IOCTL, (uintptr_t)fd, (uintptr_t)optional_actions, (uintptr_t)termios_p, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_readlink(const char *path, char *buf, size_t bufsiz) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_READLINKAT, (uintptr_t)WS_AT_FDCWD, (uintptr_t)path, (uintptr_t)buf, (uintptr_t)bufsiz, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}
int ws_pipe2(int pipefd[2], int flags) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_PIPE2, (uintptr_t)pipefd, (uintptr_t)flags, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_dup3(int oldfd, int newfd, int flags) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_DUP3, (uintptr_t)oldfd, (uintptr_t)newfd, (uintptr_t)flags, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

void *ws_mmap(void *addr, size_t length, int prot, int flags, int fd, int64_t offset) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_MMAP, (uintptr_t)addr, (uintptr_t)length, (uintptr_t)prot, (uintptr_t)flags, (uintptr_t)fd, (uintptr_t)offset, &err);
    if (r < 0) {
        errno = err;
        return (void *)-1;
    }
    return (void *)r;
}

int ws_munmap(void *addr, size_t length) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_MUNMAP, (uintptr_t)addr, (uintptr_t)length, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}
int ws_inotify_init1(int flags) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_INOTIFY_INIT1, (uintptr_t)flags, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}

int ws_inotify_add_watch(int fd, const char *pathname, uint32_t mask) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_INOTIFY_ADD_WATCH, (uintptr_t)fd, (uintptr_t)pathname, (uintptr_t)mask, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}

int ws_inotify_rm_watch(int fd, int wd) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_INOTIFY_RM_WATCH, (uintptr_t)fd, (uintptr_t)wd, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_getuid(void) {
    int err = 0;
    return (int)ws_syscall(WS_SYS_GETUID, 0, 0, 0, &err);
}

int ws_getgid(void) {
    int err = 0;
    return (int)ws_syscall(WS_SYS_GETGID, 0, 0, 0, &err);
}

int ws_geteuid(void) {
    int err = 0;
    return (int)ws_syscall(WS_SYS_GETEUID, 0, 0, 0, &err);
}

int ws_getegid(void) {
    int err = 0;
    return (int)ws_syscall(WS_SYS_GETEGID, 0, 0, 0, &err);
}

int ws_getppid(void) {
    int err = 0;
    return (int)ws_syscall(WS_SYS_GETPPID, 0, 0, 0, &err);
}
int ws_ioctl(int fd, unsigned long req, void *arg) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_IOCTL, (uintptr_t)fd, (uintptr_t)req, (uintptr_t)arg, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}

int ws_clock_gettime(int clockid, void *ts) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_CLOCK_GETTIME, (uintptr_t)clockid, (uintptr_t)ts, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_nanosleep(const void *req, void *rem) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_NANOSLEEP, (uintptr_t)req, (uintptr_t)rem, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}
int ws_clock_nanosleep(int clockid, int flags, const void *req, void *rem) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_CLOCK_NANOSLEEP, (uintptr_t)clockid, (uintptr_t)flags, (uintptr_t)req, (uintptr_t)rem, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_kill(int pid, int sig) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_KILL, (uintptr_t)pid, (uintptr_t)sig, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_wait4(int pid, int *status, int options, void *rusage) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_WAIT4, (uintptr_t)pid, (uintptr_t)status, (uintptr_t)options, (uintptr_t)rusage, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}
