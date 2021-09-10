package alog

import (
	"log"
	"os"
)

var RECOVER = log.New(os.Stdout, "[RECOVER]", log.LstdFlags|log.Lshortfile)
var ERROR = log.New(os.Stdout, "[ERROR]", log.LstdFlags|log.Lshortfile)
var INFO = log.New(os.Stdout, "[INFO]", log.LstdFlags|log.Lshortfile)
