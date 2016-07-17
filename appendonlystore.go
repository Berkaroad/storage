package storage

import (
	"fmt"
)

// AppendOnlyStore is a store to append item in store.
type AppendOnlyStore interface {
	// Constructor
	InitFunc() interface{}
	// Append data by stream name, will check expected stream version
	Append(streamName string, data []byte, expectedStreamVersion int) error
	// Read by stream name, starting from at least zero
	ReadRecords(streamName string, startingFrom, maxCount int) []DataWithKey
	// Read all, starting from at least zero
	ReadAllRecords(startingFrom, maxCount int) []DataWithKey
	// Close and dispose
	Close()
	// Reset
	Reset()
	// Get current global version
	GetCurrentVersion() int
}

// DataWithKey with key-value
type DataWithKey struct {
	// Key, stream name
	Key string
	// Data
	Data []byte
	// Stream version, start value is zero
	StreamVersion int
	// Store version, start value is zero
	StoreVersion int
}

// AppendOnlyStoreConcurrencyError is concurrency error.
type AppendOnlyStoreConcurrencyError struct {
	// Actual stream version
	ActualStreamVersion int
	// Expected stream version
	ExpectedStreamVersion int
	// Stream name
	StreamName string
}

func (concError *AppendOnlyStoreConcurrencyError) Error() string {
	message := fmt.Sprintf("AppendOnlyStoreConcurrencyError: Expected version %d in stream '%s' but got %d!", concError.ExpectedStreamVersion, concError.StreamName, concError.ActualStreamVersion)
	return message
}
