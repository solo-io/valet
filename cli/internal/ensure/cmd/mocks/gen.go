package mocks

//go:generate mockgen -destination command_runner_mock.go -self_package github.com/solo-io/valet/cli/internal/ensure/cmd/mocks -package mocks github.com/solo-io/valet/cli/internal/ensure/cmd CommandRunner
