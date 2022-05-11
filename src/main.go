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
	version = "1.0.3"
	maxActiveClientCnt = 10000
	port = GetEnvDefault("OAWS_PORT", "8080")
	openaiBackendUrl = GetEnvDefault("OPENAI_BACKEND_URL", "http://localhost:3001")
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

	var maxFlag uint
	flag.UintVar(&maxFlag, "c", uint(maxActiveClientCnt), "最大客户连接数")

	mapFlag := flag.Bool("m", false, "打印服务及API列表")
	helpFlag := flag.Bool("h", false, "使用帮助")
	versionFlag := flag.Bool("v", false, "当前版本")

	flag.Parse()

	if *mapFlag {
		fmt.Println("服务及API列表:")
		server.MapDict()
		os.Exit(0)
	}

	var usage = `使用: openai-ws [options...]

		Options:
			-p  指定服务端口, 默认为8080
			-t  认证服务器地址, 默认为http://localhost:3001
			-c  设置最大客户连接数, 默认最大10000
			-m  打印服务及API名列表
			-v  当前版本号
			-h  使用帮助文档
	`

	if *helpFlag {
		fmt.Println(usage)
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	// 启动服务
	log.WithFields(log.Fields{"Openai Ws Port": portFlag, "Openai Backend Url": tokenFlag}).Info("启动服务...")
	serverInstance := server.NewServer(tokenFlag, maxFlag)
	ctx, _:=context.WithCancel(context.Background())
	serverInstance.Listen(ctx)

	http.HandleFunc("/ws", serverInstance.HandleWebsocket)
	log.Info(http.ListenAndServe(":"+portFlag, nil))
}