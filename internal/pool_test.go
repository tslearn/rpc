package internal

import (
	"testing"
	"time"
)

func TestNewThreadPool(t *testing.T) {
	assert := NewRPCAssert(t)

	pool := newThreadPool(NewRPCProcessor(16, 16, nil, nil))
	assert(pool).IsNotNil()
	assert(pool.isRunning).IsTrue()
	assert(len(pool.threads)).Equals(numOfThreadPerThreadPool)
	assert(cap(pool.threads)).Equals(numOfThreadPerThreadPool)
	for i := 0; i < numOfThreadPerThreadPool; i++ {
		assert(pool.threads[i]).IsNotNil()
		assert(pool.threads[i].isRunning).IsTrue()
	}
	assert(pool.freeThreads).IsNotNil()
	assert(len(pool.freeThreads)).Equals(numOfThreadPerThreadPool)
	assert(cap(pool.freeThreads)).Equals(numOfThreadPerThreadPool)

	pool.stop()
}

func TestRpcThreadPool_stop(t *testing.T) {
	assert := NewRPCAssert(t)
	pool := newThreadPool(NewRPCProcessor(16, 16, nil, nil))
	assert(pool.stop()).Equals(true, []string{})
	assert(pool.freeThreads).IsNotNil()
	assert(pool.stop()).IsFalse()

	processor := NewRPCProcessor(16, 16, nil, nil)
	_ = processor.AddService(
		"user",
		NewService().Echo("sayHello", true, func(ctx *RPCContext) *RPCReturn {
			time.Sleep(99999999 * time.Second)
			return ctx.OK(true)
		}),
		GetStackString(0),
	)

	stream := NewRPCStream()
	stream.WriteString("$.user:sayHello")
	stream.WriteUint64(3)
	stream.WriteString("#")

	processor.Start()
	processor.PutStream(stream)
	processor.Stop()
}

func TestRpcThreadPool_allocThread(t *testing.T) {
	assert := NewRPCAssert(t)
	pool := newThreadPool(NewRPCProcessor(16, 16, nil, nil))
	assert(len(pool.freeThreads)).Equals(numOfThreadPerThreadPool)
	thread := pool.allocThread()
	assert(thread).IsNotNil()
	assert(len(pool.freeThreads)).Equals(numOfThreadPerThreadPool - 1)
	pool.freeThread(thread)
	pool.stop()
}

func TestRpcThreadPool_freeThread(t *testing.T) {
	assert := NewRPCAssert(t)
	pool := newThreadPool(NewRPCProcessor(16, 16, nil, nil))
	thread := pool.allocThread()
	assert(thread).IsNotNil()
	assert(len(pool.freeThreads)).Equals(numOfThreadPerThreadPool - 1)
	pool.freeThread(thread)
	assert(len(pool.freeThreads)).Equals(numOfThreadPerThreadPool)
	pool.stop()
}
