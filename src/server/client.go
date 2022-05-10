package server

import (
	"context"
	"sync"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Conn      *WsConn          // ws
	CloudConn *WsConn          // 云端连接
	Msg       chan interface{} // 客户端消息管道
	CloudMsg  chan interface{} // 云端消息管道
	Provider  string           // 提供商
	ApiName   string           // api服务名
	ClientId  string           // 客户端id
	Token     string           // 认证token
	Address   string           // 客户端ip
	Uuid      string           // Client唯一标识
	Wg        *sync.WaitGroup
}

func NewClient(
	conn *WsConn,
	cloudConn *WsConn,
	provider string,
	apiName string,
	token string,
	address string,
) *Client {
	uuid := uuid.New()
	return &Client{Conn: conn,
		CloudConn: cloudConn,
		Provider:  provider,
		ApiName:   apiName,
		Token:     token,
		Address:   address,
		Msg:       make(chan interface{}),
		CloudMsg:  make(chan interface{}),
		Uuid:      uuid.String(),
		Wg:        &sync.WaitGroup{},
	}
}

func (c *Client) Close() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("Channel is already closed...")
		}
	}()
	close(c.Msg)
	close(c.CloudMsg)
}

func (c *Client) Run(s *Server) {
	go func() {
		defer c.Close()
		ctx, cancelFunc := context.WithCancel(context.Background())
		go c.Conn.Reader(c, ctx, cancelFunc)
		go c.Conn.Writer(c, ctx)
		go c.CloudConn.CloudReader(c, ctx)
		go c.CloudConn.CloudWriter(c, ctx)

		log.Info("阻塞, 等待读写协程结束...")
		c.Wg.Wait()
		// 全部读写websocket退出, 通知Server删除客户端变量
		log.WithField("Client uuid", c.Uuid).Info("全部读写websocket退出, 将通知Server删除客户端")
		s.Rmsg <- c.Uuid
	}()
}
