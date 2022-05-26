# 压测  

### 客户端错误集  

  - dial tcp: lookup localhost: no such host  

    * ulimit -n  查看文件描述符最大限制    
    * fd数量超过了最大限制   
    * ulimit -n 20000  调整最大限制超过预设的客户端数  

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


  - 系统线程 / golang runtime 协程 M:N 设置  
 