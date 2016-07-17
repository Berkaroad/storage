/*
Package storage is to provide several store interface.
*/
package storage

import (
	"log"
	"os"
)

var consoleLog = log.New(os.Stdout, "[storage] ", log.LstdFlags)

// DEBUG is a switcher for debug
var DEBUG = false
