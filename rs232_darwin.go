// rs232_darwin.go, jpad 2013
// OS X specific implementation

package rs232

import (
	"fmt"
	"syscall"
	"unsafe"
)

const _FIONREAD = 0x4004667F

// termios bit rates ('direct' values on OS X)
// Note the rather limited max speed compared to linux...
const (
	_B200    = 200
	_B300    = 300
	_B600    = 600
	_B1200   = 1200
	_B1800   = 1800
	_B2400   = 2400
	_B4800   = 4800
	_B7200   = 7200
	_B9600   = 9600
	_B14400  = 14400
	_B19200  = 19200
	_B28800  = 28800
	_B38400  = 38400
	_B57600  = 57600
	_B78600  = 78600
	_B115200 = 115200
	_B230400 = 230400
)

const _NCC = 20 // numver of termios control codes

type termios struct {
	iflag  uint64
	oflag  uint64
	cflag  uint64
	lflag  uint64
	cc     [_NCC]uint8
	ispeed uint64
	ospeed uint64
}

func setTermios(fd uintptr, mrate, mfmt uint, vmin, vtime uint8) error {
	// terminal settings -> raw mode
	termios := termios{
		iflag: syscall.IGNPAR, // ignore framing and parity errors
		cflag: uint64(mfmt) | syscall.CREAD | syscall.CLOCAL,
		cc: [_NCC]uint8{
			syscall.VMIN:  vmin,  // minimum n bytes per transfer
			syscall.VTIME: vtime, // read timeout
		},
		ispeed: uint64(mrate),
		ospeed: uint64(mrate),
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TIOCSETA), // set immediately
		uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		return &Error{ERR_PARAMS, fmt.Sprintf("ioctl TIOCSETA %d", errno)}
	}
	return nil
}

func bitrate2mask(rate uint32) (uint, error) {
	switch rate {
	case 200:
		return _B200, nil
	case 300:
		return _B300, nil
	case 600:
		return _B600, nil
	case 1200:
		return _B1200, nil
	case 1800:
		return _B1800, nil
	case 2400:
		return _B2400, nil
	case 4800:
		return _B4800, nil
	case 7200:
		return _B7200, nil
	case 9600:
		return _B9600, nil
	case 14400:
		return _B14400, nil
	case 19200:
		return _B19200, nil
	case 28800:
		return _B28800, nil
	case 38400:
		return _B38400, nil
	case 57600:
		return _B57600, nil
	case 78600:
		return _B78600, nil
	case 115200:
		return _B115200, nil
	case 230400:
		return _B230400, nil
	}
	return 0, fmt.Errorf("invalid BitRate (%d)", rate)
}
