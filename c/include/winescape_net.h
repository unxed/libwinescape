#ifndef WINESCAPE_NET_H
#define WINESCAPE_NET_H

#include "winescape.h"
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define WS_AF_UNIX  1
#define WS_AF_LOCAL 1
#define WS_AF_INET  2
#define WS_AF_INET6 10

#define WS_SOCK_STREAM 1
#define WS_SOCK_DGRAM  2
#define WS_SOCK_CLOEXEC 0x00080000
#define WS_SOCK_NONBLOCK 0x00000800

int ws_socket(int domain, int type, int protocol);
int ws_connect(int fd, const void *addr, size_t addrlen);
int ws_bind(int fd, const void *addr, size_t addrlen);
int ws_listen(int fd, int backlog);
int ws_accept4(int fd, void *addr, uint32_t *addrlen, int flags);
int ws_dial_unix(const char *path);

#ifdef __cplusplus
}
#endif

#endif /* WINESCAPE_NET_H */
