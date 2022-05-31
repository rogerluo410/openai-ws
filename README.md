# openai-ws

代理服务供应商的WebSocket服务 和 Grpc服务 

  * 输入为结构化数据字节流, 数据结构参考具体服务供应商文档。  
  * 输出为结构化数据字节流, 数据结构参考具体服务供应商文档。  

:bullettrain_front:目前支持的服务供应商:  
   - 讯飞  
      [语音听写](https://www.xfyun.cn/doc/asr/voicedictation/API.html#%E6%8E%A5%E5%8F%A3%E8%B0%83%E7%94%A8%E6%B5%81%E7%A8%8B)     
      [语音合成](https://www.xfyun.cn/doc/tts/online_tts/API.html#%E6%8E%A5%E5%8F%A3%E8%B0%83%E7%94%A8%E6%B5%81%E7%A8%8B)    
      [语音评测](https://www.xfyun.cn/doc/Ise/IseAPI.html#%E6%8E%A5%E5%8F%A3%E8%AF%B4%E6%98%8E)  
      [实时语音转写](https://www.xfyun.cn/doc/asr/rtasr/API.html#%E6%8E%A5%E5%8F%A3%E8%AF%B4%E6%98%8E)  
   - 依图  
      [实时语音转写](https://speech.yitutech.com/devdoc/audio/liveaudio)    


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
    `cd test/xunfei_rtasr && ./xunfei_rtasr`  
    
    :worried:官方测试代码认证流程失败, [golang测试代码](https://xfyun-doc.cn-bj.ufileos.com/1536131421882586/rtasr_go_demo.zip)     
    <img width="1181" alt="image" src="https://user-images.githubusercontent.com/5260711/168715182-d36447ce-41e1-4739-8975-aa99c44a2c36.png">    

  - 依图平台 - 实时语音转写 -  测试通过 :white_check_mark:  
    `cd test/yitu_asr/real-time-asr-demo-python3-1203/real-time-demo && python real_time_asr_example.py`         

# 部署  
  `make install`    

# gRPC
  
  ### protoc 
  1. 安装grpc插件  
    `go get -u github.com/golang/protobuf/{proto,protoc-gen-go}`  

    `which protoc-gen-go`  -- 查看是否安装成功  

    `export PATH="$PATH:/$GOPATH/bin"`  -- 导出环境变量, protoc-gen-go会安装在$GOPATH/bin 中,  shell会找不到命令, 需要导出PATH.   
  

  2. proto生成golang代码   
    `protoc --go_out=plugins=grpc:. yitu_liveaudio.proto`   

# TODO

  - ~~用户OpenAI Backend认证~~  
  - ~~自动化部署~~  
  - ~~压力测试~~   
  - ~~APPKEY 配置化~~  
  - ~~优化 GRPC~~
