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

func New() *Processor {
  if pool, err := ants.NewPool(poolNum); err != nil {
    // 抛出错误
  } else {
    return &Processor{
      pool: pool,
    }
  }
  return nil
}


func (p *Processor) Handle(job *Job) {
  // p.pool.Submit()
}


