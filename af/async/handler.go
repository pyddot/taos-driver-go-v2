package async

import (
	"sync"

	"github.com/pyddot/taos-driver-go-v2/wrapper/handler"
)

const defaultPoolSize = 10000

var HandlerPool *handler.HandlerPool
var once = sync.Once{}

func SetHandlerSize(size int) {
	once.Do(func() {
		HandlerPool = handler.NewHandlerPool(size)
	})
}

func GetHandler() *handler.Handler {
	if HandlerPool == nil {
		SetHandlerSize(defaultPoolSize)
	}
	return HandlerPool.Get()
}

func PutHandler(h *handler.Handler) {
	if HandlerPool == nil {
		return
	}
	HandlerPool.Put(h)
}
