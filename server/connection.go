package server

import (
	"context"
	"errors"
	"io"
	"reflect"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 60 * time.Second

	// Time allowed to read the file from the client
	readWait = 60 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

)

// Fix to Connections support one concurrent reader and one concurrent writer.

type WsConn struct {
	Conn  *websocket.Conn
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
func (w *WsConn) Reader(client *Client, ctx context.Context, cancelFunc context.CancelFunc) {
	defer func() {
		log.Info("Client - Reader 协程退出...")
		w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
		client.Wg.Done()
	}()

	w.Conn.SetReadLimit(1024 * 1024)

	w.Conn.SetPongHandler(func(string) error {
		w.Conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(pongWait))
		return nil
	})

	w.Conn.SetCloseHandler(func(code int, text string) error {
		log.WithFields(log.Fields{"code": code, "text": text}).Info("Client - Reader 收到客户端关闭消息")
		w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
		return nil
	})

	for {
		select {
		case <- ctx.Done():
			return
		default:
			w.Conn.SetReadDeadline(time.Now().Add(readWait))
			_, msg, err := w.Conn.ReadMessage()
			l := log.WithFields(log.Fields{"Msg": string(msg), "Err": err})

			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
					w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
					l.Info("Client - Reader Websocket 连接关闭")
				} else {
					w.Conn.WriteControl(websocket.CloseAbnormalClosure, []byte(err.Error()), time.Now().Add(writeWait))
					l.Error("Client - Reader Websocket读消息失败, 将关闭websocket连接")
				}
				// 如果遇到ws读错误，则关闭websocket连接
				cancelFunc() // 通知其他协程退出  
				return
			}

			// 写入管道
			l.WithFields(log.Fields{"Msg": string(msg)}).Info("客户端发送数据, 结构化后传入云端服务")
			client.Msg <- string(msg)
		}
	}
}

func (w *WsConn) Writer(client *Client, ctx context.Context, cancelFunc context.CancelFunc) {
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		log.Info("Client - Writer 协程退出...")
		pingTicker.Stop()
		client.Wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			log.Info("Client - Writer 发送心跳包...")
			w.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
			if err != nil {
				log.Error("Client - Writer 发送心跳包失败")
				cancelFunc()
				return
			}
		case msg := <-client.CloudMsg:
			str1, _ := msg.(string)
			w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := w.Conn.WriteMessage(websocket.TextMessage, []byte(str1))
			if err != nil {
				log.WithFields(log.Fields{"Err": err}).Error("Client - Writer 发数据失败")
				cancelFunc()
				return
			}
		}
	}
}

// 云端连接处理
func (w *WsConn) CloudReader(client *Client, ctx context.Context, cancelFunc context.CancelFunc) {
	defer func() {
		log.Info("Cloud - Reader 协程退出...")
		w.Conn.WriteMessage(websocket.CloseMessage, []byte{})
		client.Wg.Done()
	}()

	w.Conn.SetReadLimit(1024 * 1024)

	w.Conn.SetPongHandler(func(string) error {
		w.Conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(pongWait))
		return nil
	})

	w.Conn.SetCloseHandler(func(code int, text string) error {
		log.WithFields(log.Fields{"code": code, "text": text}).Info("Cloud - Reader 收到客户端关闭消息")
		w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			w.Conn.SetReadDeadline(time.Now().Add(readWait))
			_, msg, err := w.Conn.ReadMessage()
			l := log.WithFields(log.Fields{"Msg": string(msg), "Err": err})

			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
					w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
					l.Info("Cloud - Reader Websocket 连接关闭")
				} else {
					w.Conn.WriteControl(websocket.CloseAbnormalClosure, []byte(err.Error()), time.Now().Add(writeWait))
					l.Error("Cloud - Reader Websocket读消息失败, 将关闭websocket连接")
				}
				// 如果遇到ws读错误，则关闭websocket连接
				cancelFunc()
				return
			}

			// 通知客户端
			l.Info("Cloud - Reader - 云端返回数据, 通知客户端接收")
			// 云端返回数据透传给客户端
			client.CloudMsg <- string(msg)
		}
	}
}

