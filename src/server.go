package main

import (
	_ "io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	client "github.com/rogerluo410/openai-ws/src/client"
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
	list       []*client.Client
}

func NewServer() *Server {
	return &Server{
		list: make([]*client.Client, initCap),
	}
}

func (s *Server) addClient(c *client.Client) {
	s.list = append(s.list, c)

	// 启动 ws读写监听
	go c.Conn.Reader(c)
	go c.Conn.Writer(c)
}

func (s *Server) listenClient() {
}

func (s *Server) clientLength() int {
  return len(s.list)
}

// handleWebsocket connection.
func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
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
  conn := client.NewWs(ws)
	client := client.NewClient(conn, provider[0], apiName[0], clientId[0], token[0], ws.RemoteAddr().String())
	s.addClient(client)
}