package server

import (
  _ "context"
  _ "encoding/json"
  _ "fmt"
  _ "time"

	_ "github.com/sirupsen/logrus"
  "github.com/panjf2000/ants/v2"
)


var poolNum = 50

type Processor struct {
	pool        *ants.PoolWithFunc
  SendMsg     chan *Job      // 处理器消息管道
}

func NewProcessor() *Processor {
  if pool, err := ants.NewPoolWithFunc(poolNum, func(i interface{}) {
		ProviderMapper(i)
	}); err != nil {
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

func (p *Processor) Listen() {
  for {
    select {
    case msg := <- p.SendMsg:
      msg.Client.SendMsg <- "我们收到了消息！"

      p.pool.Invoke(msg)
    }
  }
}


func (p *Processor) PoolRelease() {
  p.pool.Release()
}
