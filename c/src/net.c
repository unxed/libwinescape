#include "winescape_net.h"
#include "winescape_fs.h"
#include <errno.h>
#include <string.h>

int ws_socket(int domain, int type, int protocol) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_SOCKET, (uintptr_t)domain, (uintptr_t)type, (uintptr_t)protocol, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}

int ws_connect(int fd, const void *addr, size_t addrlen) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_CONNECT, (uintptr_t)fd, (uintptr_t)addr, (uintptr_t)addrlen, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_bind(int fd, const void *addr, size_t addrlen) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_BIND, (uintptr_t)fd, (uintptr_t)addr, (uintptr_t)addrlen, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_listen(int fd, int backlog) {
    int err = 0;
    intptr_t r = ws_syscall(WS_SYS_LISTEN, (uintptr_t)fd, (uintptr_t)backlog, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return 0;
}

int ws_accept4(int fd, void *addr, uint32_t *addrlen, int flags) {
    int err = 0;
    intptr_t r = ws_syscall6(WS_SYS_ACCEPT4, (uintptr_t)fd, (uintptr_t)addr, (uintptr_t)addrlen, (uintptr_t)flags, 0, 0, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return (int)r;
}

int ws_dial_unix(const char *path) {
    if (!path || strlen(path) >= 108) {
        errno = EINVAL;
        return -1;
    }
    int fd = ws_socket(WS_AF_UNIX, WS_SOCK_STREAM | WS_SOCK_CLOEXEC, 0);
    if (fd < 0) {
        return -1;
    }
    struct {
        uint16_t sun_family;
        char sun_path[108];
    } sa;
    memset(&sa, 0, sizeof(sa));
    sa.sun_family = WS_AF_UNIX;
    strncpy(sa.sun_path, path, sizeof(sa.sun_path) - 1);
    size_t sa_len = 2 + strlen(path) + 1;

    if (ws_connect(fd, &sa, sa_len) < 0) {
        int save_err = errno;
        ws_close(fd);
        errno = save_err;
        return -1;
    }
    return fd;
}
