// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// see: "github.com/beego/beego/v2/client/cache"

package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	// clock time of recycling the expired cache items in memory.
	DefaultEvery int = 60 // 1 minute
)

// Memory cache item.
type MemoryItem struct {
	val             interface{}
	Lastaccess      time.Time
	expired         int64
	expiredCallBack func(item interface{}) error
}

func (m *MemoryItem) RegisterExpiredCallback(fn func(m interface{}) error) {
	m.expiredCallBack = fn
}

// Memory cache adapter.
// it contains a RW locker for safe map storage.
type MemoryCache struct {
	lock  sync.RWMutex
	dur   time.Duration
	items map[string]*MemoryItem
	Every int // run an expiration check Every clock time
}

// NewMemoryCache returns a new MemoryCache.
func NewMemoryCache(autoGC bool) *MemoryCache {
	cache := MemoryCache{items: make(map[string]*MemoryItem)}
	if autoGC {
		cache.StartAndGC(0)
	}
	return &cache
}

func (bc *MemoryCache) ExpiredCallback(name string) error {
	if bc.items[name].expiredCallBack == nil {
		return nil
	}
	return bc.items[name].expiredCallBack(bc.items[name])
}

func (bc *MemoryCache) ThrowError(str string) error {
	log.Error(str)
	return errors.New(str)
}

// Get cache from memory.
// if non-existed or expired, return nil.
func (bc *MemoryCache) Get(name string) interface{} {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	if itm, ok := bc.items[name]; ok {
		if (time.Now().Unix() - itm.Lastaccess.Unix()) > itm.expired {
			if err := bc.ExpiredCallback(name); err != nil {
				log.Debugf("get from cache:%s success", name)
				return bc.items[name]
			}
			go bc.Delete(name)
			return nil
		}
		log.Debugf("get from cache:%s success", name)
		return itm.val
	}
	log.Debugf("get miss from cache:%s", name)
	return nil
}

// GetMulti gets caches from memory.
// if non-existed or expired, return nil.
func (bc *MemoryCache) GetMulti(names []string) []interface{} {
	var rc []interface{}
	for _, name := range names {
		rc = append(rc, bc.Get(name))
	}
	return rc
}

// Put cache to memory.
// if expired is 0, it will be cleaned by next gc operation ( default gc clock is 1 minute).
func (bc *MemoryCache) Put(name string, value interface{}, expired int64) *MemoryItem {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	/*
		if v, ok := bc.items[name]; ok {
			bc.items[name].val = value
			bc.items[name].Lastaccess = time.Now()
		}*/

	bc.items[name] = &MemoryItem{
		val:        value,
		Lastaccess: time.Now(),
		expired:    expired,
	}
	return bc.items[name]
}

/// Delete cache in memory.
func (bc *MemoryCache) Delete(name string) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	if _, ok := bc.items[name]; !ok {
		return bc.ThrowError(fmt.Sprintln("key not exist", name))
	}

	//if bc.items[name].expiredCallBack != nil {
	err := bc.ExpiredCallback(name)
	if err != nil {
		return err
	}
	log.Debugf("delete from cache:%v", name)
	//}
	//fmt.Println("aaaaaaaaaaaa", err)
	delete(bc.items, name)
	if _, ok := bc.items[name]; ok {
		return bc.ThrowError(fmt.Sprintln("delete cache error", name))
	}
	return nil
}

// Increase cache counter in memory.
// it supports int,int64,int32,uint,uint64,uint32.
func (bc *MemoryCache) Incr(key string) error {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	itm, ok := bc.items[key]
	if !ok {
		return bc.ThrowError(fmt.Sprintln("incr key not exists", key))
	}
	switch itm.val.(type) {
	case int:
		itm.val = itm.val.(int) + 1
	case int64:
		itm.val = itm.val.(int64) + 1
	case int32:
		itm.val = itm.val.(int32) + 1
	case uint:
		itm.val = itm.val.(uint) + 1
	case uint32:
		itm.val = itm.val.(uint32) + 1
	case uint64:
		itm.val = itm.val.(uint64) + 1
	default:
		return bc.ThrowError(fmt.Sprintln("item val is not int int64 int32", itm.val))
	}
	return nil
}

// Decrease counter in memory.
func (bc *MemoryCache) Decr(key string) error {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	itm, ok := bc.items[key]
	if !ok {
		return bc.ThrowError(fmt.Sprintln("decr key not exists", key))
	}
	switch itm.val.(type) {
	case int:
		itm.val = itm.val.(int) - 1
	case int64:
		itm.val = itm.val.(int64) - 1
	case int32:
		itm.val = itm.val.(int32) - 1
	case uint:
		if itm.val.(uint) > 0 {
			itm.val = itm.val.(uint) - 1
		} else {
			return bc.ThrowError(fmt.Sprintln("item val is less than 0", key))
		}
	case uint32:
		if itm.val.(uint32) > 0 {
			itm.val = itm.val.(uint32) - 1
		} else {
			return bc.ThrowError(fmt.Sprintln("item val is less than 0", key))
		}
	case uint64:
		if itm.val.(uint64) > 0 {
			itm.val = itm.val.(uint64) - 1
		} else {
			return bc.ThrowError(fmt.Sprintln("item val is less than 0", key))
		}
	default:
		return bc.ThrowError(fmt.Sprintln("item val is not int int64 int32", key))
	}
	return nil
}

// check cache exist in memory.
func (bc *MemoryCache) IsExist(name string) bool {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	_, ok := bc.items[name]
	return ok
}

// delete all cache in memory.
func (bc *MemoryCache) ClearAll() error {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	bc.items = make(map[string]*MemoryItem)
	return nil
}

// get all cache in memory.
func (bc *MemoryCache) GetAll() []interface{} {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	var out []interface{}
	out = make([]interface{}, 100)
	i := 0
	for _, val := range bc.items {
		out[i] = val.val
		i++
	}
	return out
}

func (bc *MemoryCache) GetAllEx() map[string]interface{} {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	out := make(map[string]interface{})
	for key, val := range bc.items {
		out[key] = val.val
	}
	return out
}

// start memory cache. it will check expiration in every clock time.
func (bc *MemoryCache) StartAndGC(interval int) error {
	if interval <= 0 {
		interval = DefaultEvery
	}
	dur, err := time.ParseDuration(fmt.Sprintf("%ds", interval))
	if err != nil {
		return err
	}
	bc.Every = interval
	bc.dur = dur
	go bc.vaccuum()
	return nil
}

// check expiration.
func (bc *MemoryCache) vaccuum() {
	if bc.Every < 1 {
		return
	}
	for {
		<-time.After(bc.dur)
		//fmt.Println("gc")
		if bc.items == nil {
			return
		}
		for name := range bc.items {
			bc.item_expired(name)
		}
	}
}

// item_expired returns true if an item is expired.
func (bc *MemoryCache) item_expired(name string) bool {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	itm, ok := bc.items[name]
	if !ok {
		return true
	}
	if time.Now().Unix()-itm.Lastaccess.Unix() >= itm.expired {
		if err := bc.ExpiredCallback(name); err != nil {
			return false
		}
		log.Debug("delete from cache by gc", name)
		delete(bc.items, name)
		return true
	}
	return false
}
