package cmd

import (
	"fmt"
	"io"
	"os"
)

type Reporter interface {
	Print(format string, args... interface{})
	Println(format string, args... interface{})
}

func Stdout() Reporter {
	return &Printer{
		w: os.Stdout,
	}
}

func Stderr() Reporter {
	return &Printer{
		w: os.Stderr,
	}
}

type Printer struct {
	w io.Writer
}

func (p *Printer) Print(format string, args... interface{}) {
	fmt.Printf(format, args...)
}

func (p *Printer) Println(format string, args... interface{}) {
	fmt.Printf(format + "\n", args...)
}
