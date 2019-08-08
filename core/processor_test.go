package core

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var service = newServiceMeta().
	Echo("sayHello", true, func(
		ctx Context,
		name RPCString,
	) Return {
		n, _ := name.ToString()
		if n == "true" {
			return ctx.OK(true)
		} else {
			return ctx.Error()
		}
	}).
	Echo("sayGoodBye", true, func(
		ctx Context,
		name RPCString,
	) Return {
		return nil
	})

func getProcessor() *rpcProcessor {
	logger := NewLogger()
	return newProcessor(logger, 16, 16)
}

func TestRpcProcessor_Execute(t *testing.T) {
	processor := getProcessor()
	processor.start()
	processor.AddService("user", service)

	stream := NewRPCStream()
	processor.put(stream, nil)

	time.Sleep(time.Second)
	processor.stop()
}

func BenchmarkRpcProcessor_Execute(b *testing.B) {
	processor := getProcessor()
	processor.start()
	processor.AddService("user", service)
	time.Sleep(2 * time.Second)
	stream := NewRPCStream()

	n := 5

	finish := make(chan bool, n)

	speedCount.Calculate()

	for i := 0; i < n; i++ {
		go func() {
			r := rand.New(rand.NewSource(rand.Int63()))

			for j := 0; j < 10000000; j++ {
				processor.put(stream, r)
			}

			finish <- true
		}()
	}

	for i := 0; i < n; i++ {
		<-finish
	}

	fmt.Println("Speed:", speedCount.Calculate())
}
