package main

import (
	"fmt"
	"net"

	"github.com/sammyluck/tcp-server-demo1/frame"
	"github.com/sammyluck/tcp-server-demo1/packet"
)

// 处理 packet 包数据,Packet 是业务真正需要的消息
func handlePacket(framePayload []byte) (ackFramePayload []byte, err error) {
	var p packet.Packet
	p, err = packet.Decode(framePayload)
	if err != nil {
		fmt.Println("handleConn: packet decode error:", err)
		return
	}
	switch p.(type) {
	case *packet.Submit:
		submit := p.(*packet.Submit)
		fmt.Printf("recv submit: id = %s,payload=%s \n", submit.ID, string(submit.Payload))
		submitAck := &packet.SubmitAck{
			ID:     submit.ID,
			Result: 0,
		}
		ackFramePayload, err = packet.Encode(submitAck)
		if err != nil {
			fmt.Println("handleConn: packet encode error:", err)
			return nil, err
		}
		return ackFramePayload, nil
	default:
		return nil, fmt.Errorf("unknown packet type")
	}
}

// handle client connection
func handleConn(c net.Conn) {
	defer func() {
		c.Close()
		if err := recover(); err != nil {
			fmt.Printf("handleConn occurring error: recover panic[%s] and exit\n", err)
		}
	}()
	frameCodec := frame.NewMyFrameCodec()

	for {
		// read from the connection

		// decode the frame to get the payload
		// 从 connection 中读取  client 发送的数据内容
		framePayload, err := frameCodec.Decode(c)
		if err != nil {
			fmt.Println("handleConn: frame decode error:", err)
			return
		}

		//do something with the packet
		ackFramePayload, err := handlePacket(framePayload)
		if err != nil {
			fmt.Println("handleConn:handle packet error:", err)
			return
		}

		//write ack frame to the connection
		err = frameCodec.Encode(c, ackFramePayload)
		if err != nil {
			fmt.Println("handleConn: frame encode error:", err)
			return
		}
	}
}

func main() {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}
	// DeadLoop 不断监控是否有新的连接
	for {
		//在没有新连接的时候，这个服务会阻塞在 Accept 调用上，直到有客户端连接上来，Accept 方法将返回一个 net.Conn 实例
		c, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
		}

		fmt.Println("server start ok(on *.8888)")
		// start a new goroutine to handle the new connection.
		// the new connection.
		go handleConn(c)
	}
}
