package server

// #cgo CFLAGS: -I../tuntap
// #cgo LDFLAGS: -v ../tuntap/lib/libtuntap.a
// #include "tuntap.h"
import "C"

type TunTap struct {
	name string
}
