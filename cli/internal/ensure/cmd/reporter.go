package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/logrusorgru/aurora"
)

const (
	ColorKey = "color"
)

type Reporter interface {
	Print(format string, args ...interface{})
	Println(format string, args ...interface{})
}

func getColorValue(ctx context.Context) uint8 {
	val, ok := ctx.Value(ColorKey).(uint8)
	if ok {
		return val
	}
	return 0
}

func Stdout(ctx context.Context) Reporter {
	return &Printer{
		w: os.Stdout,
		colorIndex: getColorValue(ctx),
	}
}

func Stderr(ctx context.Context) Reporter {
	return &Printer{
		w: os.Stderr,
		colorIndex: getColorValue(ctx),
	}
}

type Printer struct {
	w          io.Writer
	colorIndex uint8
}

func (p *Printer) Print(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(p.w, format, args...)
}

func (p *Printer) Println(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	stamp := aurora.Index(p.colorIndex, fmt.Sprintf("[%s]", time.Now().Format(time.RFC3339)))
	if p.colorIndex == 0 {
		stamp = aurora.Reset(fmt.Sprintf("[%s]", time.Now().Format(time.RFC3339)))
	}
	_, _ = fmt.Fprintf(p.w, "%s %s\n", stamp, msg)
}

func PromptPressAnyKeyToContinue() error {
	Stdout(context.TODO()).Println("Press any key to continue...")
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	return err
}
