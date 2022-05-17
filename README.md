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
  
  - 讯飞平台 - 语音听写 - 测试通过 :white_check_mark:  
    `cd test/xunfei_vd && ./xunfei_vd -t xxxx`   
    
    or 
    `./xunfei_vd -t xxxx -a http://xxx.com` 指定token 和 Openai websocket代理服务地址  

  - 讯飞平台 - 语音合成 - 测试通过 :white_check_mark:  
    `cd test/xunfei_tts && ./xunfei_tts -t xxxx`   

  - 讯飞平台 - 语音评测 - 测试通过 :white_check_mark:  
    `cd test/xunfei_lse && ./xunfei_lse -t xxxx`   

  - 讯飞平台 - 实时语音转写 - 测试失败 :x:    
    [原因]官方测试代码认证流程失败: https://xfyun-doc.cn-bj.ufileos.com/1536131421882586/rtasr_go_demo.zip   
    


    `cd test/xunfei_rtasr && ./xunfei_rtasr`  

  - 依图平台 - 实时语音转写 -  测试通过 :white_check_mark:  
    `cd test/yitu_asr/real-time-asr-demo-python3-1203/real-time-demo && python real_time_asr_example.py`         

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