# libwinescape

Raw POSIX syscall trampolines for Windows binaries running under Wine.

## Overview

`libwinescape` is a lightweight, zero-dependency library (in pure Go assembly and C) that allows Windows-compiled binaries (`GOOS=windows` or MinGW/MSVC) running under Wine to bypass Win32/NTDLL emulation overhead and issue POSIX system calls directly to the host operating system kernel.

Unlike approaches that try to dynamically link host ELF `.so` libraries (which is fundamentally impossible from pure PE binaries under modern Wine without cgo or custom hypervisors), `libwinescape` invokes the kernel's raw trap instruction (`SYSCALL` on x86-64, `SVC #0` on ARM64) with native host register calling conventions.

## Platform Support Matrix

| Host OS under Wine | Direct Raw Syscalls Supported | Status |
|---|---|---|
| **Linux (amd64, arm64)** | Yes (native kernel ABI) | Fully supported |
| **FreeBSD (amd64)** | Yes (carry-flag error return ABI) | Supported |
| **NetBSD, DragonFly** | Yes (presumed BSD ABI) | Pending live hardware verification |
| **OpenBSD** | **No** (`pledge`/`msyscall` requires syscalls to originate within `libc.so`) | Out of scope |
| **macOS (Darwin)** | **No** (unstable syscall numbers; `libSystem` required) | Out of scope |
| **Illumos / Solaris** | **No** (unstable kernel syscall ABI; `libc` required) | Out of scope |

> **Warning:** This library is specifically designed for code compiled for Windows that executes in a Wine environment. On native Windows, calls will fail safely without crashing.

## Architecture

1. **Generic Trampoline Layer:** A minimal assembly dispatcher per (OS, architecture, language) that maps up to 6 arguments into host registers and captures error status.
2. **Syscall Table (Single Source of Truth):** Defined in `spec/table.go`, verified against upstream OS kernel headers. Code generator in `cmd/gen-numbers` generates Go constants and C `#define` headers.
3. **POSIX Wrappers:** Typed wrappers for standard filesystem operations (`open`, `read`, `write`, `close`, `fstat`, `lseek`, `getdents64`, `unlink`, `rename`, `mkdir`, `readlink`, `access`, etc.).
4. **Optional Runtime Scheduler Integration (`go/gort`):** Dedicated `LockOSThread` worker pool to prevent Go runtime scheduler starvation during blocking syscalls.
