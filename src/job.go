package main

import (
  "encoding/json"
)

type Job struct {
  Client    *Client
  Sequence  int   // 序号
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