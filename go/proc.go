package winescape

// Getuid returns the numeric real user ID of the caller on the host.
func Getuid() int {
	r1, _, _ := Syscall(sysGetuid, 0, 0, 0)
	return int(r1)
}

// Getgid returns the numeric real group ID of the caller on the host.
func Getgid() int {
	r1, _, _ := Syscall(sysGetgid, 0, 0, 0)
	return int(r1)
}

// Geteuid returns the numeric effective user ID of the caller on the host.
func Geteuid() int {
	r1, _, _ := Syscall(sysGeteuid, 0, 0, 0)
	return int(r1)
}

// Getegid returns the numeric effective group ID of the caller on the host.
func Getegid() int {
	r1, _, _ := Syscall(sysGetegid, 0, 0, 0)
	return int(r1)
}

// Getppid returns the host process ID of the caller's parent.
func Getppid() int {
	r1, _, _ := Syscall(sysGetppid, 0, 0, 0)
	return int(r1)
}
