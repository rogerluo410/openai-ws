package main

import (
  _ "context"
  _ "encoding/json"
  _ "fmt"
  _ "time"

	_ "github.com/sirupsen/logrus"
  "github.com/panjf2000/ants/v2"
)


var poolNum = 10

type Processor struct {
	pool        *ants.Pool
  SendMsg     chan *Job      // 处理器消息管道
}

func NewProcessor() *Processor {
  if pool, err := ants.NewPool(poolNum); err != nil {
    // 抛出错误
    panic(err)
  } else {
    return &Processor{
      pool: pool,
      SendMsg: make(chan *Job),
    }
  }
  return nil
}

func (p *Processor) listen() {
  for {
    select {
    case msg := <- p.SendMsg:
      msg.Client.SendMsg <- "我们收到了消息！"
    }
  }
}


func (p *Processor) HandleTask(pf func()) {
  p.pool.Submit(pf)
}


