package divsd

import (
	"bytes"
	"github.com/ugorji/go/codec"
)

// messageType is an integer ID of a type of message that can be received
// on network channels from other members.
type messageType uint8

// The list of available message types.
const (
	MSG_DB_REQ messageType = iota
	MSG_DB_VAL
	MSG_LAST
)

/////////////////////////////////////////////////////////////////////////

// A database request message
type DbReq struct {
	Name string
}

func NewDbReqFromBuffer(in []byte) *DbReq {
	var dbReq DbReq
	if err := decode(in, &dbReq); err != nil {
		log.Fatalf("unexpected err %s", err)
	}
	return &dbReq
}

func (dbReq *DbReq) ToBuffer() []byte {
	buf, err := encode(MSG_DB_REQ, dbReq)
	if err != nil {
		log.Fatalf("unexpected err %s", err)
	}
	return buf.Bytes()
}

/////////////////////////////////////////////////////////////////////////

// A database value message
type DbVal struct {
	Name  string
	Value string
}

func NewDbValFromBuffer(in []byte) *DbVal {
	var dbVal DbVal
	if err := decode(in, &dbVal); err != nil {
		log.Fatalf("unexpected err %s", err)
	}
	return &dbVal
}

func (dbVal *DbVal) ToBuffer() []byte {
	buf, err := encode(MSG_DB_VAL, dbVal)
	if err != nil {
		log.Fatalf("unexpected err %s", err)
	}
	return buf.Bytes()
}

/////////////////////////////////////////////////////////////////////////

// Encode writes an encoded object to a new bytes buffer
func encode(msgType messageType, in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(uint8(msgType))
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(buf, &hd)
	err := enc.Encode(in)
	return buf, err
}

// Decode reverses the encode operation on a byte slice input
func decode(buf []byte, out interface{}) error {
	r := bytes.NewReader(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}
