# Generic Syscall Invocation & Generalization Strategies

Writing manual wrappers for hundreds of POSIX functions is neither necessary nor desirable. `libwinescape` supports several levels of generalization:

## 1. Generic Syscall Dispatcher (`ws_syscall6` / `WS_CALL*`)

Any Linux/BSD syscall can be issued directly without writing a wrapper function:

```c
#include "winescape_generic.h"

// Invoke any syscall by number with automatic errno translation:
intptr_t pid = WS_CALL0(WS_SYS_GETPID);
intptr_t res = WS_CALL3(WS_SYS_READ, fd, buffer, count);
```

In Go:
```go
// Direct 0-6 argument variadic dispatch (like windows.SyscallN in stdlib):
r1, r2, err := winescape.SyscallN(sysNumber, arg1, arg2, arg3)
res, err := winescape.Call(sysNumber, arg1, arg2)
```

## 4. ABI Considerations (ARM64 / Apple Silicon / Linux)

- **C Macro Variadics vs C `va_list` ABI:** `WS_CALL*` macros expand at compile-time into fixed-argument inline functions (`_ws_call_generic(nr, a1..a6)`). They do **not** use C runtime `va_list` / `...` functions, completely avoiding architecture-specific variadic stack-passing conventions (such as Apple's ARM64 variadic ABI).
- **Target Scope:** Raw syscall traps from PE binaries are only valid where the host OS supports a stable raw trap ABI (Linux on x86-64/ARM64 and FreeBSD). macOS/Darwin requires dynamic linking to `libSystem.dylib` and is out of scope.

## 2. Table-Driven Code Generation (`spec/table.go` + `gen-numbers`)

Syscall signatures are declarative metadata in `spec/table.go`. The code generator automatically generates:
- Multi-arch Go syscall number constants.
- Multi-arch C `#define` macros.
- Typed wrapper stubs for common signature patterns:
  - `Path -> int` (`mkdir`, `rmdir`, `unlink`)
  - `Path, Path -> int` (`rename`, `symlink`)
  - `Fd, Buffer, Size -> intptr_t` (`read`, `write`, `getdents64`)

## 3. Drop-in POSIX Compatibility Header (`winescape_posix.h`)

For existing C/C++ code (such as Far Manager plugins or utilities), defining `WINESCAPE_DROP_IN_POSIX` seamlessly redirects standard POSIX libc calls (`open`, `read`, `write`, `chmod`, `stat`, etc.) directly to kernel trampolines without rewriting existing source code.
