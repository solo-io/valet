package internal

import "github.com/solo-io/go-utils/errors"

var (
	RootAddError = errors.Errorf("please select a subcommand")
)
