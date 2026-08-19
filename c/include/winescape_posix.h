#ifndef WINESCAPE_POSIX_H
#define WINESCAPE_POSIX_H

/* Drop-in POSIX compatibility header.
 * Including this header remaps standard POSIX functions to winescape's
 * high-performance direct kernel wrappers under Wine. */

#include "winescape_fs.h"
#include "winescape_net.h"
#include "winescape_generic.h"

#ifdef WINESCAPE_DROP_IN_POSIX

#define open(p, f, ...)        ws_open(p, f, ##__VA_ARGS__)
#define close(fd)              ws_close(fd)
#define read(fd, b, c)         ws_read(fd, b, c)
#define write(fd, b, c)        ws_write(fd, b, c)
#define lseek(fd, off, w)      ws_lseek(fd, off, w)
#define unlink(p)              ws_unlink(p)
#define rmdir(p)               ws_rmdir(p)
#define mkdir(p, m)            ws_mkdir(p, m)
#define rename(o, n)           ws_rename(o, n)
#define chmod(p, m)            ws_chmod(p, m)
#define chown(p, u, g)         ws_chown(p, u, g)
#define access(p, m)           ws_access(p, m)
#define readlink(p, b, s)      ws_readlink(p, b, s)
#define symlink(t, l)          ws_symlink(t, l)
#define ftruncate(fd, l)       ws_ftruncate(fd, l)
#define flock(fd, h)           ws_flock(fd, h)
#define getuid()               ws_getuid()
#define getgid()               ws_getgid()
#define geteuid()              ws_geteuid()
#define getegid()              ws_getegid()
#define getppid()              ws_getppid()
#define pipe2(p, f)            ws_pipe2(p, f)
#define dup3(o, n, f)          ws_dup3(o, n, f)
#define mmap(a, l, p, f, d, o) ws_mmap(a, l, p, f, d, o)
#define munmap(a, l)           ws_munmap(a, l)

#endif /* WINESCAPE_DROP_IN_POSIX */

#endif /* WINESCAPE_POSIX_H */
