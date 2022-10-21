package frame

import (
	"encoding/binary"
	"errors"
	"io"
)

//  frame 包的职责是提供识别 TCP 流边界的编解码器
/*
Frame定义
frameHeader + framePayload(packet)
frameHeader
	4 bytes: length 整型，帧总长度(含头及payload)
framePayload
	Packet
*/

type FramePayload []byte

type StreamFrameCodec interface {
	Encode(io.Writer, FramePayload) error   // 将输入的 Frame payload 编码为一个 Frame，并写入 io.Writer 所代表的输出 TCP 流中
	Decode(io.Reader) (FramePayload, error) // 从 TCP 流的 io.Reader 中读取一个完整 Frame，并将得到的 frame payload，并返回给上层
}

var ErrShortWrite = errors.New("short write")
var ErrShortRead = errors.New("short read")

type myFrameCodec struct{}

func NewMyFrameCodec() StreamFrameCodec {
	return &myFrameCodec{}
}

// Encode 将输入的 Frame payload 编码为一个 Frame，并写入 io.Writer 所代表的输出 TCP 流中
func (m *myFrameCodec) Encode(w io.Writer, framePayload FramePayload) error {

	// totalLen = 消息总长度，含自身(4 个字节) + 后面消息体长度
	// totalLen 使用 int32，那么写入只会操作数据流中的 4 个字节。
	totalLen := int32(len(framePayload)) + 4
	// 大端字节序,根据参数的 宽度 写入对应的字节个数的字节
	err := binary.Write(w, binary.BigEndian, &totalLen)
	if err != nil {
		return err
	}

	n, err := w.Write(framePayload)
	if err != nil {
		return err
	}

	if n != len(framePayload) {
		return ErrShortWrite
	}
	return nil
}

// Decode 从 TCP 流的 io.Reader 中读取一个完整 Frame，并将得到的 frame payload，并返回给上层
func (m *myFrameCodec) Decode(r io.Reader) (FramePayload, error) {
	// totalLen 使用 int32，那么读取只会操作数据流中的 4 个字节。
	var totalLen int32
	// 大端字节序,根据参数的 宽度 读取对应的字节个数的字节
	err := binary.Read(r, binary.BigEndian, &totalLen)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, totalLen-4)
	// io.ReadFull 一般会读满所需要的字节数,除非遇到 EOF 或 ErrUnexpectedEOF
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	if n != int(totalLen-4) {
		return nil, ErrShortRead
	}

	return buf, nil
}
