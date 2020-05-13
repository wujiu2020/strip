package helpers

import (
	"fmt"
	"log"
	"os"

	"github.com/wujiu2020/strip"
	reqlogger "github.com/wujiu2020/strip/request-logger"
)

const (
	HookFlagBeforeAll HookFlag = iota + 1
	HookFlagAfterAll
)

type HookFlag int

type HookFunc func(*strip.Strip, HookFlag)

func LoadClassicFilters(sp *strip.Strip, options ...interface{}) {
	logOut := log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	var loggerOption *reqlogger.LoggerOption
	var hooks []HookFunc
	for _, option := range options {
		if v, ok := option.(reqlogger.LoggerOption); ok {
			loggerOption = &v
			continue
		}
		if h, ok := option.(HookFunc); ok {
			hooks = append(hooks, h)
			continue
		}

		panic(fmt.Sprintf("not support %t as filter options", option))
	}
	if loggerOption == nil {
		loggerOption = &reqlogger.LoggerOption{
			ColorMode:     sp.Config.RunMode.IsDev(),
			LineInfo:      true,
			ShortLine:     sp.Config.RunMode.IsProd(),
			FlatLine:      sp.Config.RunMode.IsProd(),
			LogStackLevel: strip.LevelCritical,
		}
	}

	sp.Filter(
		// 所有过滤器之前抓取 panic
		strip.RecoveryFilter(),
	)

	for _, h := range hooks {
		h(sp, HookFlagBeforeAll)
	}

	sp.Filter(
		// 在静态文件之后加入，跳过静态文件请求
		reqlogger.ReqLoggerFilter(logOut, *loggerOption),

		// 在 action 里直接返回一般请求结果
		strip.GenericOutFilter(),
	)
}
