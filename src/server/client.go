package server

import (
	_ "io"
	_ "net/http"
	_ "time"
)

type Client struct {
	Conn            *WsConn         // ws
	SendMsg        chan string      // 客户端消息管道
	Provider       string           // 提供商
	ApiName        string           // api服务名
	ClientId       string           // 客户端id
	Token          string           // 认证token
	Address        string           // 客户端ip
  Verified       bool             // token认证状态 
	Connected      bool             // ws连接状态
}

func NewClient(
	conn      *WsConn,
	provider string,
	apiName string,
	clientId string, 
	token string, 
	address string,
) *Client {
  // Verify token from openai_backend

	return &Client{ Conn: conn,
		Provider: provider,
		ApiName: apiName,
		ClientId: clientId,
		Token: token,
		Address: address,
		SendMsg: make(chan string),
	}
}