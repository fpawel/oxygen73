package gui

import "io"

func NewWriter() io.Writer {
	return writer{}
}

type writer struct{}

func (x writer) Write(p []byte) (int, error) {
	go W{}.WriteConsole(string(p))
	return len(p), nil
}
