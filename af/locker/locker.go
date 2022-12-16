package locker

import (
	"runtime"
	"sync"

	"github.com/pyddot/taos-driver-go-v2/wrapper/thread"
)

var locker *thread.Locker
var once = sync.Once{}

func Lock() {
	if locker == nil {
		SetMaxThreadSize(runtime.NumCPU())
	}
	locker.Lock()
}
func Unlock() {
	if locker == nil {
		SetMaxThreadSize(runtime.NumCPU())
	}
	locker.Unlock()
}

func SetMaxThreadSize(size int) {
	once.Do(func() {
		locker = thread.NewLocker(size)
	})
}
