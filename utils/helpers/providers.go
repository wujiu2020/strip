package helpers

import (
	"github.com/wujiu2020/strip/params"
	"github.com/wujiu2020/strip/teapot"
)

func LoadClassicProviders(tea *teapot.Teapot) {
	// params parser
	tea.Provide(params.ParamsParser())
}
