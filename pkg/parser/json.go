package parser

import (
	"io"

	"github.com/bytedance/sonic"
)

type JSON struct{}

func NewJSON() *JSON {
	return &JSON{}
}

func (JSON) Parse(format any, data io.Reader) error {
	decoder := sonic.ConfigDefault.NewDecoder(data)
	return decoder.Decode(&format)
}
