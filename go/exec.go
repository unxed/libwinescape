package winescape

import (
	"unsafe"
)

// Execve replaces the current process with the executable at path with arguments argv and environment envp.
func Execve(path string, argv []string, envp []string) error {
	p, err := BytePtrFromString(ToUnixPath(path))
	if err != nil {
		return err
	}

	argvPtrs := make([]uintptr, len(argv)+1)
	for i, arg := range argv {
		argP, err := BytePtrFromString(arg)
		if err != nil {
			return err
		}
		argvPtrs[i] = uintptr(unsafe.Pointer(argP))
	}
	argvPtrs[len(argv)] = 0

	envpPtrs := make([]uintptr, len(envp)+1)
	for i, env := range envp {
		envP, err := BytePtrFromString(env)
		if err != nil {
			return err
		}
		envpPtrs[i] = uintptr(unsafe.Pointer(envP))
	}
	envpPtrs[len(envp)] = 0

	var argvPtr, envpPtr uintptr
	if len(argv) > 0 {
		argvPtr = uintptr(unsafe.Pointer(&argvPtrs[0]))
	}
	if len(envp) > 0 {
		envpPtr = uintptr(unsafe.Pointer(&envpPtrs[0]))
	}

	_, _, errSys := Syscall(sysExecve, uintptr(unsafe.Pointer(p)), argvPtr, envpPtr)
	return errSys
}
