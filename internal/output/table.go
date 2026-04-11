package output

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type tablePrinter struct {
	w io.Writer
}

var (
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("24")).
				Padding(0, 1)

	tableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	tableAltCellStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(lipgloss.Color("236"))

	tableKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Padding(0, 1)

	tableValueStyle = lipgloss.NewStyle().
			Padding(0, 1)

	tableAltValueStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(lipgloss.Color("236"))
)

func (p *tablePrinter) Print(v any) error {
	fields, err := recordFields(v)
	if err != nil {
		return err
	}

	if len(fields) == 0 {
		_, err = io.WriteString(p.w, "\n")
		return err
	}

	labels := make([]string, 0, len(fields))
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		labels = append(labels, field.Label)
		values = append(values, field.Value)
	}

	keyWidth := maxWidth(labels)
	valueWidth := maxWidth(values)
	if valueWidth == 0 {
		valueWidth = 1
	}

	var builder strings.Builder
	for idx, field := range fields {
		valueStyle := tableValueStyle
		if idx%2 == 1 {
			valueStyle = tableAltValueStyle
		}

		builder.WriteString(tableKeyStyle.Width(keyWidth + 2).Render(field.Label))
		builder.WriteString(valueStyle.Width(valueWidth + 2).Render(field.Value))
		builder.WriteByte('\n')
	}

	_, err = io.WriteString(p.w, builder.String())
	return err
}

func (p *tablePrinter) PrintList(v any) error {
	records, err := listRecords(v)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		_, err = io.WriteString(p.w, "\n")
		return err
	}

	headers := make([]string, 0, len(records[0]))
	widths := make([]int, 0, len(records[0]))
	for _, field := range records[0] {
		headers = append(headers, field.Label)
		width := lipgloss.Width(field.Label)
		if width < 1 {
			width = 1
		}
		widths = append(widths, width)
	}

	for _, record := range records {
		for idx, field := range record {
			if width := lipgloss.Width(field.Value); width > widths[idx] {
				widths[idx] = width
			}
		}
	}

	var builder strings.Builder
	for idx, header := range headers {
		builder.WriteString(tableHeaderStyle.Width(widths[idx] + 2).Render(header))
	}
	builder.WriteByte('\n')

	for rowIdx, record := range records {
		rowStyle := tableCellStyle
		if rowIdx%2 == 1 {
			rowStyle = tableAltCellStyle
		}

		for colIdx, field := range record {
			builder.WriteString(rowStyle.Width(widths[colIdx] + 2).Render(field.Value))
		}
		builder.WriteByte('\n')
	}

	_, err = io.WriteString(p.w, builder.String())
	return err
}

func maxWidth(values []string) int {
	width := 0
	for _, value := range values {
		if current := lipgloss.Width(value); current > width {
			width = current
		}
	}

	return width
}
