package winescape

import (
	"bytes"
	"strings"
)

// HostEnviron reads and parses the raw host environment variables from /proc/self/environ.
// It bypasses Wine's environment table and returns the exact host environment.
func HostEnviron() ([]string, error) {
	fd, err := Open("/proc/self/environ", O_RDONLY|O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer Close(fd)

	var data []byte
	buf := make([]byte, 4096)
	for {
		n, err := Read(fd, buf)
		if err != nil {
			return nil, err
		}
		if n <= 0 {
			break
		}
		data = append(data, buf[:n]...)
	}

	var envs []string
	rawEntries := bytes.Split(data, []byte{0})
	for _, entry := range rawEntries {
		if len(entry) > 0 {
			envs = append(envs, string(entry))
		}
	}
	return envs, nil
}

// HostGetenv retrieves the value of the host environment variable named by key.
func HostGetenv(key string) string {
	envs, err := HostEnviron()
	if err != nil {
		return ""
	}
	prefix := key + "="
	for _, env := range envs {
		if strings.HasPrefix(env, prefix) {
			return env[len(prefix):]
		}
	}
	return ""
}
