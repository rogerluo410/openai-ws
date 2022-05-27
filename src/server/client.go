package server

import (
	"context"
	"sync"

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
	Wg        sync.WaitGroup
	Actived   bool             // 活跃标记  
}

func NewClient(
	conn *WsConn,
	cloudConn *WsConn,
	provider string,
	apiName string,
	token string,
	address string,
) *Client {
	return &Client{Conn: conn,
		CloudConn: cloudConn,
		Provider:  provider,
		ApiName:   apiName,
		Token:     token,
		Address:   address,
		Msg:       make(chan interface{}, 10),
		CloudMsg:  make(chan interface{}, 10),
		Wg:        sync.WaitGroup{},
		Actived:   true,
	}
}

func (c *Client) Close() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("Channel is already closed...")
		}
	}()

	c.Actived = false

	// 关闭channel
	close(c.Msg)
	close(c.CloudMsg)

	// 关闭ws conn连接
	c.Conn.Close()
	c.CloudConn.Close()
}

func (c *Client) Run() {
	defer c.Close()
	c.Wg.Add(4)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go c.Conn.Reader(c, ctx, cancelFunc)
	go c.Conn.Writer(c, ctx)
	go c.CloudConn.CloudReader(c, ctx)
	go c.CloudConn.CloudWriter(c, ctx)

	log.Info("阻塞, 等待读写协程结束...")
	c.Wg.Wait()
	// 全部读写websocket退出, 通知Server删除客户端变量
	log.WithField("Client Address", c.Address).Info("全部读写websocket退出, 将通知Server删除客户端")
}

func (c *Client) RunEcho() {
	defer c.Close()
	c.Wg.Add(2)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go c.Conn.ReaderEcho(c, ctx, cancelFunc)
	go c.Conn.WriterEcho(c, ctx)

	c.Wg.Wait()
}

