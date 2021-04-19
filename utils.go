package alog

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"sync/atomic"
	"time"
)

var _seedCounter = new(int64)

func init() {
	src := rand.NewSource(time.Now().UnixNano())
	*_seedCounter = src.Int63()
}

// RandomBytes Generates random alphanumeric bytes.
func RandomBytes(length int) []byte {
	const randomCodeCharSet = `1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`
	// rand.Source is not concurrency-safe. So create one every time (on stack?)
	src := rand.NewSource(atomic.AddInt64(_seedCounter, 1))
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = randomCodeCharSet[src.Int63()%62]
	}
	return code
}

func main() {
	InitAlog("v1.0.0", "https://...", "812793r713452d") // 记得初始化！
	g := gin.New()
	g.Use(GinWithLogger(), GinWithTracker(), GinWithRecover()) // 注意顺序！
}
