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
	colorKey = "color"
)

type clusterPrinter struct {
	id         string
	colorIndex uint8
}

func NewPrinterContext(ctx context.Context, index uint8, id string) context.Context {
	return context.WithValue(ctx, colorKey, clusterPrinter{
		id:         id,
		colorIndex: index,
	})
}

func UpdatePrinterContext(ctx context.Context, id string) context.Context {
	val, ok := ctx.Value(colorKey).(clusterPrinter)
	if ok {
		val.id = id
		return context.WithValue(ctx, colorKey, val)
	}
	return context.WithValue(ctx, colorKey, clusterPrinter{
		id:         id,
	})
}

func clusterPrinterFromContext(ctx context.Context) clusterPrinter {
	val, ok := ctx.Value(colorKey).(clusterPrinter)
	if ok {
		return val
	}
	return clusterPrinter{
		id:         "unknown",
		colorIndex: 0,
	}
}

type Reporter interface {
	Println(format string)
	Printf(format string, args ...interface{})
}

func Stdout(ctx context.Context) Reporter {
	return &Printer{
		w:       os.Stdout,
		cluster: clusterPrinterFromContext(ctx),
	}
}

func Stderr(ctx context.Context) Reporter {
	return &Printer{
		w:       os.Stderr,
		cluster: clusterPrinterFromContext(ctx),
	}
}

type Printer struct {
	w       io.Writer
	cluster clusterPrinter
}

func (p *Printer) Println(format string) {
	p.Printf(format)
}

func (p *Printer) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	stamp := aurora.Index(p.cluster.colorIndex, fmt.Sprintf("[%s, id: %s]",
		time.Now().Format(time.RFC3339), p.cluster.id))
	if p.cluster.colorIndex == 0 {
		stamp = aurora.Reset(fmt.Sprintf("[%s, id: %s]", time.Now().Format(time.RFC3339), p.cluster.id))
	}
	_, _ = fmt.Fprintf(p.w, "%s %s\n", stamp, msg)
}

func PromptPressAnyKeyToContinue() error {
	Stdout(context.TODO()).Printf("Press any key to continue...")
	reader := bufio.NewReader(os.Stdin)
	_, _, err := reader.ReadRune()
	return err
}
