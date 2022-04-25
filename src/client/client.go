package client

import (
	_ "io"
	_ "net/http"
	_ "time"
)

type Client struct {
	pWs            *Ws              // ws
	sendMsg        chan string      // 客户端发送消息管道
	receiveMsg     chan string      // 客户端接收消息管道
	route          string           // 路由
	uuid           string           // 客户端id
	token          string           // 认证token
	address        string           // 客户端ip
  verified       bool             // token认证状态 
	connected      bool             // ws连接状态
}

func New(route string, uuid string, token string, address string) *Client {
  // Verify token from openai_backend

	return &Client{
		route: route,
		uuid: uuid,
		token: token,
		address: address,
		sendMsg: make(chan string),
		receiveMsg: make(chan string),
	}
}