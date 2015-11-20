package storage

import (
	"fmt"
)

type initializer interface {
	InitFunc() interface{}
}

// 只追加方式的存储
type AppendOnlyStore interface {
	// 构造函数
	initializer
	// 根据传入的Stream名称，追加数据
	Append(streamName string, data []byte, expectedStreamVersion int) error
	// 根据传入的Stream名称，读取数据
	ReadRecords(streamName string, startingFrom, maxCount int) []DataWithKey
	// 读取所有数据
	ReadAllRecords(startingFrom, maxCount int) []DataWithKey
	// 释放相关的资源，如网络连接等
	Close()
	// 重置存储
	Reset()
	// 获取当前的存储版本号（每个streamName独立的版本号计数，此为总的版本号计数）
	GetCurrentVersion() int
}

// 键值对存储单元
type DataWithKey struct {
	// 键，这里记录Stream名称
	Key string
	// 数据
	Data []byte
	// Stream版本
	StreamVersion int
	// 存储版本
	StoreVersion int
}

// 只追加存储并发错误
type AppendOnlyStoreConcurrencyError struct {
	// 实际Stream版本号
	ActualStreamVersion int
	// 预期Stream版本号
	ExpectedStreamVersion int
	// Stream名称
	StreamName string
}

func (self *AppendOnlyStoreConcurrencyError) Error() string {
	message := fmt.Sprintf("AppendOnlyStoreConcurrencyError: Expected version %d in stream '%s' but got %d!", self.ExpectedStreamVersion, self.StreamName, self.ActualStreamVersion)
	return message
}