func (w *WsConn) CloudWriter(client *Client, ctx context.Context, cancelFunc context.CancelFunc) {
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		log.Info("Cloud - Writer 协程退出...")
		pingTicker.Stop()
		client.Wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			log.Info("Cloud - Writer 发送心跳包...")

			err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
			if err != nil {
				log.Error("Cloud - Writer 发送心跳包失败")
				return
			}
		case msg := <-client.Msg:
			var err error
			if m := ProviderRequestMapper(msg, client); m != nil {
				v := reflect.ValueOf(m)
				switch v.Kind() {
				case reflect.String:
					str, _ := m.(string)
					log.WithField("发给云端服务的字节流:", str).Info("发送字节流")
					w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
					err = w.Conn.WriteMessage(websocket.TextMessage, []byte(str))
					break
				case reflect.Map:
					log.WithField("发给云端服务的json:", m).Info("发送json")
					w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
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
					cancelFunc()
					return
				}

				// 休眠一秒钟
				// time.Sleep(1 * time.Second)
			}
		}
	}
}

// 客户端连接处理Echo
func (w *WsConn) ReaderEcho(client *Client, ctx context.Context, cancelFunc context.CancelFunc) {
	defer func() {
		log.Info("Client - ReaderEcho 协程退出...")
		w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
		client.Wg.Done()
	}()

	w.Conn.SetReadLimit(1024 * 1024)

	w.Conn.SetPongHandler(func(string) error {
		w.Conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(pongWait))
		return nil
	})

	w.Conn.SetCloseHandler(func(code int, text string) error {
		log.WithFields(log.Fields{"code": code, "text": text}).Info("Client - Reader 收到客户端关闭消息")
		w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
		return nil
	})

	for {
		select {
		case <- ctx.Done():
			return
		default:
			w.Conn.SetReadDeadline(time.Now().Add(readWait))
			_, msg, err := w.Conn.ReadMessage()
			l := log.WithFields(log.Fields{"Msg": string(msg), "Err": err})

			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
					w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
					l.Info("Client - ReaderEcho Websocket 连接关闭")
				} else {
					w.Conn.WriteControl(websocket.CloseAbnormalClosure, []byte(err.Error()), time.Now().Add(writeWait))
					l.Error("Client - ReaderEcho Websocket读消息失败, 将关闭websocket连接")
				}
				// 如果遇到ws读错误，则关闭websocket连接
				cancelFunc()
				return
			}

			// 写入管道
			// l.WithFields(log.Fields{"Msg": string(msg)}).Info("客户端发送数据, WriterEcho回显")
			client.Msg <- string(msg)
		}
	}
}

func (w *WsConn) WriterEcho(client *Client, ctx context.Context, cancelFunc context.CancelFunc) {
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		log.Info("Client - WriterEcho 协程退出...")
		pingTicker.Stop()
		client.Wg.Done()
	}()

	for {
		select {
		case <- ctx.Done():
			return
		case <-pingTicker.C:
			// log.Info("Client - WriterEcho 发送心跳包...")
			err := w.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
			if err != nil {
				log.Error("Client - WriterEcho 发送心跳包失败")
				cancelFunc()
				return
			}
		case msg := <- client.Msg:
			str1, _ := msg.(string)
			now := time.Now()
			str1 = "服务端收到消息: `" + str1 + "`, 回显时间: " + now.Format("2006-01-02 15:04:05")
			w.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := w.Conn.WriteMessage(websocket.TextMessage, []byte(str1))
			if err != nil {
				log.WithFields(log.Fields{"Err": err}).Error("Client - WriterEcho 发数据失败")
				cancelFunc()
				return
			}
		}
	}
}