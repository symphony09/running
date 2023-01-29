package common

import (
	"context"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/symphony09/running"
)

type DebugWrapper struct {
	running.BaseWrapper

	Keys []string

	logger *log.Logger
}

func NewDebugWrapper(name string, props running.Props) (running.Node, error) {
	wrapper := new(DebugWrapper)
	wrapper.logger = log.New(os.Stdout, "[RUNNING DEBUG] ", log.LstdFlags)

	keys, _ := props.SubGet(name, "debug")
	if keysStr, ok := keys.(string); ok {
		wrapper.Keys = strings.Split(keysStr, ",")
	}

	for i, key := range wrapper.Keys {
		wrapper.Keys[i] = strings.TrimSpace(key)
	}

	return wrapper, nil
}

func (wrapper *DebugWrapper) Bind(state running.State) {
	defer func() {
		if r := recover(); r != nil {
			wrapper.logger.Printf("node panic when bind state, node name: %s, panic info: %v\n stacktrace:\n %s\n",
				wrapper.Target.Name(), r, string(debug.Stack()))

			panic(r)
		}
	}()

	wrapper.BaseWrapper.Bind(state)
}

func (wrapper *DebugWrapper) Run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			wrapper.logger.Printf("node panic when running, node name: %s, panic info: %v\n stacktrace:\n %s\n",
				wrapper.Target.Name(), r, string(debug.Stack()))

			panic(r)
		}
	}()

	wrapper.debug(ctx, true)

	wrapper.logger.Printf("%s is start running.\n", wrapper.Target.Name())

	start := time.Now()

	wrapper.Target.Run(ctx)

	wrapper.logger.Printf("%s is completed, cost %s.\n", wrapper.Target.Name(), time.Since(start).String())

	wrapper.debug(ctx, false)
}

func (wrapper *DebugWrapper) debug(ctx context.Context, before bool) {
	const (
		FlagsCtx = 1 << iota
		FlagsStatesBefore
		FlagsStatesAfter
	)

	for _, key := range wrapper.Keys {
		var flags int

		if strings.HasPrefix(key, "ctx:") {
			key = strings.TrimPrefix(key, "ctx:")
			flags |= FlagsCtx
		} else if strings.HasPrefix(key, "state:") {
			key = strings.TrimPrefix(key, "state:")
			flags |= FlagsStatesBefore | FlagsStatesAfter
		} else if strings.HasPrefix(key, "state_in:") {
			key = strings.TrimPrefix(key, "state_in:")
			flags |= FlagsStatesBefore
		} else if strings.HasPrefix(key, "state_out:") {
			key = strings.TrimPrefix(key, "state_out:")
			flags |= FlagsStatesAfter
		} else {
			flags |= FlagsCtx | FlagsStatesBefore | FlagsStatesAfter
		}

		if before {
			flags &= 0xF ^ FlagsStatesAfter
		} else {
			flags &= 0xF ^ FlagsCtx ^ FlagsStatesBefore
		}

		if flags&FlagsCtx == FlagsCtx {
			if v := ctx.Value(key); v != nil {
				wrapper.logger.Printf("found %s in context of %s, type = %T\tvalue = %v\n",
					key, wrapper.Target.Name(), v, v)
			} else {
				wrapper.logger.Printf("%s not found in context of %s",
					key, wrapper.Target.Name())
			}
		}

		if flags&FlagsStatesBefore == FlagsStatesBefore {
			if v, ok := wrapper.State.Query(key); ok {
				if v != nil {
					wrapper.logger.Printf("found %s in state(before) of %s, type = %T\tvalue = %v\n",
						key, wrapper.Target.Name(), v, v)
				} else {
					wrapper.logger.Printf("%s in state(before) of %s had been set to nil\n", key, wrapper.Target.Name())
				}

			} else {
				wrapper.logger.Printf("%s not found in state(before) of %s",
					key, wrapper.Target.Name())
			}
		}

		if flags&FlagsStatesAfter == FlagsStatesAfter {
			if v, ok := wrapper.State.Query(key); ok {
				if v != nil {
					wrapper.logger.Printf("found %s in state(after) of %s, type = %T\tvalue = %v\n",
						key, wrapper.Target.Name(), v, v)
				} else {
					wrapper.logger.Printf("%s in state(after) of %s had been set to nil\n", key, wrapper.Target.Name())
				}
			} else {
				wrapper.logger.Printf("%s not found in state(after) of %s\n",
					key, wrapper.Target.Name())
			}
		}
	}
}

func (wrapper *DebugWrapper) Reset() {
	defer func() {
		if r := recover(); r != nil {
			wrapper.logger.Printf("node panic when reset, node name: %s, panic info: %v\n stacktrace:\n %s\n",
				wrapper.Target.Name(), r, string(debug.Stack()))

			panic(r)
		}
	}()

	wrapper.BaseWrapper.Reset()
}
