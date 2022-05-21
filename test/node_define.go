package test

import (
	"context"
	"math/rand"
	"time"

	"running"
)

type BaseTestNode struct {
	running.Base
}

func (node *BaseTestNode) Run(ctx context.Context) {
	start := time.Now()
	rand.Seed(start.Unix())
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	end := time.Now()

	AddLog(node.State, node.Name(), start, end, "", nil)
}

type SetStateNode struct {
	running.Base

	key string

	value interface{}
}

func (node *SetStateNode) Run(ctx context.Context) {
	node.State.Update(node.key, node.value)
	AddLog(node.State, node.Name(), time.Now(), time.Now(), "", nil)
}
