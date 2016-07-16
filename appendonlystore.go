package storage

import (
	"fmt"
)

// AppendOnlyStore is to append item in store.
type AppendOnlyStore interface {
	// Constructor
	InitFunc() interface{}
	// Append data with expected stream version by stream name
	Append(streamName string, data []byte, expectedStreamVersion int) error
	// Read by stream name
	ReadRecords(streamName string, startingFrom, maxCount int) []DataWithKey
	// Read all
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
	// Stream version
	StreamVersion int
	// Store version
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
