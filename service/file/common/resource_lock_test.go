package common

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestResourceLock(t *testing.T) {
	lockBucket := NewResourceLockBucket()
	wg := sync.WaitGroup{}
	imax := 0
	i := 0
	for loop := 0; loop < 5; loop++ {
		i = imax
		imax += rand.Intn(20)
		for ; i < imax; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				lock := lockBucket.NewResourceLockV2("config", "/path/this")
				defer lockBucket.ReleaseLock(lock)
				lock.Lock()
				defer lock.Unlock()
				fmt.Printf("run process [%d]\n", i)
				// do something
				time.Sleep(time.Second * time.Duration(1+rand.Float64()))
			}(i)
		}
		fmt.Printf("create goroutine [%d]\n", imax)
		// after a period of time a new request comes
		time.Sleep(time.Second * time.Duration(1+rand.Float64()))
	}
	wg.Wait()

}
