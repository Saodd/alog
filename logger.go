package alog

import (
	"log"
	"os"
)

var RECOVER = log.New(os.Stdout, "[RECOVER]", log.LstdFlags)
var ERROR = log.New(os.Stdout, "[ERROR]", log.LstdFlags)
var INFO = log.New(os.Stdout, "[INFO]", log.LstdFlags)
