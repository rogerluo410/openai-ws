package server

import (
	_ "io"
	"net/http"
	"time"

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
	list       []*Client
}

func NewServer() *Server {
	return &Server{
		list: make([]*Client, initCap),
	}
}

func (s *Server) addClient(c *Client) {
	s.list = append(s.list, c)
}

func (s *Server) listenClient() {
}

func (s *Server) ClientLength() int {
  return len(s.list)
}

// handleWebsocket connection.
func (s *Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	log.Info("验证Websocket连接串...")

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

	clientId := r.URL.Query()["client_id"]
	if len(clientId) == 0 {
		http.Error(w, "No client id error", http.StatusMethodNotAllowed)
		return
	}

	token := r.URL.Query()["token"]
	if len(token) == 0 {
		http.Error(w, "No token error", http.StatusMethodNotAllowed)
		return
	}

	log.Info("Websocket 连接成功...")

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
	log.WithFields(log.Fields{"服务提供商": provider[0], "api": apiName[0]}).Info("创建云端服务提供商连接串成功")

	client := NewClient(wsConn, cloudWsConn, provider[0], apiName[0], clientId[0], token[0], ws.RemoteAddr().String())
	s.addClient(client)
	// 启动Client
	client.Run()
}