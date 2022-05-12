module github.com/rogerluo410/openai-ws

go 1.13

require (
	github.com/fatih/structs v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0
	github.com/panjf2000/ants/v2 v2.4.8
	github.com/rogerluo410/openai-ws/src/grpc v0.0.0-00010101000000-000000000000
	github.com/rogerluo410/openai-ws/src/server v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/grpc v1.46.0
	google.golang.org/grpc/examples v0.0.0-20220510235641-db79903af928
	google.golang.org/protobuf v1.28.0 // indirect
)

replace github.com/rogerluo410/openai-ws/src/server => ./server

replace github.com/rogerluo410/openai-ws/src/grpc => ./grpc
