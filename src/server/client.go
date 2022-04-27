package server

// import (
// 	"github.com/gorilla/websocket"
// )

type Client struct {
	Conn           *WsConn           // ws
	CloudConn      *WsConn           // 云端连接
	Msg            chan interface{}  // 客户端消息管道
	CloudMsg       chan interface{}  // 云端消息管道
	Provider       string            // 提供商
	ApiName        string            // api服务名
	ClientId       string            // 客户端id
	Token          string            // 认证token
	Address        string            // 客户端ip
  Verified       bool              // token认证状态 
	Connected      bool              // ws连接状态
}

func NewClient(
	conn      *WsConn,
	cloudConn *WsConn,
	provider string,
	apiName string,
	clientId string, 
	token string, 
	address string,
) *Client {
  // Verify token from openai_backend

	return &Client{ Conn: conn,
		CloudConn: cloudConn,
		Provider: provider,
		ApiName: apiName,
		ClientId: clientId,
		Token: token,
		Address: address,
		Msg: make(chan interface{}),
		CloudMsg: make(chan interface{}),
	}
}

func (c *Client) Run() {
	go c.Conn.Reader(c)
	go c.Conn.Writer(c)
	go c.CloudConn.CloudReader(c)
	go c.CloudConn.CloudWriter(c)
}