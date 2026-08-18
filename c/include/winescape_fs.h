#ifndef WINESCAPE_FS_H
#define WINESCAPE_FS_H

#include "winescape.h"
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define WS_O_RDONLY   0x0000
#define WS_O_WRONLY   0x0001
#define WS_O_RDWR     0x0002
#define WS_O_CREAT    0x0040
#define WS_O_EXCL     0x0080
#define WS_O_TRUNC    0x0200
#define WS_O_APPEND   0x0400
#define WS_O_NONBLOCK 0x0800
#define WS_O_DIRECTORY 0x10000
#define WS_O_CLOEXEC  0x80000

#define WS_AT_FDCWD   (-100)
#define WS_AT_SYMLINK_NOFOLLOW 0x100
#define WS_AT_REMOVEDIR 0x200

int ws_open(const char *path, int flags, unsigned mode);
int ws_close(int fd);
intptr_t ws_read(int fd, void *buf, size_t count);
intptr_t ws_write(int fd, const void *buf, size_t count);
int64_t ws_lseek(int fd, int64_t offset, int whence);
int ws_unlink(const char *path);
int ws_rmdir(const char *path);
int ws_mkdir(const char *path, unsigned mode);
int ws_rename(const char *oldpath, const char *newpath);
int ws_access(const char *path, unsigned mode);
int ws_readlink(const char *path, char *buf, size_t bufsiz);
int ws_pipe2(int pipefd[2], int flags);
int ws_dup3(int oldfd, int newfd, int flags);
void *ws_mmap(void *addr, size_t length, int prot, int flags, int fd, int64_t offset);
int ws_munmap(void *addr, size_t length);
int ws_inotify_init1(int flags);
int ws_inotify_add_watch(int fd, const char *pathname, uint32_t mask);
int ws_inotify_rm_watch(int fd, int wd);
int ws_getuid(void);
int ws_getgid(void);
int ws_geteuid(void);
int ws_getegid(void);
int ws_getppid(void);
int ws_ioctl(int fd, unsigned long req, void *arg);
int ws_clock_gettime(int clockid, void *ts);
int ws_nanosleep(const void *req, void *rem);
int ws_clock_nanosleep(int clockid, int flags, const void *req, void *rem);
int ws_kill(int pid, int sig);
int ws_wait4(int pid, int *status, int options, void *rusage);

#ifdef __cplusplus
}
#endif

#endif /* WINESCAPE_FS_H */
