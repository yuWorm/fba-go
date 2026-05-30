package fba

import (
	"github.com/yuWorm/fba-go/core/app"
	"github.com/yuWorm/fba-go/core/config"
)

type Application = app.Application
type Options = config.Options

func NewApplication(opts Options) (Application, error) {
	return app.New(opts)
}
