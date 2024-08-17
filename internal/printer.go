package internal

import (
	"fmt"
	"io"
	"math"
	"os"

	"golang.org/x/term"
)

type Printer struct {
	InReader  io.Reader
	OutWriter io.Writer
	ErrWriter io.Writer

	disableStyling bool
}

func NewPrinter(
	inReader io.Reader,
	outWriter io.Writer,
	errWriter io.Writer,
) *Printer {
	p := &Printer{
		InReader:  inReader,
		OutWriter: outWriter,
		ErrWriter: errWriter,
	}
	f, ok := outWriter.(*os.File)
	if ok {
		p.disableStyling = !term.IsTerminal(int(f.Fd()))
	} else {
		p.disableStyling = true
	}
	return p
}

func (p *Printer) Print(a ...any) {
	fmt.Fprint(p.OutWriter, a...)
}

func (p *Printer) Println(a ...any) {
	fmt.Fprintln(p.OutWriter, a...)
}

func (p *Printer) Printf(format string, a ...any) {
	fmt.Fprintf(p.OutWriter, format, a...)
}

func (p *Printer) ErrPrint(a ...any) {
	fmt.Fprint(p.ErrWriter, a...)
}

func (p *Printer) ErrPrintln(a ...any) {
	fmt.Fprintln(p.ErrWriter, a...)
}

func (p *Printer) ErrPrintf(format string, a ...any) {
	fmt.Fprintf(p.ErrWriter, format, a...)
}

func (p *Printer) ColorForeground(s string, color uint8) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", color, s)
}

func (p *Printer) ColorBackground(s string, color uint8) string {
	if p.disableStyling {
		return s
	}
	return fmt.Sprintf("\033[48;5;%dm%s\033[0m", color, s)
}

func (p *Printer) GetStyling() bool {
	return !p.disableStyling
}

func (p *Printer) SetStyling(enable bool) {
	p.disableStyling = !enable
}

func (p *Printer) GetSize() (width, height int) {
	f, ok := p.OutWriter.(*os.File)
	if !ok {
		return math.MaxInt, math.MaxInt
	}
	w, h, _ := term.GetSize(int(f.Fd()))
	if w <= 0 {
		w = math.MaxInt
	}
	if h <= 0 {
		h = math.MaxInt
	}
	return w, h
}
