package helpers

import (
	"context"
	"log"
	"os"

	"github.com/wujiu2020/strip"
	"github.com/wujiu2020/strip/utils"
)

var X strip.LoggerAdv

func init() {
	logOut := log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	// 每机器每进程的 RequestId
	hostname, _ := os.Hostname()

	logger := strip.NewReqLogger(logOut, hostname)
	logger.SetLineInfo(true)
	logger.SetShortLine(true)
	logger.SetColorMode(false)
	logger.EnableLogStack(strip.LevelCritical)

	X = logger
}

func GetLogger(ctx context.Context) (log strip.Logger) {
	if err := utils.CtxFindValue(ctx, &log); err != nil {
		log = X
	}
	return
}

func WithLogger(ctx context.Context, log strip.Logger) context.Context {
	ctx = utils.CtxWithValue(ctx, &log)
	return ctx
}
