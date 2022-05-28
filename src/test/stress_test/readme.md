# 压测FAQ  

### 客户端错误集  

  - dial tcp: lookup localhost: no such host  

    * ulimit -n  查看文件描述符最大限制    
    * fd数量超过了最大限制   
    * ulimit -n 20000  调整最大限制超过预设的客户端数  

    * 但是这个只能临时修改，具体永久修改方法不在这里说明，文件是/etc/security/limits.conf  

    * on MacOS   
      `launchctl limit maxfiles`   -- show maximum fd limits  
      `sudo launchctl limit maxfiles 65536 200000`   -- The first number is the “soft” limit and the second number is the “hard” limit.    

    * 为什么 ulimit -n 20000 没有效果？   

      change the ulimit to avoid the error "too many open files" by default max ulimit is 4096 for linux and 1024 for mac, u can change ulimit to 4096 by typing ulimit -n 4096 for beyond 4096 you need to modify limits.conf in /etc/security folder for linux and set hard limit to 100000 by adding this line "* hard core 100000"    
 

  - 单机最大客户端数10000, 单机1000多个连接时, 客户端panic: unexpected EOF
    
    服务器崩了导致...  


  - 单机最大客户端数10000, 单机1000多个连接时, 客户端read: connection reset by peer, 服务器主动关闭了连接   
    
    可能是客户端握手超时, 之前设置5秒没握上手就超时，会报该错误， 设置为60秒就没问题了，websocket 默认握手超时就是60秒。  

    ```golang
      d := websocket.Dialer{
        HandshakeTimeout: 60 * time.Second, // 5秒握手超时，太短了
      }
    ```

  - 客户端信息panic: dial tcp [::1]:8080: connect: connection refused
    
    服务器崩了导致...  


  - 客户端单机最大客户端数10000,  http: Accept error: accept tcp [::]:8080: accept: too many open files; retrying in 160ms     

    文件描述符不够导致...  


  - 单机最大客户端数1000, 服务端 w.Conn.WriteMessage(websocket.PongMessage, []byte{})出错：  
      
    * read tcp [::1]:8080->[::1]:51983: use of closed network connection  

    * concurrent write to websocket connection  

      gorilla/websocket 官方文档说 Connections support one concurrent reader and one concurrent writer.  一个连接只支持一个并发读和一个并发写。  

      我其实只用了一个协程去写数据，但是在一个协程中，有多处写数据，包括(websocket.CloseMessage / websocket.PongMessage / websocket.CloseAbnormalClosure / websocket.TextMessage ), 估计是WriteMessage底层触发了竞态条件。  
      改用WriteControl去做(websocket.CloseMessage / websocket.PongMessage / websocket.CloseAbnormalClosure) 就没panic这个错误了。  
   
      github.com/gorilla/websocket  WriteMessage 不是线程安全的， 需要自己加锁互斥。 

        https://github.com/gorilla/websocket/issues/698  -- Use a lock per connection to ensure that the application does not write concurrently to a connection.   
        
        https://github.com/gorilla/websocket/issues/652  -- 使用 WriteControl  
    
    * Data Race Detector 竞态条件检测  
    
      https://go.dev/doc/articles/race_detector  

  
  - 单机最大客户端数10000, 创建5000多个连接时, 客户端报read tcp [::1]:55290->[::1]:8080: i/o timeout  

  - 引入协程池后, 单机最大客户端数10000， 客户端报write tcp [::1]:63518->[::1]:8080: write: broken pipe
  
    Usually, you get the broken pipe error when you write to the connection after the RST is sent, and when you read from the connection after the RST instead, you get the connection reset by peer error.    

  - 系统线程 / golang runtime 协程 M:N 设置  
 

  - append 不是线程安全的  
   
     https://www.fushengwushi.com/archives/1480  

    ```golang
      func (s *Server) addClient(c *Client) {
        s.lock.Lock()
        defer s.lock.Unlock()
        s.list = append(s.list, c)
      }

      func (s *Server) removeClients() {
        s.lock.Lock()
        defer s.lock.Unlock()
        for index, client := range s.list {
          if client.Actived == false {
            s.list = append(s.list[:index], s.list[index+1:]...)
          }
        }
      }

      // 加了锁， 还是会panic: panic: runtime error: slice bounds out of range [6:4]   
      // 加了锁， 还是报数组访问越界的panic    
      // 这就是因为线程不安全导致的 (实际业务中，可以通过 go run -race main.go 进行检测程序的安全性) 
    ```

    Solution:  

    ```golang
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
    ```

  - 客户端 read message error: websocket: close 1005 (no status)  

    ```golang
      // Close codes defined in RFC 6455, section 11.7.
      const (
        CloseNormalClosure           = 1000
        CloseGoingAway               = 1001
        CloseProtocolError           = 1002
        CloseUnsupportedData         = 1003
        CloseNoStatusReceived        = 1005
        CloseAbnormalClosure         = 1006
        CloseInvalidFramePayloadData = 1007
        ClosePolicyViolation         = 1008
        CloseMessageTooBig           = 1009
        CloseMandatoryExtension      = 1010
        CloseInternalServerErr       = 1011
        CloseServiceRestart          = 1012
        CloseTryAgainLater           = 1013
        CloseTLSHandshake            = 1015
      )

      if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
          log.Printf("error: %v", err)
      }
    ```
    
    客户端设置为每20秒发送一次消息:  
    ```golang
      // 每20秒发送一次消息
			case <- time.After(20 * time.Second):  
    ```  

    服务端读写等待时间是60秒:  
    ```golang
      w.Conn.SetReadDeadline(time.Now().Add(pongWait))  
      w.Conn.SetWriteDeadline(time.Now().Add(writeWait))    
    ```

    w.Conn.SetReadDeadline(time.Now().Add(readWait)) 设置错了地方  

    应该每次读写之前设置超时最后期限， 我理解为了Timeout(每次读等待多少秒后超时)， 在循环外面设置了一个总的超时。。。  

    ```golang
      func (w *WsConn) Reader(client *Client, ctx context.Context, cancelFunc context.CancelFunc) {
        defer func() {
          log.Info("Client - Reader 协程退出...")
          w.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
          client.Wg.Done()
        }()

        w.Conn.SetReadLimit(1024 * 1024)

        // 之前在这里设置超时最后期限是当前时间+60秒， 导致60秒后读毕定超时， 
        // 客户端接收到1005错误  
        // SetReadDeadline的行为是超时后conn连接状态是被破坏的, 所有读操作返回错误  
        // SetReadDeadline sets the read deadline on the underlying network connection. After a read has timed out, the websocket connection state is corrupt and all future reads will return an error. A zero value for t means reads will not time out.   
        // w.Conn.SetReadDeadline(time.Now().Add(readWait))    

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
    ```

  - 查看进程占用资源  

  ```shell
    ps aux | grep openai-ws
    top -pid 53578  
  ```