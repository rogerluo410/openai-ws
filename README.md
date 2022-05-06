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

# TODO

  - ~~用户OpenAI Backend认证~~  
  - ~~自动化部署~~  
  - 压力测试 