#ifndef WINESCAPE_H
#define WINESCAPE_H

#include <stdint.h>
#include <stddef.h>
#include "winescape_numbers.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Generic 6-argument raw syscall trampoline.
 * Returns raw kernel return value (>=0 on success, -1 on error); sets *err_out to positive errno if non-NULL. */
intptr_t ws_syscall6(uintptr_t nr, uintptr_t a1, uintptr_t a2,
                     uintptr_t a3, uintptr_t a4, uintptr_t a5,
                     uintptr_t a6, int *err_out);

static inline intptr_t ws_syscall(uintptr_t nr, uintptr_t a1, uintptr_t a2,
                                  uintptr_t a3, int *err_out) {
    return ws_syscall6(nr, a1, a2, a3, 0, 0, 0, err_out);
}

#ifdef __cplusplus
}
#endif

#endif /* WINESCAPE_H */
