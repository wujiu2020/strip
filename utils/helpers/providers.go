package helpers

import (
	"github.com/wujiu2020/strip"
	"github.com/wujiu2020/strip/params"
)

func LoadClassicProviders(sp *strip.Strip) {
	// params parser
	sp.Provide(params.ParamsParser())
}
