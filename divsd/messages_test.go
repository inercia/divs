package divsd

import "testing"

func TestEncodeDecode(t *testing.T) {
	type ping struct {
		SeqNo int
	}

	msg := &ping{SeqNo: 100}
	buf, err := encode(MSG_LAST, msg)
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
	var out ping
	if err := decode(buf.Bytes()[1:], &out); err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
		if msg.SeqNo != out.SeqNo {
		t.Fatalf("bad sequence no")
	}
}

// Assert we can serialize/deserialize a DbReq
func TestDbReqSerialization(t *testing.T) {
	dbreq := DbReq{Name: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}
	dbReqBuffer := dbreq.ToBuffer()
	dbReqRes := NewDbReqFromBuffer(dbReqBuffer[1:])
	if dbReqRes.Name != "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		t.Errorf("received: %s", dbReqRes.Name)
	}
}

func BenchmarkDbReqSerialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dbreq := DbReq{Name: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}
		dbReqBuffer := dbreq.ToBuffer()
		NewDbReqFromBuffer(dbReqBuffer[1:])
	}
}

// Assert we can serialize/deserialize a DbVal
func TestDbValSerialization(t *testing.T) {
	dbVal := DbVal{Name: "ABCDEFGHIJKLMNOPQRSTUVWXYZ", Value: "1234567890"}
	dbValBuffer := dbVal.ToBuffer()
	dbValRes := NewDbReqFromBuffer(dbValBuffer[1:])
	if dbValRes.Name != "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		t.Errorf("received: %s", dbValRes.Name)
	}
}

func BenchmarkDbValSerialization(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dbVal := DbVal{Name: "ABCDEFGHIJKLMNOPQRSTUVWXYZ", Value: "1234567890"}
		dbValBuffer := dbVal.ToBuffer()
		NewDbReqFromBuffer(dbValBuffer[1:])
	}
}
