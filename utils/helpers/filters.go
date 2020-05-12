package helpers

import (
	"fmt"
	"log"
	"os"

	"github.com/teapots/request-logger"
	"github.com/teapots/teapot"
)

const (
	HookFlagBeforeAll HookFlag = iota + 1
	HookFlagAfterAll
)

type HookFlag int

type HookFunc func(*teapot.Teapot, HookFlag)

func LoadClassicFilters(tea *teapot.Teapot, options ...interface{}) {
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
			ColorMode:     tea.Config.RunMode.IsDev(),
			LineInfo:      true,
			ShortLine:     tea.Config.RunMode.IsProd(),
			FlatLine:      tea.Config.RunMode.IsProd(),
			LogStackLevel: teapot.LevelCritical,
		}
	}

	tea.Filter(
		// 所有过滤器之前抓取 panic
		teapot.RecoveryFilter(),
	)

	for _, h := range hooks {
		h(tea, HookFlagBeforeAll)
	}

	tea.Filter(
		// 在静态文件之后加入，跳过静态文件请求
		reqlogger.ReqLoggerFilter(logOut, *loggerOption),

		// 在 action 里直接返回一般请求结果
		teapot.GenericOutFilter(),
	)
}
