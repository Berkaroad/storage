package storage

import (
	"sync"
)

// MemoryAppendOnlyStore is used for test.
// For simulate appendonlystore.
type MemoryAppendOnlyStore struct {
	isInitialized bool
	cache         *lockingInMemoryCache
}

func (store *MemoryAppendOnlyStore) InitFunc() interface{} {
	return func() {
		if !store.isInitialized {
			store.cache = &lockingInMemoryCache{cacheByKey: make(map[string][]DataWithKey, 0),
				cacheFull: []DataWithKey{},
				locker:    new(sync.Mutex)}
		}
		store.isInitialized = true
	}
}

func (store *MemoryAppendOnlyStore) Append(streamName string, data []byte, expectedStreamVersion int) error {
	return store.cache.ConcurrentAppend(streamName, data, func(version, storeVersion int) {
		// commit
	}, expectedStreamVersion)
}

func (store *MemoryAppendOnlyStore) ReadRecords(streamName string, startingFrom, maxCount int) []DataWithKey {
	return store.cache.ReadStream(streamName, startingFrom, maxCount)
}

func (store *MemoryAppendOnlyStore) ReadAllRecords(startingFrom, maxCount int) []DataWithKey {
	return store.cache.ReadAll(startingFrom, maxCount)
}

func (store *MemoryAppendOnlyStore) Close() {
	if DEBUG {
		consoleLog.Printf("Close store.\n")
	}
}

func (store *MemoryAppendOnlyStore) Reset() {
	if DEBUG {
		consoleLog.Printf("Reset store.\n")
	}
	store.cache.Clear(func() {})
}

func (store *MemoryAppendOnlyStore) GetCurrentVersion() int {
	if DEBUG {
		consoleLog.Printf("Get current store version.\n")
	}
	return store.cache.StoreVersion
}

type lockingInMemoryCache struct {
	StoreVersion int
	cacheByKey   map[string][]DataWithKey
	cacheFull    []DataWithKey
	locker       *sync.Mutex
}

func (cache *lockingInMemoryCache) ConcurrentAppend(streamName string, data []byte, commit func(streamVersion, storeVersion int), expectedStreamVersion int) error {
	defer cache.locker.Unlock()
	cache.locker.Lock()

	dataList := cache.cacheByKey[streamName]
	if dataList == nil {
		cache.cacheByKey[streamName] = []DataWithKey{}
		dataList = cache.cacheByKey[streamName]
	}

	actualStreamVersion := len(dataList)
	if expectedStreamVersion >= 0 {
		if actualStreamVersion != expectedStreamVersion {
			return &AppendOnlyStoreConcurrencyError{ExpectedStreamVersion: expectedStreamVersion, ActualStreamVersion: actualStreamVersion, StreamName: streamName}
		}
	}
	newStreamVersion := actualStreamVersion + 1
	newStoreVersion := cache.StoreVersion + 1
	commit(newStreamVersion, newStoreVersion)

	// update in-memory cache only after real commit completed
	dataWithKey := DataWithKey{Key: streamName, Data: data, StreamVersion: newStreamVersion, StoreVersion: newStoreVersion}
	cache.cacheFull = append(cache.cacheFull, dataWithKey)
	dataList = append(dataList, dataWithKey)

	cache.cacheByKey[streamName] = dataList
	cache.StoreVersion = newStoreVersion
	if DEBUG {
		consoleLog.Printf("Append to stream '%s' with v%d.\n", streamName, newStreamVersion)
	}
	return nil
}

func (cache *lockingInMemoryCache) ReadStream(streamName string, afterStreamVersion, maxCount int) []DataWithKey {
	if streamName == "" {
		panic("streamName is empty!")
	}
	if afterStreamVersion < 0 {
		panic("afterStreamVersion must be zero or greater.")
	}
	if maxCount <= 0 {
		panic("maxCount must be more than zero.")
	}

	defer cache.locker.Unlock()
	cache.locker.Lock()
	filterDataList := []DataWithKey{}
	if dataList := cache.cacheByKey[streamName]; dataList != nil && len(dataList) > 0 {
		for _, item := range dataList {
			if item.StreamVersion > afterStreamVersion {
				filterDataList = append(filterDataList, item)
			}
		}
		if maxCount < len(filterDataList) {
			filterDataList = filterDataList[:maxCount]
		}
	}
	if DEBUG {
		consoleLog.Printf("Read stream '%s' after v%d, maxCount is %d.\n", streamName, afterStreamVersion, maxCount)
	}
	return filterDataList
}

func (cache *lockingInMemoryCache) ReadAll(afterStoreVersion, maxCount int) []DataWithKey {
	if afterStoreVersion < 0 {
		panic("afterStoreVersion must be zero or greater.")
	}
	if maxCount <= 0 {
		panic("maxCount must be more than zero.")
	}

	defer cache.locker.Unlock()
	cache.locker.Lock()
	var filterDataList = make([]DataWithKey, 0)
	if dataList := cache.cacheFull; dataList != nil && len(dataList) > 0 {
		for _, item := range dataList {
			if item.StoreVersion > afterStoreVersion {
				filterDataList = append(filterDataList, item)
			}
		}
		if maxCount < len(filterDataList) {
			filterDataList = filterDataList[:maxCount]
		}
	}
	if DEBUG {
		consoleLog.Printf("Read all after v%d, maxCount is %d.\n", afterStoreVersion, maxCount)
	}
	return filterDataList
}

func (cache *lockingInMemoryCache) Clear(executeWhenCommitting func()) {
	defer cache.locker.Unlock()
	cache.locker.Lock()

	executeWhenCommitting()
	cache.cacheFull = []DataWithKey{}
	cache.cacheByKey = make(map[string][]DataWithKey, 0)
	cache.StoreVersion = 0
}
