package io

import "io"

type Out interface {
	Write(msg string)
}

type DefaultOut struct {
	writer io.Writer
}

func NewDefaultOut(writer io.Writer) *DefaultOut {
	return &DefaultOut{writer: writer}
}

func (o *DefaultOut) Write(msg string) {
	o.writer.Write([]byte(msg))
}
