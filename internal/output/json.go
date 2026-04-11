package output

import (
	"encoding/json"
	"io"
)

type jsonPrinter struct {
	w io.Writer
}

func (p *jsonPrinter) Print(v any) error {
	return p.write(v)
}

func (p *jsonPrinter) PrintList(v any) error {
	return p.write(v)
}

func (p *jsonPrinter) write(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	_, err = p.w.Write(data)

	return err
}
