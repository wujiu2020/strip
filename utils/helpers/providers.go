package helpers

import (
	"github.com/teapots/params"
	"github.com/teapots/teapot"
)

func LoadClassicProviders(tea *teapot.Teapot) {
	// params parser
	tea.Provide(params.ParamsParser())
}
