package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/lucasepe/codename"
	"github.com/sammyluck/tcp-server-demo1/frame"
	"github.com/sammyluck/tcp-server-demo1/packet"
)

var num = 10

func main() {
	var wg sync.WaitGroup
	//num := 1
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func(i int) {
			defer wg.Done()
			startClient(i)
		}(i + 1)
	}
	wg.Wait()
}

func startClient(i int) {
	quit := make(chan struct{})
	done := make(chan struct{})
	conn, err := net.Dial("tcp", ":8888")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}

	defer conn.Close()
	fmt.Printf("%s [client %d]: dial ok \n", time.Now().Format("2006-01-02 15:04:05"), i)

	// 生成 payload
	rng, err := codename.DefaultRNG()
	if err != nil {
		panic(err)
	}

	frameCodec := frame.NewMyFrameCodec()
	var counter int

	go func() {
		// handle ack
		for {
			select {
			case <-quit:
				done <- struct{}{}
				return
			default:
			}
			// 设置读操作的超时时间，当超时后仍然没有数据可读的情况下，Read 操作会解除阻塞并返回超时错误
			_ = conn.SetReadDeadline(time.Now().Add(time.Second * 5))
			// 从 TCP 流的 io.Reader 中读取一个完整 Frame，并将得到的 frame payload，并返回给上层
			// 如果没有数据可读，则阻塞 5s
			ackFramePayload, err := frameCodec.Decode(conn)
			if err != nil {
				if e, ok := err.(net.Error); ok {
					if e.Timeout() {
						// 进行其他业务逻辑的处理
						continue
					}
				}
				panic(err)
			}

			p, err := packet.Decode(ackFramePayload)
			submitAck, ok := p.(*packet.SubmitAck)
			if !ok {
				panic("not submitack")
			}

			fmt.Printf("%s [client %d]: the result of submit ack[%s] is %d \n", time.Now().Format("2006-01-02 15:04:05"), i, submitAck.ID, submitAck.Result)
		}
	}()

	for {
		// send submit
		counter++
		id := fmt.Sprintf("%08d", counter) // 8 byte string
		payload := codename.Generate(rng, 4)
		s := &packet.Submit{
			ID:      id,
			Payload: []byte(payload),
		}
		// 编码 packet (packet header + packet body) 包，即编码 frame body
		framePayload, err := packet.Encode(s)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s [client %d]: send submit id = %s,payload=%s,frame length = %d \n", time.Now().Format("2006-01-02 15:04:05"), i, s.ID, s.Payload, len(framePayload)+4)

		// encode the payload for the frame
		// 把数据内容通过 connection 发给 server
		err = frameCodec.Encode(conn, framePayload)
		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
		if counter >= num {
			quit <- struct{}{}
			<-done
			fmt.Printf("%s [client %d]:exit ok \n", time.Now().Format("2006-01-02 15:04:05"), i)
			return
		}
	}
}
