package server

import (
	"sync"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

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
	Uuid           string            // Client唯一标识
	Wg             *sync.WaitGroup   
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

	uuid := uuid.New()
	return &Client{ Conn: conn,
		CloudConn: cloudConn,
		Provider: provider,
		ApiName: apiName,
		ClientId: clientId,
		Token: token,
		Address: address,
		Msg: make(chan interface{}),
		CloudMsg: make(chan interface{}),
		Uuid: uuid.String(),
		Wg: &sync.WaitGroup{},
	}
}

func (c *Client) Run(s *Server) {
	go func() {
    go c.Conn.Reader(c)
		go c.Conn.Writer(c)
		go c.CloudConn.CloudReader(c)
		go c.CloudConn.CloudWriter(c)
    
    log.Info("阻塞, 等待读写协程结束...")
		c.Wg.Wait()
		// 全部读写websocket退出, 通知Server删除客户端变量
		log.WithField("Client uuid", c.Uuid).Info("全部读写websocket退出, 将通知Server删除客户端")
    s.Rmsg <- c.Uuid
	}()
}