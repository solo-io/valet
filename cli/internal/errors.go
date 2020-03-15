package internal

import errors "github.com/rotisserie/eris"

var (
	RootAddError = errors.Errorf("please select a subcommand")
)
