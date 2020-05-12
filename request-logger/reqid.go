package reqlogger

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

var rnd = uint32(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(999999999))

func NewReqId() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], rnd)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

func DecodeReqId(reqId string) (uint, int64) {
	b, err := base64.URLEncoding.DecodeString(reqId)
	if err != nil || len(b) < 12 {
		return 0, 0
	}
	rnd := binary.LittleEndian.Uint32(b[:4])
	unixNano := binary.LittleEndian.Uint64(b[4:])
	return uint(rnd), int64(unixNano)
}

var goroutineSpace = []byte("goroutine ")

func curGoroutineID() string {
	debug.PrintStack()
	bp := littleBuf.Get().(*[]byte)
	defer littleBuf.Put(bp)
	b := *bp
	b = b[:runtime.Stack(b, false)]
	// Parse the 4707 out of "goroutine 4707 ["
	b = bytes.TrimPrefix(b, goroutineSpace)
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		panic(fmt.Sprintf("No space found in %q", b))
	}
	b = b[:i]
	return string(b)
}

var littleBuf = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 64)
		return &buf
	},
}
