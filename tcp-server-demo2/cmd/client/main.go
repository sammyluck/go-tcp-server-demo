package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/lucasepe/codename"
	"github.com/sammyluck/tcp-server-demo2/frame"
	"github.com/sammyluck/tcp-server-demo2/packet"
)

var num = 50

func startNewConn() {
	conn, err := net.Dial("tcp", ":8888")
	if err != nil {
		log.Println("dial error:", err)
		return
	}

	defer conn.Close()
	log.Printf("%s : dial ok \n", time.Now().Format("2006-01-02 15:04:05"))

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
			// 从 TCP 流的 io.Reader 中读取一个完整 Frame，并将得到的 frame payload，并返回给上层
			ackFramePayload, err := frameCodec.Decode(conn)
			if err != nil {
				panic(err)
			}

			p, err := packet.Decode(ackFramePayload)
			_, ok := p.(*packet.SubmitAck)
			if !ok {
				panic("not submitack")
			}

			//fmt.Printf("%s [client %d]: the result of submit ack[%s] is %d \n", time.Now().Format("2006-01-02 15:04:05"), i, submitAck.ID, submitAck.Result)
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

		//fmt.Printf("%s [client %d]: send submit id = %s,payload=%s,frame length = %d \n", time.Now().Format("2006-01-02 15:04:05"), i, s.ID, s.Payload, len(framePayload)+4)

		// encode the payload for the frame
		// 把数据内容通过 connection 发给 server
		err = frameCodec.Encode(conn, framePayload)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	var wg sync.WaitGroup
	//num := 1
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			startNewConn()
		}()
	}
	wg.Wait()
}
