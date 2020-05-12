package helpers

import (
	"context"
	"log"
	"os"

	"github.com/teapots/teapot"
	"github.com/teapots/utils"
)

var X teapot.LoggerAdv

func init() {
	logOut := log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	// 每机器每进程的 RequestId
	hostname, _ := os.Hostname()

	logger := teapot.NewReqLogger(logOut, hostname)
	logger.SetLineInfo(true)
	logger.SetShortLine(true)
	logger.SetColorMode(false)
	logger.EnableLogStack(teapot.LevelCritical)

	X = logger
}

func GetLogger(ctx context.Context) (log teapot.Logger) {
	if err := utils.CtxFindValue(ctx, &log); err != nil {
		log = X
	}
	return
}

func WithLogger(ctx context.Context, log teapot.Logger) context.Context {
	ctx = utils.CtxWithValue(ctx, &log)
	return ctx
}
