package main

import (
	"io"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 60 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type Ws struct {
	conn *websocket.Conn
}

func NewWs(conn *websocket.Conn) *Ws {
	return &Ws{
		conn: conn,
	}
}

func (w *Ws) Close() {
  w.conn.Close()
}

func (w *Ws) Reader(client *Client, processor *Processor) {
	defer func() {
		log.Fatal("Reader 退出...")
		w.conn.WriteMessage(websocket.CloseMessage, []byte{})
		w.conn.Close()
	} ()

	w.conn.SetReadLimit(1024 * 1024)
	w.conn.SetReadDeadline(time.Now().Add(pongWait))

	w.conn.SetPongHandler(func(string) error { 
		w.conn.SetReadDeadline(time.Now().Add(pongWait))
		w.conn.WriteMessage(websocket.PongMessage, []byte{})
		return nil 
	})

	w.conn.SetCloseHandler(func(code int, text string) error {
		log.WithFields(log.Fields{"code": code, "text":text}).Info("收到客户端关闭消息")
		w.conn.WriteMessage(websocket.CloseMessage, []byte{})
    return nil
	})

	for {
		_, msg, err := w.conn.ReadMessage()
		l := log.WithFields(log.Fields{ "Msg": string(msg), "Err": err})

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
				w.conn.WriteMessage(websocket.CloseMessage, []byte{})
				l.Info("Websocket 连接关闭!")
			} else {
				w.conn.WriteMessage(websocket.CloseAbnormalClosure, []byte(err.Error()))
				l.Error("Websocket读消息失败, 将关闭websocket连接")
			}
			// 如果遇到ws读错误，则关闭websocket连接
			break
		}

		// 写入管道
		l.Info("客户端发送数据!")
		job := &Job{Client: client, Sequence: 1, Type: 0, Message: string(msg)}
		processor.SendMsg <- job
	}
}

func (w *Ws) Writer(client *Client, processor *Processor) {
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		w.conn.Close()
	}()

	for {
		select {
			case <- pingTicker.C:
				log.Info("发送心跳包...")
			  w.conn.SetWriteDeadline(time.Now().Add(writeWait))

				err := w.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Error("发送心跳包失败")
					return
				}
		  case msg := <- client.SendMsg:
				if err := w.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					// 如果遇到ws写错误，则关闭websocket连接
					log.WithFields(log.Fields{
						"data": msg,
						"err":  err,
					}).Error("Websocket写消息失败, 将关闭websocket连接")
					return
				}
		}
	}
}



