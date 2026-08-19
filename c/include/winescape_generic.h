#ifndef WINESCAPE_GENERIC_H
#define WINESCAPE_GENERIC_H

#include "winescape.h"
#include <errno.h>

#ifdef __cplusplus
extern "C" {
#endif

/* WS_CALL: Generic macro to invoke any raw syscall by number with 0 to 6 arguments,
 * automatically setting errno and returning (intptr_t)-1 on error. */
#define _WS_ARG7(a1, a2, a3, a4, a5, a6, a7, ...) a7
#define _WS_SYSCALL6(nr, a1, a2, a3, a4, a5, a6) \
    ({ int _err = 0; \
       intptr_t _r = ws_syscall6((uintptr_t)(nr), (uintptr_t)(a1), (uintptr_t)(a2), \
                                 (uintptr_t)(a3), (uintptr_t)(a4), (uintptr_t)(a5), \
                                 (uintptr_t)(a6), &_err); \
       if (_r < 0) errno = _err; \
       _r; })

#define _WS_NARG(...) _WS_ARG7(__VA_ARGS__, 6, 5, 4, 3, 2, 1, 0)

static inline intptr_t _ws_call_generic(uintptr_t nr, uintptr_t a1, uintptr_t a2,
                                        uintptr_t a3, uintptr_t a4, uintptr_t a5,
                                        uintptr_t a6) {
    int err = 0;
    intptr_t r = ws_syscall6(nr, a1, a2, a3, a4, a5, a6, &err);
    if (r < 0) {
        errno = err;
        return -1;
    }
    return r;
}

#define WS_CALL0(nr)                         _ws_call_generic((uintptr_t)(nr), 0, 0, 0, 0, 0, 0)
#define WS_CALL1(nr, a1)                     _ws_call_generic((uintptr_t)(nr), (uintptr_t)(a1), 0, 0, 0, 0, 0)
#define WS_CALL2(nr, a1, a2)                 _ws_call_generic((uintptr_t)(nr), (uintptr_t)(a1), (uintptr_t)(a2), 0, 0, 0, 0)
#define WS_CALL3(nr, a1, a2, a3)             _ws_call_generic((uintptr_t)(nr), (uintptr_t)(a1), (uintptr_t)(a2), (uintptr_t)(a3), 0, 0, 0)
#define WS_CALL4(nr, a1, a2, a3, a4)         _ws_call_generic((uintptr_t)(nr), (uintptr_t)(a1), (uintptr_t)(a2), (uintptr_t)(a3), (uintptr_t)(a4), 0, 0)
#define WS_CALL5(nr, a1, a2, a3, a4, a5)     _ws_call_generic((uintptr_t)(nr), (uintptr_t)(a1), (uintptr_t)(a2), (uintptr_t)(a3), (uintptr_t)(a4), (uintptr_t)(a5), 0)
#define WS_CALL6(nr, a1, a2, a3, a4, a5, a6) _ws_call_generic((uintptr_t)(nr), (uintptr_t)(a1), (uintptr_t)(a2), (uintptr_t)(a3), (uintptr_t)(a4), (uintptr_t)(a5), (uintptr_t)(a6))

#ifdef __cplusplus
}
#endif

#endif /* WINESCAPE_GENERIC_H */
