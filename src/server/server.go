package server

import (
	"net/http"
	"net/url"
	"strings"
	"time"
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/alitto/pond"
	log "github.com/sirupsen/logrus"
)

var (
	initCap = 100
	poolNum = 1000
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		HandshakeTimeout: 180 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			//r.URL *url.URL
      //r.Header Header
			return true
		},
	}
)

type Server struct {
	list         []*Client
	Rmsg         chan string     // 接收客户端退出消息
	MaxCnt       uint            // 最大客户连接数
	VerfiyUrl    string          // 认证服务器地址
	lock         *sync.Mutex       
	pool         *pond.WorkerPool
}

func NewServer(tokenUrl string, maxClientCnt uint) *Server {
	return &Server{
		list: make([]*Client, 0, initCap),
		Rmsg: make(chan string),
		MaxCnt: maxClientCnt,
		VerfiyUrl: tokenUrl,
		lock: &sync.Mutex{},
		pool: pond.New(poolNum, int(maxClientCnt), pond.Strategy(pond.Balanced())),
	}
}

func (s *Server) addClient(c *Client) {
	s.lock.Lock()
  defer s.lock.Unlock()
	s.list = append(s.list, c)
}

func (s *Server) removeClients() {
	for index, client := range s.list {
		if client.Actived == false {
			s.lock.Lock()
			len := len(s.list)

			// panic: panic: runtime error: slice bounds out of range [6:4]
			// 为了解决非线程安全的数据越界问题， 需加上数组下标小于数组长度的判断  
			if index < len {
				s.list = append(s.list[:index], s.list[index+1:]...)
			}
			s.lock.Unlock()
		}
	}
}

func (s *Server) Close() {
	s.pool.Stop()
	close(s.Rmsg)
}

func (s *Server) Listen(ctx context.Context) {
  go func() {
    for {
			select {
			// 每隔10秒查看一次客户数量
			case <- time.After(10 * time.Second):
				s.removeClients()
				log.WithField("Num", s.ActiveClients()).Info("当前活跃用户数")
			case m := <- s.Rmsg:
				log.WithField("Client Address", m).Info("Server收到客户端结束通信")
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

func (s *Server) VerifyToken(token string) bool {
	if len(token) == 0 {
		return false
	}
  apiUrl := s.VerfiyUrl
	resource := "/api/v1/verfiy_token"
	data := url.Values{}
	data.Set("token", token)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String() // "https://xxxx.com/api/v1/verfiy_token"

	client := &http.Client{}
	r, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload

	if err != nil {
		log.Error(err)
		return false
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)

	defer func() {
    r.Close = true
		r.Body.Close();
		if resp != nil {
      resp.Body.Close();
		}
	}()

	if err != nil {
		log.Error(err)
		return false
	}

	log.WithField("status", resp.Status).Info("token验证结果...")
  if "204 No Content" == resp.Status {
		return true
	} else {
		return false
	}
}

// handleWebsocket connection.
func (s *Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	defer func() {
    r.Close = true
		r.Body.Close();
	}()

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

	// 查询提供的服务
	cloudWs, err := ProviderWsMapper(provider[0], apiName[0])
	if err != nil {
		log.WithField("err", err).Error("Create cloud ws conn failed...")
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	// Verify token from openai_backend
	if provider[0] != "local" {
    if s.VerifyToken(token[0]) {
			log.Info("路由参数解析成功并且token认证成功, 将升级为Websocket服务...")
		} else {
			http.Error(w, "Token is invalid", http.StatusForbidden)
			return
		}
	}
	
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m := "Unable to upgrade to websockets"
		log.WithField("err", err).Println(m)
		http.Error(w, m, http.StatusBadRequest)
		return
	}
	
	//注册client
  wsConn := NewWsConn(ws)
	cloudWsConn := NewWsConn(cloudWs)
	log.WithFields(log.Fields{"服务提供商": provider[0], "api": apiName[0]}).Info("创建云端服务提供商Websocket连接串成功")

	client := NewClient(wsConn, cloudWsConn, provider[0], apiName[0], token[0], ws.RemoteAddr().String())
	s.addClient(client)

	if provider[0] == "local" {
	  s.pool.Submit(client.RunEcho)
	} else {
    // 启动Client
	  s.pool.Submit(client.Run)
	}
	
}