package divsd

import "testing"
import (
	"bytes"
	"code.google.com/p/gopacket/layers"
	"net"
)

// A general test for the encoding/decoding functions
func TestEncodeDecode(t *testing.T) {
	type ping struct {
		SeqNo int
	}

	msg := &ping{SeqNo: 100}
	buf, err := encodeMsg(MSG_LAST, msg)
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}

	typ, encodedMsg, err := getTypeAndEncodedMsg(buf.Bytes())
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
	if typ != MSG_LAST {
		t.Fatalf("unexpected message type: %s", err)
	}

	var out ping
	err = decodeMsg(encodedMsg, &out)
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
	if msg.SeqNo != out.SeqNo {
		t.Fatalf("bad sequence no")
	}
}

// Assert we can serialize/deserialize a ethernet package
func TestPkgEtherSerialization(t *testing.T) {
	pkg := EthernetPacket{
		layers.Ethernet{
			BaseLayer: layers.BaseLayer{
				Payload: []byte("0123456789"),
			},
			SrcMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			EthernetType: layers.EthernetTypeIPv4,
			Length:       10,
		},
	}

	pkgBuffer, _ := pkg.Encode()
	pkgRes := NewEthernetPacketFromBuffer(pkgBuffer[1:])
	if bytes.Compare(pkgRes.Payload, []byte("0123456789")) != 0 {
		t.Errorf("received: %s", pkgRes.Payload)
	}
}

// Benchmark for ethernet package serialization/deserializations
func BenchmarkDbReqSerialization(b *testing.B) {
	pkg := EthernetPacket{
		layers.Ethernet{
			BaseLayer: layers.BaseLayer{
				Payload: []byte("0123456789"),
			},
			SrcMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			EthernetType: layers.EthernetTypeIPv4,
			Length:       10,
		},
	}

	for i := 0; i < b.N; i++ {
		pkgBuffer, _ := pkg.Encode()
		NewEthernetPacketFromBuffer(pkgBuffer[1:])
	}
}
