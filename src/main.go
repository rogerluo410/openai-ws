package main

import (
	"net/http"
	"os"
	"context"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"

	server "github.com/rogerluo410/openai-ws/src/server"
)

var (
	port = GetEnvDefault("OAWS_PORT", "8080")
	openaiBackendUrl = GetEnvDefault("OPENAI_BACKEND_URL", "localhost:3001")
)

func GetEnvDefault(key, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

func main() {
  var portFlag string
  flag.StringVar(&portFlag, "p",  port, "Websocket服务监听端口")

	var tokenFlag string
  flag.StringVar(&tokenFlag, "t",  openaiBackendUrl, "token认证服务地址")

	mapFlag := flag.Bool("m", false, "打印服务及API名")

	flag.Parse()

	if *mapFlag {
		fmt.Println("打印服务及API名:")
		server.MapDict()
		return
	}

	// 启动服务
	log.WithFields(log.Fields{"Openai Ws Port": portFlag, "Openai Backend Url": tokenFlag}).Info("启动服务...")
	serverInstance := server.NewServer()
	ctx, _:=context.WithCancel(context.Background())
	serverInstance.Listen(ctx)

	http.HandleFunc("/ws", serverInstance.HandleWebsocket)
	log.Info(http.ListenAndServe(":"+portFlag, nil))
}