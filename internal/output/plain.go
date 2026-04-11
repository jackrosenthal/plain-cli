package output

import (
	"fmt"
	"io"
	"strings"
)

type plainPrinter struct {
	w io.Writer
}

func (p *plainPrinter) Print(v any) error {
	fields, err := recordFields(v)
	if err != nil {
		return err
	}

	return writePlainRecord(p.w, fields)
}

func (p *plainPrinter) PrintList(v any) error {
	records, err := listRecords(v)
	if err != nil {
		return err
	}

	for idx, record := range records {
		if idx > 0 {
			if _, err := fmt.Fprintln(p.w); err != nil {
				return err
			}
		}

		if err := writePlainRecord(p.w, record); err != nil {
			return err
		}
	}

	return nil
}

func writePlainRecord(w io.Writer, fields []fieldValue) error {
	lines := make([]string, 0, len(fields))
	for _, field := range fields {
		valueLines := strings.Split(field.Value, "\n")
		indent := strings.Repeat(" ", len(field.Key)+2)
		for i, vl := range valueLines {
			if i == 0 {
				lines = append(lines, fmt.Sprintf("%s: %s", field.Key, vl))
			} else {
				lines = append(lines, indent+vl)
			}
		}
	}

	if len(lines) == 0 {
		_, err := io.WriteString(w, "\n")
		return err
	}

	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")

	return err
}
