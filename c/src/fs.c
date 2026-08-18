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
