package server

// #include "tuntap.h"
import "C"

type TunTap struct {
	device *C.struct_device
	name   string
}

// Creates a new server.
func NewTun(name string) *TunTap {
	t := &TunTap{
		device: C.tuntap_init(),
		name:   name,
	}

	C.tuntap_start(t.device, C.TUNTAP_MODE_TUNNEL, C.TUNTAP_ID_ANY)
	return t
}

// see http://golang.org/pkg/io/#Reader
func (t *TunTap) Read(p []byte) (n int, err error) {
	// TODO
	return 0, nil
}

// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
// see http://golang.org/pkg/io/#Writer
func (t *TunTap) Write(p []byte) (n int, err error) {
	// TODO
	return 0, nil
}

// Returns the device name
func (t *TunTap) Ifname() string {
	return C.GoString(C.tuntap_get_ifname(t.device))
}

func (t *TunTap) Close() error {
	C.tuntap_destroy(t.device)
	return nil
}
