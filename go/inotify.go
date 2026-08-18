package winescape

import (
	"encoding/binary"
	"unsafe"
)

// Linux inotify event mask constants.
const (
	IN_ACCESS        = 0x00000001
	IN_MODIFY        = 0x00000002
	IN_ATTRIB        = 0x00000004
	IN_CLOSE_WRITE   = 0x00000008
	IN_CLOSE_NOWRITE = 0x00000010
	IN_CLOSE         = (IN_CLOSE_WRITE | IN_CLOSE_NOWRITE)
	IN_OPEN          = 0x00000020
	IN_MOVED_FROM    = 0x00000040
	IN_MOVED_TO      = 0x00000080
	IN_MOVE          = (IN_MOVED_FROM | IN_MOVED_TO)
	IN_CREATE        = 0x00000100
	IN_DELETE        = 0x00000200
	IN_DELETE_SELF   = 0x00000400
	IN_MOVE_SELF     = 0x00000800
	IN_ALL_EVENTS    = 0x00000fff
	IN_ISDIR         = 0x40000000
	IN_NONBLOCK      = 0x00000800
	IN_CLOEXEC       = 0x00080000
)

// InotifyEvent represents a single event notification from an inotify descriptor.
type InotifyEvent struct {
	Wd     int32
	Mask   uint32
	Cookie uint32
	Name   string
}

// InotifyInit1 creates a new inotify instance with flags (e.g. IN_CLOEXEC, IN_NONBLOCK).
func InotifyInit1(flags int) (int, error) {
	r1, _, err := Syscall(sysInotifyInit1, uintptr(flags), 0, 0)
	if err != nil {
		return -1, err
	}
	return int(r1), nil
}

// InotifyInit creates a new inotify instance with default flags.
func InotifyInit() (int, error) {
	return InotifyInit1(0)
}

// InotifyAddWatch adds or modifies a watch on the file or directory at path.
func InotifyAddWatch(fd int, path string, mask uint32) (int, error) {
	p, err := BytePtrFromString(ToUnixPath(path))
	if err != nil {
		return -1, err
	}
	r1, _, errSys := Syscall(sysInotifyAddWatch, uintptr(fd), uintptr(unsafe.Pointer(p)), uintptr(mask))
	if errSys != nil {
		return -1, errSys
	}
	return int(r1), nil
}

// InotifyRmWatch removes a watch from an inotify instance.
func InotifyRmWatch(fd int, wd int) error {
	_, _, err := Syscall(sysInotifyRmWatch, uintptr(fd), uintptr(wd), 0)
	return err
}

// ParseInotifyEvents parses structured inotify events from a raw read buffer.
func ParseInotifyEvents(buf []byte) ([]InotifyEvent, error) {
	var events []InotifyEvent
	cur := 0
	for cur+16 <= len(buf) {
		wd := int32(binary.LittleEndian.Uint32(buf[cur : cur+4]))
		mask := binary.LittleEndian.Uint32(buf[cur+4 : cur+8])
		cookie := binary.LittleEndian.Uint32(buf[cur+8 : cur+12])
		length := binary.LittleEndian.Uint32(buf[cur+12 : cur+16])
		if cur+16+int(length) > len(buf) {
			break
		}
		var name string
		if length > 0 {
			nameBytes := buf[cur+16 : cur+16+int(length)]
			end := 0
			for end < len(nameBytes) && nameBytes[end] != 0 {
				end++
			}
			name = string(nameBytes[:end])
		}
		events = append(events, InotifyEvent{
			Wd:     wd,
			Mask:   mask,
			Cookie: cookie,
			Name:   name,
		})
		cur += 16 + int(length)
	}
	return events, nil
}
