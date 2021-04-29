package common

import (
	"path/filepath"
	"sync"
)

type ResourceLockBucket struct {
	lockBucket map[string]*resourceLock
	mtx        sync.Mutex
}

type resourceLock struct {
	rscTag string
	mtx    sync.Mutex
	refNum int
}

func NewResourceLockBucket() *ResourceLockBucket {
	return &ResourceLockBucket{
		lockBucket: make(map[string]*resourceLock),
		mtx:        sync.Mutex{},
	}
}

func (r *ResourceLockBucket) checkExist(tag string) (*resourceLock, bool) {
	if oldLock, ok := r.lockBucket[tag]; ok {
		return oldLock, true
	}
	return nil, false
}

func (r *ResourceLockBucket) NewResourceLock(tag string) *resourceLock {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if oldLock, ok := r.checkExist(tag); ok {
		oldLock.refNum += 1
		return oldLock
	}
	lock := &resourceLock{
		rscTag: tag,
		refNum: 1,
	}
	lock.mtx = sync.Mutex{}
	r.lockBucket[lock.rscTag] = lock
	return lock
}

func (r *ResourceLockBucket) NewResourceLockV2(path ...string) *resourceLock {
	return r.NewResourceLock(genResourceTag(path...))
}

func (r *ResourceLockBucket) ReleaseLock(lock *resourceLock) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if lock.refNum > 0 {
		lock.refNum -= 1
	} else {
		delete(r.lockBucket, lock.rscTag)
	}
}

func (r *resourceLock) Lock() {
	r.mtx.Lock()
}

func (r *resourceLock) Unlock() {
	r.mtx.Unlock()
}

func genResourceTag(str ...string) string {
	var tag string
	for _, s := range str {
		tag = filepath.Join(tag, "/"+s)
	}
	return tag
}
