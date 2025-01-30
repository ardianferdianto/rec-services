package logger

import (
	"io"
	"os"
)

type Option struct {
	Level  string
	Writer io.Writer
}

func NewOption(options ...func(*Option)) *Option {
	opt := &Option{
		Level:  "WARN",
		Writer: os.Stdout,
	}
	for _, o := range options {
		o(opt)
	}
	return opt
}

func WithLevel(level string) func(*Option) {
	return func(o *Option) {
		o.Level = level
	}
}

func WithWriter(writer io.Writer) func(*Option) {
	return func(o *Option) {
		o.Writer = writer
	}
}
