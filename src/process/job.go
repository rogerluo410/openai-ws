package process

import (
  "encoding/json"

  client "github.com/rogerluo410/openai-ws/src/client"
)

type Job struct {
  Client    *client.Client
  sequence  int   // 序号
	Type      int   //0: 输入消息； 1: 输出消息
  Message   string
}

// 序列化
func (j *Job) Bytes() []byte {
  b, err := json.Marshal(j)
  if err != nil {
    panic(err)
  }
  return b
}