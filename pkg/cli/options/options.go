package options

import (
	"context"
)

type Options struct {
	Top    Top
	Run    Run
	Config Config

	GenDocs GenDocs
}

type Top struct {
	Ctx context.Context
}

type Run struct {
	File   string
	Values map[string]string
	Debug  bool
}

type GenDocs struct {
	Template string
	Output   string
}

type Config struct {
	GlobalConfigPath string
}
