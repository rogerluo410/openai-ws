package server

import (
	"net/http"
	"time"
	"context"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var (
	initCap = 100
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

type Server struct {
	list         []*Client
	indexes      map[string]int  // 客户端位置索引
	Rmsg         chan string     // 接收客户端退出消息
	MaxCnt       uint            // 最大客户连接数
}

func NewServer(maxClientCnt uint) *Server {
	return &Server{
		list: make([]*Client, 0, initCap),
    indexes: make(map[string]int),
		Rmsg: make(chan string),
		MaxCnt: maxClientCnt,
	}
}

func (s *Server) addClient(c *Client) {
	s.list = append(s.list, c)
	s.indexes[c.Uuid] = len(s.list) - 1
}

func (s *Server) removeClient(uuid string) {
	index := s.indexes[uuid]
	if index >= 0 {
		s.list = append(s.list[:index], s.list[index+1:]...)
	}
}

func (s *Server) Listen(ctx context.Context) {
  go func() {
    for {
			select {
			// 每隔10分钟查看一次客户数量
			case <- time.After(10 * time.Minute):
				log.WithField("Num", s.ActiveClients()).Info("当前活跃用户数")
			case m := <- s.Rmsg:
        s.removeClient(m)
			case <- ctx.Done():
				log.Info("Server listen canceled...")
				return
			}
		}
	}()
}

func (s *Server) ActiveClients() int {
  return len(s.list)
}

// handleWebsocket connection.
func (s *Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	if uint(s.ActiveClients()) >= s.MaxCnt {
    http.Error(w, "Up to max connections", http.StatusForbidden)
		return
	}

	log.Info("接收到请求, 开始验证Websocket连接串...")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	provider := r.URL.Query()["provider"]
	if len(provider) == 0 {
		http.Error(w, "No provider error", http.StatusMethodNotAllowed)
		return
	}

	apiName := r.URL.Query()["api_name"]
	if len(apiName) == 0 {
		http.Error(w, "No api name error", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query()["token"]
	if len(token) == 0 {
		http.Error(w, "No token error", http.StatusMethodNotAllowed)
		return
	}

	// Verify token from openai_backend

	log.Info("路由参数解析成功并且token认证成功, 将升级为Websocket服务...")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m := "Unable to upgrade to websockets"
		log.WithField("err", err).Println(m)
		http.Error(w, m, http.StatusBadRequest)
		return
	}
	
	//注册client
  wsConn := NewWsConn(ws)
  
	cloudWs, err := ProviderWsMapper(provider[0], apiName[0])
	if err != nil {
		log.WithField("err", err).Error("Create cloud ws conn failed...")
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	cloudWsConn := NewWsConn(cloudWs)
	log.WithFields(log.Fields{"服务提供商": provider[0], "api": apiName[0]}).Info("创建云端服务提供商Websocket连接串成功")

	client := NewClient(wsConn, cloudWsConn, provider[0], apiName[0], token[0], ws.RemoteAddr().String())
	s.addClient(client)
	// 启动Client
	client.Run(s)
}