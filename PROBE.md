# Probing Wine Raw Syscall Execution

Before relying on `libwinescape` in production, you can verify whether your Wine setup passes raw kernel trap instructions using the included standalone probe.

## Building the Probe

From the repository root:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o probe/wine_syscall_probe.exe ./probe
```

## Running under Wine

```bash
wine probe/wine_syscall_probe.exe
```

Expected output:

```
issuing raw Linux SYSCALL instruction (getpid, number 39) from a windows/amd64 binary...
raw syscall returned: 12345
RESULT: PASS -- looks like a plausible Linux PID. Wine let the raw syscall through.
```

Compare the returned PID with the actual PID of the `wine64` process on your host (`ps aux | grep wine`). If they match, raw syscall traps are working properly on your system.
