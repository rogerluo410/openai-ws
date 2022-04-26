package main

import (
	"net/http"
	"os"

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
  port := port
	if port == "" {
		log.WithField("PORT", port).Fatal("$PORT must be set")
	}

	// 启动服务
	log.Info("开启服务中...")
	serverInstance := server.NewServer()
	http.HandleFunc("/ws", serverInstance.HandleWebsocket)
	log.Info(http.ListenAndServe(":"+port, nil))
}