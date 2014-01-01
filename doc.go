// doc.go, jpad 2013

/*
Package rs232 offers RS-232 serial port communication functionality.
The serial port is set to raw mode, no flow control. Read timeouts are
supported, as well as full access to the RTS/CTS DTR/DSR control signals.

rs232 implements the io.ReadWriteCloser interface. Read() and Write() are
blocking operations that can be performed simultaneously from different
goroutines.

Control signals:
	- RTS (Request To Send)     output : data is ready to be sent.
	- CTS (Clear To Send)       input  : ready to receive data.

	- DTR (Data Terminal Ready) output : ready for communication.
	- DSR (Data Set Ready)      input  : ready to receive data.

Supported OS: Linux, OS X. This package does not depend on CGO.

TODO: setting the timeout at any time?
*/
package rs232
