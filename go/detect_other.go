//go:build !windows

package winescape

// IsWine returns false on non-Windows platforms.
func IsWine() bool { return false }

// HostOS returns empty string on non-Windows platforms.
func HostOS() string { return "" }

// Available returns false on non-Windows platforms.
func Available() bool { return false }
