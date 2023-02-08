package console_print

import (
	tm "github.com/buger/goterm"
)

type ConsolePrint struct {
	top, bottom, center, log []string
	maxCenter, maxLog        int
}

func NewConsolePrint(top, bottom []string) ConsolePrint {
	return ConsolePrint{
		top:       top,
		bottom:    bottom,
		maxCenter: 3,
		maxLog:    3,
	}
}

func (p ConsolePrint) Update(outStrings ...string) {
	p.update()
}

func (p ConsolePrint) update() {
	tm.Clear() // Clear current screen
	tm.MoveCursor(1, 1)
	p.printTop()
	for _, s := range p.center {
		tm.Println(s)
	}
	for _, s := range p.log {
		tm.Println(s)
	}
	p.printBottom()
	tm.Flush()
}

func (p ConsolePrint) printTop() {
	for _, s := range p.top {
		tm.Println(tm.Color(tm.Bold(s), tm.GREEN))
	}
}

func (p ConsolePrint) printBottom() {
	for _, s := range p.bottom {
		tm.Println(tm.Color(tm.Bold(s), tm.YELLOW))
	}
}

func (p ConsolePrint) Log(string string) {
	p.log = append(p.log, string)
	p.update()
}

type Logger interface {
	Log(string string)
}
