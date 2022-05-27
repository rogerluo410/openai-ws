package main

import (
	"fmt"
	"time"
	"context"
	"flag"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	hostUrl   = "ws://localhost:8080/ws"
	queryStr  = "provider=local&api_name=echo"
	clientNum = 10000
)

func main() {
	var num int
  flag.IntVar(&num, "c",  clientNum, "客户端数量")

	var address string
  flag.StringVar(&address, "a",  hostUrl, "代理ws服务地址")

	flag.Parse()

	var wg sync.WaitGroup
	handle := func(i int, ctx context.Context, cancelFunc context.CancelFunc) {
		d := websocket.Dialer{
			HandshakeTimeout: 180 * time.Second, // 5秒握手超时太短了
		}
		//握手并建立websocket 连接
		url := address+"?"+queryStr+"&token=1234"
		fmt.Printf("Clinet %d connect to %s\n", i, url)
	
		conn, resp, err := d.Dial(url, nil)
		if err != nil {
			fmt.Println(err.Error() + ", err != nil")
			return
		} else if resp.StatusCode != 101 {
			fmt.Println(err.Error() + ", resp.StatusCode != 101")
			return
		}

		defer func() {
			wg.Done()
			conn.Close()
		}()

	  conn.SetReadLimit(1024 * 1024)

		conn.SetPongHandler(func(string) error {
			conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(60 * time.Second))
			return nil
		})
		
		go func() {
      for {
				select {
				// 每20秒发送一次消息
				case <- time.After(20 * time.Second):
					now := time.Now()
					conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Client %d send message at %s", i, now.Format("2006-01-02 15:04:05"))) )
				case <- ctx.Done():
					fmt.Println("session end ---")
					return
				}
			}
		}()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("Client %d read message error: %s\n", i, err.Error())
				cancelFunc()
				break
			}
			fmt.Printf("Client %d read message: %s\n", i, string(msg))
		}
		
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	wg.Add(num)

	for i := 0; i < num; i++ {
		go handle(i, ctx, cancelFunc)

		// 休眠1秒
		time.Sleep(1 * time.Second)
	}

	wg.Wait()
	fmt.Println("Exit...")
}