package collect

import (
	"io"
	"net/http"
	"time"
	"github.com/sirupsen/logrus"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		HandshakeTimeout: 62 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			//r.URL *url.URL
      //r.Header Header
			return true
		},
	}
)

type Client struct {
	pWs            *Ws              // ws
	message        chan string      // 消息管道
	uuid           string           // 客户端id
	token          string           // 认证token
	address        string           // 客户端ip
  verified       bool             // token认证状态 
	connected      bool             // ws连接状态
}


func (c *Client) create() {
	c.message = make(chan Message)
}

func (c *Client) handler() {
	for {
		select {
			case msg := <- c.message:
				// 处理消息
		
		}
	}
}