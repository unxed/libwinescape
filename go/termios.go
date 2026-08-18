package winescape

import (
	"unsafe"
)

// Linux termios ioctl and flag constants.
const (
	TCGETS  = 0x5401
	TCSETS  = 0x5402
	TCSETSW = 0x5403
	TCSETSF = 0x5404

	// Input flags (c_iflag)
	IGNBRK = 0000001
	BRKINT = 0000002
	IGNPAR = 0000004
	PARMRK = 0000010
	INPCK  = 0000020
	ISTRIP = 0000040
	INLCR  = 0000100
	IGNCR  = 0000200
	ICRNL  = 0000400
	IXON   = 0002000
	IXOFF  = 0010000

	// Output flags (c_oflag)
	OPOST = 0000001

	// Control flags (c_cflag)
	CS8 = 0000060

	// Local flags (c_lflag)
	ISIG   = 0000001
	ICANON = 0000002
	ECHO   = 0000010
	ECHOE  = 0000020
	ECHOK  = 0000040
	ECHONL = 0000100
	NOFLSH = 0000200
	TOSTOP = 0000400
	IEXTEN = 0000100000

	// Control character indices (c_cc)
	VMIN  = 6
	VTIME = 5
)

// Termios matches Linux struct termios (60 bytes).
type Termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Line   uint8
	Cc     [32]uint8
	Pad0   [3]byte
	Ispeed uint32
	Ospeed uint32
}

// Tcgetattr gets the parameters associated with the terminal referenced by fd.
func Tcgetattr(fd int, termios *Termios) error {
	return Ioctl(fd, TCGETS, unsafe.Pointer(termios))
}

// Tcsetattr sets the parameters associated with the terminal referenced by fd.
func Tcsetattr(fd int, opt int, termios *Termios) error {
	return Ioctl(fd, uintptr(opt), unsafe.Pointer(termios))
}

// MakeRaw puts the terminal referenced by fd into raw mode, returning the original state.
func MakeRaw(fd int) (*Termios, error) {
	var orig Termios
	if err := Tcgetattr(fd, &orig); err != nil {
		return nil, err
	}

	raw := orig
	raw.Iflag &^= (IGNBRK | BRKINT | PARMRK | ISTRIP | INLCR | IGNCR | ICRNL | IXON)
	raw.Oflag &^= OPOST
	raw.Lflag &^= (ECHO | ECHONL | ICANON | ISIG | IEXTEN)
	raw.Cflag &^= 0000060
	raw.Cflag |= CS8
	raw.Cc[VMIN] = 1
	raw.Cc[VTIME] = 0

	if err := Tcsetattr(fd, TCSETSF, &raw); err != nil {
		return nil, err
	}
	return &orig, nil
}
