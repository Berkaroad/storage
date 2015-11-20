//
// 内存追加存储
// －－－－－－－－－－
// 此存储主要用于测试
//
package storage

import (
	"sync"
)

type MemoryAppendOnlyStore struct {
	isInitialized bool
	cache         *lockingInMemoryCache
}

func (self *MemoryAppendOnlyStore) InitFunc() interface{} {
	return func() {
		if !self.isInitialized {
			self.cache = &lockingInMemoryCache{cacheByKey: make(map[string][]DataWithKey, 0),
				cacheFull: []DataWithKey{},
				locker:    new(sync.Mutex)}
		}
		self.isInitialized = true
	}
}

func (self *MemoryAppendOnlyStore) Append(streamName string, data []byte, expectedStreamVersion int) error {
	return self.cache.ConcurrentAppend(streamName, data, func(version, storeVersion int) {
		// commit
	}, expectedStreamVersion)
}

func (self *MemoryAppendOnlyStore) ReadRecords(streamName string, startingFrom, maxCount int) []DataWithKey {
	return self.cache.ReadStream(streamName, startingFrom, maxCount)
}

func (self *MemoryAppendOnlyStore) ReadAllRecords(startingFrom, maxCount int) []DataWithKey {
	return self.cache.ReadAll(startingFrom, maxCount)
}

func (self *MemoryAppendOnlyStore) Close() {
	consoleLog.Printf("Close store.\n")

}

func (self *MemoryAppendOnlyStore) Reset() {
	consoleLog.Printf("Reset store.\n")
	self.cache.Clear(func() {})
}

func (self *MemoryAppendOnlyStore) GetCurrentVersion() int {
	consoleLog.Printf("Get current store version.\n")
	return self.cache.StoreVersion
}

type lockingInMemoryCache struct {
	StoreVersion int
	cacheByKey   map[string][]DataWithKey
	cacheFull    []DataWithKey
	locker       *sync.Mutex
}

func (self *lockingInMemoryCache) ConcurrentAppend(streamName string, data []byte, commit func(streamVersion, storeVersion int), expectedStreamVersion int) error {
	defer self.locker.Unlock()
	self.locker.Lock()

	dataList := self.cacheByKey[streamName]
	if dataList == nil {
		self.cacheByKey[streamName] = []DataWithKey{}
		dataList = self.cacheByKey[streamName]
	}

	actualStreamVersion := len(dataList)
	if expectedStreamVersion >= 0 {
		if actualStreamVersion != expectedStreamVersion {
			return &AppendOnlyStoreConcurrencyError{ExpectedStreamVersion: expectedStreamVersion, ActualStreamVersion: actualStreamVersion, StreamName: streamName}
		}
	}
	newStreamVersion := actualStreamVersion + 1
	newStoreVersion := self.StoreVersion + 1
	commit(newStreamVersion, newStoreVersion)

	// update in-memory cache only after real commit completed
	dataWithKey := DataWithKey{Key: streamName, Data: data, StreamVersion: newStreamVersion, StoreVersion: newStoreVersion}
	self.cacheFull = append(self.cacheFull, dataWithKey)
	dataList = append(dataList, dataWithKey)

	self.cacheByKey[streamName] = dataList
	self.StoreVersion = newStoreVersion
	consoleLog.Printf("Append to stream '%s' with v%d.\n", streamName, newStreamVersion)
	return nil
}

func (self *lockingInMemoryCache) ReadStream(streamName string, afterStreamVersion, maxCount int) []DataWithKey {
	if streamName == "" {
		panic("streamName is empty!")
	}
	if afterStreamVersion < 0 {
		panic("afterStreamVersion must be zero or greater.")
	}
	if maxCount <= 0 {
		panic("maxCount must be more than zero.")
	}

	// no lock is needed.
	filterDataList := []DataWithKey{}
	if dataList := self.cacheByKey[streamName]; dataList != nil && len(dataList) > 0 {
		for _, item := range dataList {
			if item.StreamVersion > afterStreamVersion {
				filterDataList = append(filterDataList, item)
			}
		}
		if maxCount < len(filterDataList) {
			filterDataList = filterDataList[:maxCount]
		}
	}
	consoleLog.Printf("Read stream '%s' after v%d, maxCount is %d.\n", streamName, afterStreamVersion, maxCount)
	return filterDataList
}

func (self *lockingInMemoryCache) ReadAll(afterStoreVersion, maxCount int) []DataWithKey {
	if afterStoreVersion < 0 {
		panic("afterStoreVersion must be zero or greater.")
	}
	if maxCount <= 0 {
		panic("maxCount must be more than zero.")
	}

	filterDataList := make([]DataWithKey, 0)
	if dataList := self.cacheFull; dataList != nil && len(dataList) > 0 {
		for _, item := range dataList {
			if item.StoreVersion > afterStoreVersion {
				filterDataList = append(filterDataList, item)
			}
		}
		if maxCount < len(filterDataList) {
			filterDataList = filterDataList[:maxCount]
		}
	}
	consoleLog.Printf("Read all after v%d, maxCount is %d.\n", afterStoreVersion, maxCount)
	return filterDataList
}

func (self *lockingInMemoryCache) Clear(executeWhenCommitting func()) {
	defer self.locker.Unlock()
	self.locker.Lock()

	executeWhenCommitting()
	self.cacheFull = []DataWithKey{}
	self.cacheByKey = make(map[string][]DataWithKey, 0)
	self.StoreVersion = 0
}
