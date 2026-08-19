# f4 requirement check: WINE.md Part E — 2026-08-19

Direct answer to "does the library do everything f4 needs", checked against
[`f4`'s `WINE.md`](https://github.com/unxed/f4/blob/main/WINE.md) §13.5 (the
list of things that must switch to a POSIX personality for the mode to be
real) and §13.9's Stage E5 (the f4-side work not yet done), not assumed from
memory.

## What §13.5 actually needs, checked one item at a time

| §13.5 item | Needs from `libwinescape` | Status |
|---|---|---|
| Hidden-file semantics (dot-prefix vs `FILE_ATTRIBUTE_HIDDEN`) | Nothing new — pure string check on a name f4 already has | Already covered |
| Drive list / `Alt+F1` (`vfs/disks_windows.go`) | `/sys/class/block` enumeration — plain `ReadDir`/`Open`/`Read` on text files, exactly what `vfs/disks_unix.go` already does today on the real POSIX build | Already covered |
| File times, physical size (`os_vfs_platform_windows.go`, `os_vfs_physical_windows.go`) | `Stat_t.Atim/Mtim/Ctim`, `Stat_t.Blocks` | Already covered — **but see the type-shape gotcha below, this is an f4-side integration detail, not a missing capability** |
| `rename_noreplace_windows.go` | Atomic no-clobber rename (`renameat2`/`RENAME_NOREPLACE`) | **Was missing. Added in this session** (`Renameat2`, `RenameNoReplace`) — see below |
| `trash_windows.go` (XDG trash in posix mode) | `MkdirAll`, `Rename`, file write for `.trashinfo` | Already covered (all via existing wrappers; f4 already has a POSIX implementation of this exact spec in `vfs/trash_freedesktop.go`, gated `!windows` today — Stage E5's job is making it reachable under `GOOS=windows` posix mode, not writing new library capability) |
| `UnixMode`/`0111` executable bit | `Stat_t.Permissions()` | Already covered |
| Reparse points / junctions | N/A in posix mode by design (§13.5: "в POSIX-режиме их не существует, только симлинки") | N/A |
| `$HOME`/`$XDG_CONFIG_HOME` | Nothing — plain `os.Getenv`, Wine already passes the host environment through (confirmed in an earlier session's `debug.log`) | Already covered, no library involvement needed |
| Case-sensitive sort/mask matching | Nothing — pure Go string comparison | Already covered |
| Shell/process launch (`/bin/sh` instead of `cmd.exe`, §5 Stage B5) | `Execve`, `Pipe2`, `Ioctl`/`TIOCGWINSZ`, `Tcgetattr`/`Tcsetattr`/`MakeRaw` | Already covered |

## The one real gap: found and closed

`vfs/rename_noreplace_linux.go` (the file this behavior actually needs to
match, since that's the real-POSIX-build behavior posix mode is supposed to
be indistinguishable from) uses `unix.Renameat2` with `RENAME_NOREPLACE` —
atomic "fail with `EEXIST` if the destination exists" semantics that a
userspace check-then-rename can't provide race-free. `libwinescape` didn't
have this. Added in this session:

- `spec/table.go`: `renameat2`, Linux only (amd64 `316`, arm64 `276`,
  verified against Linux kernel uapi headers, not assumed from `renameat`'s
  neighboring number — FreeBSD has no equivalent syscall to give a number
  for, so it's genuinely absent there rather than guessed).
- `go/fs.go`: `Renameat2` (returns `syscall.ENOSYS` cleanly on any target
  where the syscall number is unset, rather than letting a zero number
  silently dispatch the wrong syscall) and `RenameNoReplace` as the direct,
  drop-in equivalent of what `vfs/rename_noreplace_linux.go` already does.

## Not a missing capability, but an integration detail f4 will hit in Stage E5

`vfs/os_vfs_physical_unix.go`'s `fillPhysicalSizeCheap` type-asserts
`info.Sys().(*syscall.Stat_t)` — the standard library's own stat type. In
posix mode, `hostfs`'s `wineFileInfo.Sys()` returns `*winescape.Stat_t`
instead (a different named type, even though the data — including `.Blocks`
— is the same). The assertion will silently fail (take the "no physical size
available" branch) rather than error, so this won't crash Stage E5, but it
will quietly under-report physical size for every posix-mode file until
`fillPhysicalSizeCheap` also handles `*winescape.Stat_t`. Recorded here so
whoever does Stage E5 in f4 isn't surprised by it; the fix belongs in `f4`
(`vfs/os_vfs_physical_windows.go`'s posix branch), not in this library —
`winescape.Stat_t` already carries everything needed.

## Conclusion

Yes, with one real gap that's now closed. Before this session:
`renameat2`/`RENAME_NOREPLACE` was the only concrete missing primitive
against everything §13.5 actually asks for; everything else was already
implemented, and Stage E5 remaining work is f4-side wiring (making
already-`!windows`-gated files reachable under a `posixMode` runtime branch),
not new library capability. The `Stat_t`/`syscall.Stat_t` type-shape note
above is worth having on hand when that wiring happens, but it's not a gap in
this library either.
