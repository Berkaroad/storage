package storage

import (
	"log"
	"os"
)

var consoleLog = log.New(os.Stdout, "[storage] ", log.LstdFlags)
