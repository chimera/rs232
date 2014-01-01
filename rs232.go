// rs232.go, jpad 2013

package rs232

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Serial port options.
//
// Valid bit rates (linux):
// 200, 300, 600, 1200, 1800, 2400, 4800, 9600, 19200, 38400, 57600,
// 115200, 230400, 460800, 500000, 576000, 921600, 1000000, 1152000,
// 1500000, 2000000, 2500000, 3000000, 3500000, 4000000.
//
// Valid bit rates (OS X):
// 200, 300, 600, 1200, 1800, 2400, 4800, 7200, 9600, 14400, 19200, 28800,
// 38400, 57600, 78600, 115200, 230400
type Options struct {
	BitRate  uint32     // number of bits per second (baud)
	DataBits uint8      // number of data bits (5, 6, 7, 8)
	StopBits uint8      // number of stop bits (1, 2)
	Parity   ParityMode // none, odd, even
	Timeout  uint8      // read timeout in units of 0.1 seconds (0 = disable)
}

type ParityMode uint8

// Parity modes
const (
	PARITY_NONE = ParityMode(0)
	PARITY_ODD  = ParityMode(1)
	PARITY_EVEN = ParityMode(2)
)

//// Port //////////////////////////////////////////////////////////////////////

type Port struct {
	options Options
	file    *os.File
}

// Open will open the named serial port device with the given options.
// Open returns an *rs232.Error which implements the built-in error interface
// but additionally allows access to specific error codes. See Error.
//
// See type Options for valid parameter ranges.
func Open(name string, opt Options) (port *Port, err error) {
	// validate options
	mrate, mformat, err := validateOptions(&opt)
	if err != nil {
		return nil, err
	}

	// open special device file
	// O_NOCTTY: if the file is a tty, don't make it the controlling terminal.
	file, err := os.OpenFile(
		name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &Error{ERR_DEVICE, err.Error()}
		} else if os.IsPermission(err) {
			return nil, &Error{ERR_ACCESS, err.Error()}
		} else {
			return nil, &Error{ERR_PARAMS, err.Error()}
		}
	}
	fd := file.Fd()

	// timeout settings
	vmin := uint8(0)
	if opt.Timeout == 0 {
		vmin = 1
	}

	// termios settings
	err = setTermios(fd, mrate, mformat, vmin, opt.Timeout)
	if err != nil {
		file.Close()
		return nil, err
	}

	// set device file to blocking mode
	err = syscall.SetNonblock(int(fd), false)
	if err != nil {
		file.Close()
		return nil, &Error{ERR_PARAMS, err.Error()}
	}

	return &Port{opt, file}, nil
}

func (p *Port) String() string {
	pari := ""
	switch p.options.Parity {
	case PARITY_NONE:
		pari = "N"
	case PARITY_ODD:
		pari = "O"
	case PARITY_EVEN:
		pari = "E"
	default:
		pari = "?"
	}
	return fmt.Sprintf("'%s' %d%s%d @ %d bits/s",
		p.file.Name(),
		p.options.DataBits, pari, p.options.StopBits, p.options.BitRate)
}

func (p *Port) Close() error {
	return p.file.Close()
}

func (p *Port) Read(b []byte) (n int, err error) {
	return p.file.Read(b)
}

func (p *Port) Write(b []byte) (n int, err error) {
	return p.file.Write(b)
}

// GetRTS gets the 'Request To Send' control signal level.
func (p *Port) GetRTS() (bool, error) {
	return p.getControlSignal(syscall.TIOCM_RTS)
}

// GetCTS gets the 'Clear To Send' control signal level.
func (p *Port) GetCTS() (bool, error) {
	return p.getControlSignal(syscall.TIOCM_CTS)
}

// GetDTR gets the 'Data Terminal Ready' control signal level.
func (p *Port) GetDTR() (bool, error) {
	return p.getControlSignal(syscall.TIOCM_DTR)
}

// GetDSR gets the 'Data Set Ready' control signal level.
func (p *Port) GetDSR() (bool, error) {
	return p.getControlSignal(syscall.TIOCM_DSR)
}

// SetDTR sets the 'Data Terminal Ready' control signal level.
func (p *Port) SetDTR(level bool) error {
	return p.setControlSignal(syscall.TIOCM_DTR, level)
}

