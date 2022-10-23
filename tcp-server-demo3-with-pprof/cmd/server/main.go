package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/sammyluck/tcp-server-demo3/frame"
	"github.com/sammyluck/tcp-server-demo3/metrics"
	"github.com/sammyluck/tcp-server-demo3/packet"
)

/**
version 3 在可观测性监控基础上增加 read write buffer ，减少系统调用
*/
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
		//fmt.Printf("recv submit: id = %s,payload=%s \n", submit.ID, string(submit.Payload))
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
	metrics.ClientConnected.Inc() //连接建立，ClientConnected + 1
	defer func() {
		metrics.ClientConnected.Dec() // 连接断开，ClientConnected - 1
		c.Close()
		if err := recover(); err != nil {
			fmt.Printf("handleConn occurring error: recover panic[%s] and exit\n", err)
		}
	}()
	frameCodec := frame.NewMyFrameCodec()

	// 读缓存变量，避免每次都从 net.Conn 读取，降低 Syscall 调用频率
	rbuf := bufio.NewReader(c)
	// 写缓存变量
	wbuf := bufio.NewWriter(c)
	defer wbuf.Flush()
	for {
		// read from the connection

		// decode the frame to get the payload
		// 从 connection 中读取  client 发送的数据内容
		framePayload, err := frameCodec.Decode(rbuf)
		if err != nil {
			fmt.Println("handleConn: frame decode error:", err)
			return
		}

		metrics.ReqRecvTotal.Add(1) // 收到并解码一个消息请求，ReqRecvTotal 消息计数器 +1

		//do something with the packet
		ackFramePayload, err := handlePacket(framePayload)
		if err != nil {
			fmt.Println("handleConn:handle packet error:", err)
			return
		}

		//write ack frame to the connection
		err = frameCodec.Encode(wbuf, ackFramePayload)
		if err != nil {
			fmt.Println("handleConn: frame encode error:", err)
			return
		}
		metrics.RspSendTotal.Add(1) //返回响应后，RspSendTotal 消息计数器 -1
	}
}

func main() {
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}

	fmt.Println("server start ok(on *:8888)")
	// DeadLoop 不断监控是否有新的连接
	for {
		//在没有新连接的时候，这个服务会阻塞在 Accept 调用上，直到有客户端连接上来，Accept 方法将返回一个 net.Conn 实例
		c, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			break
		}

		// start a new goroutine to handle the new connection.
		// the new connection
		go handleConn(c)
	}
}
