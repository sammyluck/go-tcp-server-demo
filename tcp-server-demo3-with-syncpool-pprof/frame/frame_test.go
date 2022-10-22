package frame

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func Test_myFrameCodec_Encode(t *testing.T) {
	// 初始化 myFrameCodec
	codec := NewMyFrameCodec()
	buf := make([]byte, 0, 128)
	rw := bytes.NewBuffer(buf)
	// Encode 将输入的 Frame payload 编码为一个 Frame，并写入 io.Writer 所代表的输出 TCP 流中
	err := codec.Encode(rw, []byte("hello Gopher"))
	if err != nil {
		t.Errorf("want nil,actual %s ", err.Error())
	}

	// 验证 Encode 的正确性
	var totalLen int32
	err = binary.Read(rw, binary.BigEndian, &totalLen)
	if err != nil {
		t.Errorf("want nil, actual %s ", err.Error())
	}

	if totalLen != 16 {
		t.Errorf("want 16,actual %d ", totalLen)
	}

	left := rw.Bytes()
	if string(left) != "hello Gopher" {
		t.Errorf("want hello Gopher,actual %s", string(left))
	}
}

func Test_myFrameCodec_Decode(t *testing.T) {
	codec := NewMyFrameCodec()
	data := []byte{0x0, 0x0, 0x0, 0x10, 'h', 'e', 'l', 'l', 'o', ' ', 'G', 'o', 'p', 'h', 'e', 'r'}

	payload, err := codec.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("want nil,actual %s", string(payload))
	}
}

type ReturnErrorWriter struct {
	W  io.Writer
	Wn int // 第几次调用 Write 返回错误
	wc int // 写操作次数计数
}

func (w *ReturnErrorWriter) Write(p []byte) (n int, err error) {
	w.wc++
	if w.wc >= w.Wn {
		return 0, errors.New("write error")
	}
	return w.W.Write(p)
}

func NewReturnErrorWriter(data []byte, n int) *ReturnErrorWriter {
	return &ReturnErrorWriter{
		W:  bytes.NewBuffer(data),
		Wn: n,
	}
}

type ReturnErrorReader struct {
	R  io.Reader
	Rn int // 第几次调用 Read 返回错误
	rc int // 读操作次数计数
}

func (r *ReturnErrorReader) Read(p []byte) (n int, err error) {
	r.rc++
	if r.rc >= r.Rn {
		return 0, errors.New("read error")
	}
	return r.R.Read(p)
}

func NewReturnErrorReader(data []byte, n int) *ReturnErrorReader {
	return &ReturnErrorReader{
		R:  bytes.NewReader(data),
		Rn: n,
	}
}
func TestEncodeWithWriteFail(t *testing.T) {
	codec := NewMyFrameCodec()
	buf := make([]byte, 0, 128)

	// 模拟 binary.Write 返回错误
	err := codec.Encode(NewReturnErrorWriter(buf, 1), []byte("hello"))
	if err == nil {
		t.Errorf("want no-nil,actual nil")
	}

	// 模拟 w.Write 返回错误
	err = codec.Encode(NewReturnErrorWriter(buf, 2), []byte("hello"))
	if err == nil {
		t.Errorf("want non-nil,actual nil")
	}
}

func TestDecodeWithReadFail(t *testing.T) {
	codec := NewMyFrameCodec()
	data := []byte{0x0, 0x0, 0x0, 0x9, 'h', 'e', 'l', 'l', 'o'}

	// 模拟 binary.Read 返回错误
	_, err := codec.Decode(NewReturnErrorReader(data, 1))
	if err == nil {
		t.Errorf("want non-nil,actual nil")
	}

	// 模拟 io.ReadFull 返回错误
	_, err = codec.Decode(NewReturnErrorReader(data, 2))
	if err == nil {
		t.Errorf("want non-nil,actual nil")
	}
}
