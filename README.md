# openai-ws

代理服务提供商的WebSocket服务  

  * 输入为结构化数据字节流, 数据结构参考具体服务提供商文档。  
  * 输出为结构化数据字节流, 数据结构参考具体服务提供商文档。  

# 启动服务 

  - 运行  
  `./openai-ws`  or `./openai-ws -p 8081`  

  - 查看服务供应商提供的服务列表  
  `./openai-ws -m`  

  - 查看命令行参数  
  `./openai-ws -h`  

# API功能测试  
  
  - 讯飞 语音听写  
  `cd test && ./xufeiVoiceDictationTest` or `./xufeiVoiceDictationTest -t xxxx -a http://xxx.com` 指定token和代理ws服务地址  

# 部署  
  `make install`    

# Grpc
  
  ## protoc 
  1. 安装grpc插件  
  `go get -u github.com/golang/protobuf/{proto,protoc-gen-go}`  
  `which protoc-gen-go`  -- 查看是否安装成功 
  `export PATH="$PATH:/$GOPATH/bin"`  -- 导出环境变量, protoc-gen-go会安装在$GOPATH/bin 中,  shell会找不到命令, 需要导出PATH.  
  

  2. proto生成golang代码  
  `protoc --go_out=plugins=grpc:. yitu_liveaudio.proto`  

# TODO

  - ~~用户OpenAI Backend认证~~  
  - ~~自动化部署~~  
  - 压力测试 
  - APPKEY 配置化
  - 优化 GRPC