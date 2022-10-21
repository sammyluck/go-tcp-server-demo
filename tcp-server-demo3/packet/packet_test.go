package packet

import (
	"bytes"
	"fmt"
	"github.com/lucasepe/codename"
	"testing"
)

func TestSubmit_Encode(t *testing.T) {
	// 生成 payload
	rng, err := codename.DefaultRNG()
	if err != nil {
		t.Errorf("want nil,actual %s", err.Error())
	}
	id := fmt.Sprintf("%08d", 10) // 8 byte string
	payload := codename.Generate(rng, 4)
	s := &Submit{
		ID:      id,
		Payload: []byte(payload),
	}
	framePayload, err := Encode(s)
	if err != nil {
		t.Errorf("want nil,actual %s", err.Error())
	}

	// 验证 Encode 正确性
	commandID := int(framePayload[0])
	pktBody := framePayload[1:]
	ss := Submit{}
	ss.ID = string(pktBody[:8])
	ss.Payload = pktBody[8:]

	if commandID != CommandSubmit {
		t.Errorf("want %d,actual %d", CommandSubmit, commandID)
	}

	if ss.ID != id {
		t.Errorf("want %s,actual %s", id, ss.ID)
	}

	if string(ss.Payload) != payload {
		t.Errorf("want %s,actual %s", payload, string(ss.Payload))
	}
}

func TestSubmit_Decode(t *testing.T) {
	rng, err := codename.DefaultRNG()
	if err != nil {
		t.Errorf("want nil,actual %s", err.Error())
	}
	id := fmt.Sprintf("%08d", 10) // 8 byte string
	payload := codename.Generate(rng, 4)
	pktBody := bytes.Join([][]byte{[]byte(id[:8]), []byte(payload)}, nil)
	pkt := bytes.Join([][]byte{[]byte{CommandSubmit}, pktBody}, nil)
	p, err := Decode(pkt)
	if err != nil {
		t.Errorf("want nil,actual %s", err.Error())
	}
	ps := p.(*Submit)
	if ps.ID != id {
		t.Errorf("want %s,actual %s", id, ps.ID)
	}
	if string(ps.Payload) != payload {
		t.Errorf("want %s,actual %s", payload, string(ps.Payload))
	}
}

func TestSubmitAck_Encode(t *testing.T) {
	id := fmt.Sprintf("%08d", 10) // 8 byte string
	ackResult := uint8(0)
	ack := &SubmitAck{
		ID:     id,
		Result: ackResult,
	}
	pkt, err := Encode(ack)
	if err != nil {
		t.Errorf("want nil,actual %s", err.Error())
	}

	commandID := int(pkt[0])
	pb := pkt[1:]
	ID := string(pb[:8])
	result := pb[8]
	if commandID != CommandSubmitAck {
		t.Errorf("want %d,actual %d", CommandSubmitAck, commandID)
	}

	if ID != id {
		t.Errorf("want %s,actual %s", id, ID)
	}

	if ackResult != result {
		t.Errorf("want %d,actual %d", ackResult, result)
	}

}

func TestSubmitAck_Decode(t *testing.T) {
	id := fmt.Sprintf("%08d", 10) // 8 byte string
	sa := SubmitAck{
		ID:     id,
		Result: 0,
	}
	b, err := sa.Encode()
	if err != nil {
		t.Errorf("want nil,actual %s", err.Error())
	}
	pkg := bytes.Join([][]byte{[]byte{CommandSubmitAck}, b}, nil)

	d, err := Decode(pkg)
	if err != nil {
		t.Errorf("want nil,actual %s", err.Error())
	}

	arr := d.(*SubmitAck)
	if arr.ID != id {
		t.Errorf("want %s,actual %s", id, arr.ID)
	}
	if arr.Result != 0 {
		t.Errorf("want 0,actual %d", arr.Result)
	}
}

type FailPacket struct {
}

func (f *FailPacket) Decode(i []byte) error {
	return nil
}

func (f *FailPacket) Encode() ([]byte, error) {
	return nil, nil
}

func TestEncodeWithFail(t *testing.T) {
	// switch default branch
	fp := &FailPacket{}
	_, err := Encode(fp)
	if err == nil {
		t.Errorf("want non-nil,actual nil")
	}
}

func TestDecodeWithFail(t *testing.T) {
	// switch default branch
	data := bytes.Join([][]byte{[]byte{0x21}, []byte{}}, nil)
	_, err := Decode(data)
	if err == nil {
		t.Errorf("want non-nil,actual nil")
	}
}
