package process

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
	pool     *ants.Pool  
}

func NewProcessor() *Processor {
  if pool, err := ants.NewPool(poolNum); err != nil {
    // 抛出错误
    panic(err)
  } else {
    return &Processor{
      pool: pool,
    }
  }
  return nil
}

func (p *Processor) listen() {
  for {
    select {
      
    }
  }
}


func (p *Processor) HandleTask(pf func()) {
  p.pool.Submit(pf)
}