// SetRTS sets the 'Request To Send' control signal level.
func (p *Port) SetRTS(level bool) error {
	return p.setControlSignal(syscall.TIOCM_RTS, level)
}

// BytesAvailable returns the number of bytes available in the serial port
// input buffer. Read() will block as long as there are no bytes available.
func (p *Port) BytesAvailable() (int, error) {
	n := int(0)
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		p.file.Fd(),
		uintptr(_FIONREAD),
		uintptr(unsafe.Pointer(&n)))
	if errno != 0 {
		return 0, fmt.Errorf("ioctl TIOCIMQ %d", errno)
	}
	return n, nil
}

//// Error /////////////////////////////////////////////////////////////////////

// Error holds an error code and the corresponding human-readable message.
type Error struct {
	Code int
	Msg  string
}

// Error implements the built-in error interface.
func (e *Error) Error() string {
	return e.Msg
}

// rs232 error codes.
const (
	ERR_NONE   = 0 // no error
	ERR_DEVICE = 1 // device error
	ERR_ACCESS = 2 // no access permissions
	ERR_PARAMS = 3 // invalid parameters
)

//// low-level utilities ///////////////////////////////////////////////////////

func (p *Port) getControlSignal(sigmask uint) (bool, error) {
	state := uint(0)
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		p.file.Fd(),
		uintptr(syscall.TIOCMGET),
		uintptr(unsafe.Pointer(&state)))
	if errno != 0 {
		return false, fmt.Errorf("ioctl TIOCMGET %d", errno)
	}
	return ((state & sigmask) != 0), nil
}

func (p *Port) setControlSignal(sigmask uint, level bool) error {
	state := uint(0)
	// get current state
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		p.file.Fd(),
		uintptr(syscall.TIOCMGET),
		uintptr(unsafe.Pointer(&state)))
	if errno != 0 {
		return fmt.Errorf("ioctl TIOCMGET %d", errno)
	}
	// modify state
	switch level {
	case true:
		state |= sigmask // set
	case false:
		state &^= sigmask // clear
	}
	// set new state
	_, _, errno = syscall.Syscall(
		syscall.SYS_IOCTL,
		p.file.Fd(),
		uintptr(syscall.TIOCMSET),
		uintptr(unsafe.Pointer(&state)))
	if errno != 0 {
		return fmt.Errorf("ioctl TIOCMSET %d", errno)
	}
	return nil
}

func validateOptions(o *Options) (uint, uint, error) {
	mrate, err := bitrate2mask(o.BitRate)
	if err != nil {
		return 0, 0, &Error{ERR_PARAMS, err.Error()}
	}
	mdata, err := databits2mask(o.DataBits)
	if err != nil {
		return 0, 0, &Error{ERR_PARAMS, err.Error()}
	}
	mpari, err := parity2mask(o.Parity)
	if err != nil {
		return 0, 0, &Error{ERR_PARAMS, err.Error()}
	}
	mstop, err := stopbits2mask(o.StopBits)
	if err != nil {
		return 0, 0, &Error{ERR_PARAMS, err.Error()}
	}
	return mrate, (mdata | mpari | mstop), nil
}

func databits2mask(nbits uint8) (uint, error) {
	switch nbits {
	case 5:
		return syscall.CS5, nil
	case 6:
		return syscall.CS6, nil
	case 7:
		return syscall.CS7, nil
	case 8:
		return syscall.CS8, nil
	}
	return 0, fmt.Errorf("invalid DataBits (%d)", nbits)
}

func stopbits2mask(nbits uint8) (uint, error) {
	switch nbits {
	case 1:
		return 0, nil
	case 2:
		return syscall.CSTOPB, nil
	}
	return 0, fmt.Errorf("invalid StopBits (%d)", nbits)
}

func parity2mask(parity ParityMode) (uint, error) {
	switch parity {
	case PARITY_NONE:
		return 0, nil
	case PARITY_ODD:
		return (syscall.PARENB | syscall.PARODD), nil
	case PARITY_EVEN:
		return syscall.PARENB, nil
	}
	return 0, fmt.Errorf("invalid Parity (%d)", parity)
}
