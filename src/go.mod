module github.com/rogerluo410/openai-ws

go 1.13

require (
	github.com/gorilla/websocket v1.5.0
	github.com/panjf2000/ants/v2 v2.4.8
	github.com/rogerluo410/openai-ws/src/server v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)

replace github.com/rogerluo410/openai-ws/src/server => ./server
