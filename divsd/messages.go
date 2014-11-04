package divsd

import (
	"bytes"
	"code.google.com/p/gopacket/layers"
	"github.com/ugorji/go/codec"
)

// messageType is an integer ID of a type of message that can be received
// on network channels from other members.
type messageType uint8

// The list of available message types.
const (
	MSG_DIVS_PKG_ETH messageType = iota
	MSG_LAST
)

type Encodeable interface {
	Encode() (data []byte, err error)
}

/////////////////////////////////////////////////////////////////////////

// An encapsulated ethernet packet
type EthernetPacket struct {
	layers.Ethernet
}

// Decode a ethernet packet from a buffer
func NewEthernetPacketFromBuffer(in []byte) *EthernetPacket {
	var pkt EthernetPacket
	if err := decodeMsg(in, &pkt); err != nil {
		log.Fatalf("unexpected err %s", err)
	}
	return &pkt
}

// Encode a ethernet packet to a ready-to-send buffer
func (pkt EthernetPacket) Encode() (data []byte, err error) {
	buf, err := encodeMsg(MSG_DIVS_PKG_ETH, pkt)
	if err != nil {
		log.Fatalf("unexpected err %s", err)
	}
	return buf.Bytes(), nil
}

/////////////////////////////////////////////////////////////////////////

// Encode writes an encoded object to a new bytes buffer
func encodeMsg(msgType messageType, in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(uint8(msgType))
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(buf, &hd)
	err := enc.Encode(in)
	return buf, err
}

// Decode reverses the encode operation on a byte slice input
// Note that encode() prepends one byte to the message for identifying
// the message type, you you should not provide the `res []bytes` returned
// by `encode()` but res[1:] instead...
func decodeMsg(buf []byte, out interface{}) error {
	r := bytes.NewReader(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}

// Peek the first bytes two bytes of a buffer for identifying the
// message type
func peekMsgType(buf []byte) (messageType, error) {
	var msgType messageType = messageType(buf[0])
	return msgType, nil
}

// Get the message type and a buffer with the encoded message
// You can then apply `decode()` in the buffer result
func getTypeAndEncodedMsg(buf []byte) (messageType, []byte, error) {
	msgType, err := peekMsgType(buf)
	return msgType, buf[1:], err
}
