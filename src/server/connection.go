package server

import (
	"io"
	"time"
	"errors"
	"reflect"

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

type WsConn struct {
	Conn *websocket.Conn
}

func NewWsConn(conn *websocket.Conn) *WsConn {
	return &WsConn{
		Conn: conn,
	}
}

func (w *WsConn) Close() {
  w.Conn.Close()
}

// 客户端连接处理
func (w *WsConn) Reader(client *Client) {
	defer func() {
		log.Info("Client - Reader 协程退出...")
		w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
		w.Close()
		client.Wg.Done()
	}()

	client.Wg.Add(1)

	w.Conn.SetReadLimit(1024 * 1024)
	w.Conn.SetReadDeadline(time.Now().Add(pongWait))

	w.Conn.SetPongHandler(func(string) error { 
		w.Conn.SetReadDeadline(time.Now().Add(pongWait))
		w.Conn.WriteMessage(websocket.PongMessage, []byte{})
		return nil 
	})

	w.Conn.SetCloseHandler(func(code int, text string) error {
		log.WithFields(log.Fields{"code": code, "text": text}).Info("Client - Reader 收到客户端关闭消息")
		w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
    return nil
	})

	for {
		_, msg, err := w.Conn.ReadMessage()
		l := log.WithFields(log.Fields{ "Msg": string(msg), "Err": err})

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
				w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				l.Info("Client - Reader Websocket 连接关闭")
			} else {
				w.Conn.WriteMessage(websocket.CloseAbnormalClosure, []byte(err.Error()))
				l.Error("Client - Reader Websocket读消息失败, 将关闭websocket连接")
			}
			// 如果遇到ws读错误，则关闭websocket连接
			return
		}

		// 写入管道
		l.WithFields(log.Fields{ "Msg": string(msg)}).Info("客户端发送数据, 结构化后传入云端服务")
		client.Msg <- string(msg)
	}
}

func (w *WsConn) Writer(client *Client) {
	pingTicker := time.NewTicker(pingPeriod)
	client.Wg.Add(1)

	defer func() {
		log.Info("Client - Writer 协程退出...")
		pingTicker.Stop()
		w.Close()
		client.Wg.Done()
	}()

	for {
		select {
			case <- pingTicker.C:
				log.Info("Client - Writer 发送心跳包...")
			  w.Conn.SetWriteDeadline(time.Now().Add(writeWait))

				err := w.Conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Error("Client - Writer 发送心跳包失败")
					return
				}
		  case msg := <- client.CloudMsg:
				str1, _ := msg.(string)
				w.Conn.WriteMessage(websocket.TextMessage, []byte(str1))				
		}
	}
}

// 云端连接处理
func (w *WsConn) CloudReader(client *Client) {
	defer func() {
		log.Info("Cloud - Reader 协程退出...")
		w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
		w.Close()
		client.Wg.Done()
	} ()

	client.Wg.Add(1)
	w.Conn.SetReadLimit(1024 * 1024)
	w.Conn.SetReadDeadline(time.Now().Add(pongWait))

	w.Conn.SetPongHandler(func(string) error { 
		w.Conn.SetReadDeadline(time.Now().Add(pongWait))
		w.Conn.WriteMessage(websocket.PongMessage, []byte{})
		return nil 
	})

	w.Conn.SetCloseHandler(func(code int, text string) error {
		log.WithFields(log.Fields{"code": code, "text": text}).Info("Cloud - Reader 收到客户端关闭消息")
		w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
    return nil
	})

	for {
		_, msg, err := w.Conn.ReadMessage()
		l := log.WithFields(log.Fields{ "Msg": string(msg), "Err": err})

		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
				w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				l.Info("Cloud - Reader Websocket 连接关闭")
			} else {
				w.Conn.WriteMessage(websocket.CloseAbnormalClosure, []byte(err.Error()))
				l.Error("Cloud - Reader Websocket读消息失败, 将关闭websocket连接")
			}
			// 如果遇到ws读错误，则关闭websocket连接
			return
		}

		// 通知客户端
		l.Info("Cloud - Reader - 云端返回数据, 通知客户端接收")
		// resp := ProviderResponseMapper(msg, client)

		// 云端返回数据透传给客户端
		client.CloudMsg <- string(msg)
	}
}

func (w *WsConn) CloudWriter(client *Client) {
	client.Wg.Add(1)
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		log.Info("Cloud - Writer 协程退出...")
		pingTicker.Stop()
		w.Close()
		client.Wg.Done()
	}()

	for {
		select {
			case <- pingTicker.C:
				log.Info("Cloud - Writer 发送心跳包...")
			  w.Conn.SetWriteDeadline(time.Now().Add(writeWait))

				err := w.Conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Error("Cloud - Writer 发送心跳包失败")
					return
				}
		  case msg := <- client.Msg:
        var err error
				if m := ProviderRequestMapper(msg, client); m != nil {
					v := reflect.ValueOf(m)
					switch v.Kind() {
					case reflect.String:
						str, _ := m.(string)
						err = w.Conn.WriteMessage(websocket.TextMessage, []byte(str))
						break
					case reflect.Map:
						err = w.Conn.WriteJSON(m)
						break
					default:
						err = errors.New("Cloud - Writer 未知消息类型")
					}

          if err != nil {
						// 如果遇到ws写错误，则关闭websocket连接
						log.WithFields(log.Fields{
							"data": m,
							"err":  err,
						}).Error("Cloud - Writer Websocket写消息失败, 将关闭websocket连接")
						return
					}
				}
		}
	}
}


