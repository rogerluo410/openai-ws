package main

import (
	"net/http"
	"os"
	"context"
	"flag"
	"fmt"
	"sync"
	"runtime"
	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"

	server "github.com/rogerluo410/openai-ws/src/server"
	grpc "github.com/rogerluo410/openai-ws/src/grpc"
)

var (
	version = "1.0.3"
	maxActiveClientCnt = 10000
	port = GetEnvDefault("OAWS_PORT", "8080")
	grpcPort = GetEnvDefault("OAGRPC_PORT", "8090")
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

	var grpcPortFlag string
  flag.StringVar(&grpcPortFlag, "g",  grpcPort, "Grpc服务监听端口")

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
			-p  指定Websocket服务端口, 默认为8080
			-g  指定Grpc服务端口, 默认为8090
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

	log.Infof("Cpu num: %d\n", runtime.NumCPU()) // 8核cpu

	// 虽然goroutine是并发执行的，但是它们并不是并行运行的。如果不告诉Go额外的东西，同
	// 一时刻只会有一个goroutine执行。利用runtime.GOMAXPROCS(n)可以设置goroutine
	// 并行执行的数量。GOMAXPROCS 设置了同时运行的CPU 的最大数量，并返回之前的设置。
	val := runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	log.Infof("Last goroutine num: %d, latest goroutine num: %d \n", val, runtime.NumCPU() * 4) // 8个

	var wg = sync.WaitGroup{}
	wg.Add(2)

	// 启动服务
	go func() {
    log.WithFields(log.Fields{"Openai Ws Port": portFlag, "Openai Backend Url": tokenFlag}).Info("Websocket服务启动...")
		serverInstance := server.NewServer(tokenFlag, maxFlag)
		ctx, _:=context.WithCancel(context.Background())
		serverInstance.Listen(ctx)

		defer func() {
      serverInstance.Close()
			wg.Done()
		}() 

		http.HandleFunc("/ws", serverInstance.HandleWebsocket)
		err := http.ListenAndServe(":"+portFlag, nil)
		if err != nil {
      log.WithField("Err", err).Fatalf("Failed to start up websocket server listening on %s", portFlag)
		}
	}()

	go func() {
		defer wg.Done()
		log.WithFields(log.Fields{"Openai Grpc Port": grpcPortFlag, "Openai Backend Url": tokenFlag}).Info("Grpc服务启动...")
		grpc.StartGrpcServer(grpcPortFlag, tokenFlag)
	}()
	
	wg.Wait()
	log.Info("程序将退出...")
	os.Exit(0)
}