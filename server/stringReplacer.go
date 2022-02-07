package server

import (
	"bytes"
	"context"
	"io"
	"strings"
)

type replacerCollection struct {
	replacer []replacer
}

type replacer interface {
	Replace(ctx context.Context, w io.Writer, input string) error
}

type staticCopy struct {
	data []byte
}

type inputCopy struct{}

func (replacer *staticCopy) Replace(ctx context.Context, w io.Writer, input string) error {
	r := bytes.NewReader(replacer.data)
	_, err := io.Copy(w, r)
	return err
}

func (replacer *inputCopy) Replace(ctx context.Context, w io.Writer, input string) error {
	data := []byte(input)
	r := bytes.NewReader(data)
	_, err := io.Copy(w, r)
	return err
}

func (replacer *replacerCollection) Replace(ctx context.Context, w io.Writer, input string) error {
	for _, subreplacer := range replacer.replacer {
		err := subreplacer.Replace(ctx, w, input)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReplacerCollectionFromInput(data []byte, toReplace string) (*replacerCollection, error) {
	fragments := strings.Split(string(data), toReplace)
	// build replacement chain
	replacer := make([]replacer, 0)
	for i := 0; i < len(fragments)-1; i++ {
		data = []byte(fragments[i])
		replacer = append(replacer, &staticCopy{data: data})
		replacer = append(replacer, &inputCopy{})
	}
	data = []byte(fragments[len(fragments)-1])
	replacer = append(replacer, &staticCopy{data: data})
	return &replacerCollection{replacer: replacer}, nil
}
