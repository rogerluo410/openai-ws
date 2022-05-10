# 可执行文件
EXEC="./openai-ws"
# 服务器地址
REMOTE="deploy@47.106.176.202:/srv/openai_ws"

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build

build-linux32:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build

build-linux64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

build-win32:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build

build-win64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build

install: build-linux64
	rsync -zP $(EXEC) $(REMOTE)

.PHONY: clean 
clean : 
	rm $(EXEC) 
