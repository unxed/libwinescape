package spec

type TargetOS string
type TargetArch string

const (
	OSLinux   TargetOS = "linux"
	OSFreeBSD TargetOS = "freebsd"

	ArchAMD64   TargetArch = "amd64"
	ArchARM64   TargetArch = "arm64"
	ArchRISCV64 TargetArch = "riscv64"
)

type SyscallEntry struct {
	Name    string
	Args    int
	Numbers map[TargetOS]map[TargetArch]uint64
	Note    string
}

// SyscallTable is the single source of truth for raw syscall numbers.
// Numbers are verified against kernel headers (Linux asm/unistd_64.h,
// asm-generic/unistd.h, FreeBSD sys/syscall.h).
var SyscallTable = []SyscallEntry{
	{
		Name: "read",
		Args: 3,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 0, ArchARM64: 63},
			OSFreeBSD: {ArchAMD64: 3},
		},
	},
	{
		Name: "write",
		Args: 3,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 1, ArchARM64: 64},
			OSFreeBSD: {ArchAMD64: 4},
		},
	},
	{
		Name: "open",
		Args: 3,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 2},
			OSFreeBSD: {ArchAMD64: 5},
		},
		Note: "Linux arm64 uses openat (56)",
	},
	{
		Name: "close",
		Args: 1,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 3, ArchARM64: 57},
			OSFreeBSD: {ArchAMD64: 6},
		},
	},
	{
		Name: "fstat",
		Args: 2,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 5, ArchARM64: 80},
			OSFreeBSD: {ArchAMD64: 551},
		},
	},
	{
		Name: "lseek",
		Args: 3,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 8, ArchARM64: 62},
			OSFreeBSD: {ArchAMD64: 478},
		},
	},
	{
		Name: "pread64",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 17, ArchARM64: 67},
			OSFreeBSD: {ArchAMD64: 475},
		},
	},
	{
		Name: "pwrite64",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 18, ArchARM64: 68},
			OSFreeBSD: {ArchAMD64: 476},
		},
	},
	{
		Name: "access",
		Args: 2,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 21},
			OSFreeBSD: {ArchAMD64: 33},
		},
		Note: "Linux arm64 uses faccessat (48)",
	},
	{
		Name: "getpid",
		Args: 0,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 39, ArchARM64: 172},
			OSFreeBSD: {ArchAMD64: 20},
		},
	},
	{
		Name: "mkdir",
		Args: 2,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 83},
			OSFreeBSD: {ArchAMD64: 136},
		},
		Note: "Linux arm64 uses mkdirat (34)",
	},
	{
		Name: "rmdir",
		Args: 1,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 84},
			OSFreeBSD: {ArchAMD64: 137},
		},
		Note: "Linux arm64 uses unlinkat with AT_REMOVEDIR (35)",
	},
	{
		Name: "rename",
		Args: 2,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 82},
			OSFreeBSD: {ArchAMD64: 128},
		},
		Note: "Linux arm64 uses renameat (38)",
	},
	{
		Name: "unlink",
		Args: 1,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 87},
			OSFreeBSD: {ArchAMD64: 10},
		},
		Note: "Linux arm64 uses unlinkat (35)",
	},
	{
		Name: "symlink",
		Args: 2,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 88},
			OSFreeBSD: {ArchAMD64: 57},
		},
		Note: "Linux arm64 uses symlinkat (36)",
	},
	{
		Name: "readlink",
		Args: 3,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 89},
			OSFreeBSD: {ArchAMD64: 58},
		},
		Note: "Linux arm64 uses readlinkat (78)",
	},
	{
		Name: "getdents64",
		Args: 3,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux: {ArchAMD64: 217, ArchARM64: 61},
		},
		Note: "FreeBSD uses getdirentries (554)",
	},
	{
		Name: "getdirentries",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSFreeBSD: {ArchAMD64: 554},
		},
	},
	{
		Name: "openat",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 257, ArchARM64: 56},
			OSFreeBSD: {ArchAMD64: 499},
		},
	},
	{
		Name: "mkdirat",
		Args: 3,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 258, ArchARM64: 34},
			OSFreeBSD: {ArchAMD64: 500},
		},
	},
	{
		Name: "newfstatat",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux: {ArchAMD64: 262, ArchARM64: 79},
		},
		Note: "FreeBSD uses fstatat (552)",
	},
	{
		Name: "fstatat",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSFreeBSD: {ArchAMD64: 552},
		},
	},
	{
		Name: "unlinkat",
		Args: 3,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 263, ArchARM64: 35},
			OSFreeBSD: {ArchAMD64: 503},
		},
	},
	{
		Name: "renameat",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 264, ArchARM64: 38},
			OSFreeBSD: {ArchAMD64: 501},
		},
	},
	{
		Name: "readlinkat",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 267, ArchARM64: 78},
			OSFreeBSD: {ArchAMD64: 505},
		},
	},
	{
		Name: "faccessat",
		Args: 4,
		Numbers: map[TargetOS]map[TargetArch]uint64{
			OSLinux:   {ArchAMD64: 269, ArchARM64: 48},
			OSFreeBSD: {ArchAMD64: 489},
		},
	},
}
