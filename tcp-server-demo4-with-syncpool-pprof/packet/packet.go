package packet

import (
	"bytes"
	"fmt"
	"sync"
)

// Packet协议定义
/*
### packet header
1 byte: commandID
### packet body(Submit packet)
8字节 ID 字符串
任意字节 payload
### packet body(Submit ack packet)
8字节 ID 字符串
1字节 result
*/

// Packet header，用于表示这个消息的类型
const (
	CommandConn   = iota + 0x01 // 0x01，连接请求包
	CommandSubmit               // 0x02，消息请求包
)

// commandID: Packet header，用于表示这个消息的类型
const (
	CommandConnAck   = iota + 0x81 // 0x81,连接请求的响应包
	CommandSubmitAck               // 0x82,消息请求的响应包
)

type Packet interface {
	Decode([]byte) error     // []byte -> struct
	Encode() ([]byte, error) // struct -> []byte
}

// Conn 连接请求包
type Conn struct {
}

func (c *Conn) Decode(i []byte) error {
	//TODO
	return nil
}

func (c *Conn) Encode() ([]byte, error) {
	//TODO
	return nil, nil
}

// 连接响应包
type connAck struct {
}

func (c *connAck) Decode(i []byte) error {
	//TODO
	return nil
}

func (c *connAck) Encode() ([]byte, error) {
	//TODO
	return nil, nil
}

// Submit 消息请求包(packet body)，ID 和 payload
type Submit struct {
	ID      string
	Payload []byte
}

// Decode 解码 packet 包体
func (s *Submit) Decode(pktBody []byte) error {
	s.ID = string(pktBody[:8]) // 消息流水号(顺序累加，步长为1，循环使用)
	s.Payload = pktBody[8:]    // 消息的有效荷载，应用层需要的有效数据
	return nil
}

// Encode 编码 packet 包体
func (s *Submit) Encode() ([]byte, error) {
	return bytes.Join([][]byte{[]byte(s.ID[:8]), s.Payload}, nil), nil
}

// SubmitAck 消息响应包(packet body),ID 和 Result
type SubmitAck struct {
	ID     string // 消息流水号(顺序累加，步长为1，循环使用)
	Result uint8  // 响应状态（0：正常；1：错误）
}

func (s *SubmitAck) Decode(pktBody []byte) error {
	s.ID = string(pktBody[:8])
	s.Result = uint8(pktBody[8])
	return nil
}

func (s *SubmitAck) Encode() ([]byte, error) {
	return bytes.Join([][]byte{[]byte(s.ID[:8]), []byte{s.Result}}, nil), nil
}

var SubmitPool = sync.Pool{
	New: func() interface{} {
		return &Submit{}
	},
}

var SubmitAckPool = sync.Pool{
	New: func() interface{} {
		return &SubmitAck{}
	},
}

// Decode 解码 packet 包数据，负责从字节流中解析出对应的类型(根据 commandID)
func Decode(packet []byte) (Packet, error) {
	commandID := packet[0] // packet header
	pktBody := packet[1:]  // packet body

	switch commandID {
	case CommandConn:
		return nil, nil
	case CommandConnAck:
		return nil, nil
	case CommandSubmit:
		//s := Submit{}
		s := SubmitPool.Get().(*Submit) // 从 SubmitPool 池中获取一个 Submit 内存对象
		err := s.Decode(pktBody)
		if err != nil {
			return nil, err
		}
		return s, nil
	case CommandSubmitAck:
		//s := SubmitAck{}
		s := SubmitAckPool.Get().(*SubmitAck) // 从 SubmitPool 池中获取一个 SubmitAck 内存对象
		err := s.Decode(pktBody)
		if err != nil {
			return nil, err
		}
		return s, err
	default:
		return nil, fmt.Errorf("unknown commandID [%d]", commandID)
	}
}

// Encode 编码 packet 包数据，根据传入的 packet 类型调用对应的 Encode 方法现实对象的编码
func Encode(p Packet) ([]byte, error) {
	var commandID byte
	var pktBody []byte
	var err error

	switch t := p.(type) {
	case *Submit:
		commandID = CommandSubmit
		pktBody, err = p.Encode()
		if err != nil {
			return nil, err
		}
	case *SubmitAck:
		commandID = CommandSubmitAck
		pktBody, err = p.Encode()
		if err != nil {
			return nil, err
		}
	case *Conn:
		commandID = CommandConn
		pktBody, err = p.Encode()
		if err != nil {
			return nil, err
		}
	case *connAck:
		commandID = CommandConnAck
		pktBody, err = p.Encode()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown type [%s]", t)
	}
	// 封装 packet 包头和包体
	return bytes.Join([][]byte{[]byte{commandID}, pktBody}, nil), nil
}
