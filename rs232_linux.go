// rs232_linux.go, jpad 2013
// linux specific implementation
// see http://ulisse.elettra.trieste.it/services/doc/serial/config.html

package rs232

import (
	"fmt"
	"syscall"
	"unsafe"
)

const _FIONREAD = syscall.TIOCINQ

func setTermios(fd uintptr, mrate, mfmt uint, vmin, vtime uint8) error {
	// terminal settings -> raw mode
	termios := syscall.Termios{
		Iflag: syscall.IGNPAR, // ignore framing and parity errors
		Cflag: uint32(mrate) | uint32(mfmt) | syscall.CREAD | syscall.CLOCAL,
		Cc: [32]uint8{
			syscall.VMIN:  vmin,  // minimum n bytes per transfer
			syscall.VTIME: vtime, // read timeout
		},
		Ispeed: uint32(mrate),
		Ospeed: uint32(mrate),
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TCSETS), // set immediately
		uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		return &Error{ERR_PARAMS, fmt.Sprintf("ioctl TCSETS %d", errno)}
	}
	return nil
}

func bitrate2mask(rate uint32) (uint, error) {
	switch rate {
	case 200:
		return syscall.B200, nil
	case 300:
		return syscall.B300, nil
	case 600:
		return syscall.B600, nil
	case 1200:
		return syscall.B1200, nil
	case 1800:
		return syscall.B1800, nil
	case 2400:
		return syscall.B2400, nil
	case 4800:
		return syscall.B4800, nil
	case 9600:
		return syscall.B9600, nil
	case 19200:
		return syscall.B19200, nil
	case 38400:
		return syscall.B38400, nil
	case 57600:
		return syscall.B57600, nil
	case 115200:
		return syscall.B115200, nil
	case 230400:
		return syscall.B230400, nil
	case 460800:
		return syscall.B460800, nil
	case 500000:
		return syscall.B500000, nil
	case 576000:
		return syscall.B576000, nil
	case 921600:
		return syscall.B921600, nil
	case 1000000:
		return syscall.B1000000, nil
	case 1152000:
		return syscall.B1152000, nil
	case 1500000:
		return syscall.B1500000, nil
	case 2000000:
		return syscall.B2000000, nil
	case 2500000:
		return syscall.B2500000, nil
	case 3000000:
		return syscall.B3000000, nil
	case 3500000:
		return syscall.B3500000, nil
	case 4000000:
		return syscall.B4000000, nil
	}
	return 0, fmt.Errorf("invalid BitRate (%d)", rate)
}
