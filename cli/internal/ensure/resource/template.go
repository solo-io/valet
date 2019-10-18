package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

type Template struct {
	FormatFile *FormatFile `yaml:"formatFile"`
}

func (t *Template) Ensure(ctx context.Context, command cmd.Factory) error {
	if t.FormatFile != nil {
		return t.FormatFile.Ensure(ctx, command)
	}
	return nil
}

func (t *Template) Teardown(ctx context.Context, command cmd.Factory) error {
	if t.FormatFile != nil {
		return t.FormatFile.Teardown(ctx, command)
	}
	return nil
}

type FormatFile struct {
	Path string   `yaml:"path"`
	Args []string `yaml:"args"`
}

func (f *FormatFile) Ensure(ctx context.Context, command cmd.Factory) error {
	template, err := LoadFormatFile(ctx, f.Path)
	if err != nil {
		return err
	}
	var args []interface{}
	for _, arg := range f.Args {
		args = append(args, arg)
	}
	rendered := fmt.Sprintf(template, args...)
	return command.Kubectl().ApplyStdIn(rendered).Cmd().Run(ctx)
}

func (f *FormatFile) Teardown(ctx context.Context, command cmd.Factory) error {
	rendered, err := f.render(ctx)
	if err != nil {
		return err
	}
	return command.Kubectl().DeleteStdIn(rendered).Cmd().Run(ctx)
}

func (f *FormatFile) render(ctx context.Context) (string, error) {
	template, err := LoadFormatFile(ctx, f.Path)
	if err != nil {
		return "", err
	}
	var args []interface{}
	for _, arg := range f.Args {
		args = append(args, arg)
	}
	return fmt.Sprintf(template, args...), nil
}

func LoadFormatFile(ctx context.Context, path string) (string, error) {
	b, err := loadBytesFromPath(ctx, path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
