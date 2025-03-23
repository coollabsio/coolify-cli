package clitable

import (
	"fmt"
	"io"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	isatty "github.com/mattn/go-isatty"
)

// isColorEnabled returns true if color output should be used based on terminal support and user flag.
func isColorEnabled(disableColor bool) bool {
	if disableColor {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// Table wraps a go-pretty table.Writer with alignment and output options.
type Table struct {
	Writer      table.Writer
	output      io.Writer
	align       []text.Align
	alignHeader []text.Align
}

// NewStyledTable returns a table with rounded box styles and optional color.
func NewStyledTable(out io.Writer, disableColor bool) *Table {
	writer := table.NewWriter()
	useColor := isColorEnabled(disableColor)

	colorOpts := table.ColorOptions{}
	if useColor {
		colorOpts.Header = text.Colors{text.Italic}
		colorOpts.Border = text.Colors{text.FgHiBlack}
		colorOpts.Separator = text.Colors{text.FgHiBlack}
	}

	boxStyle := table.StyleBoxDefault
	if useColor {
		boxStyle = table.StyleBoxRounded
	}

	writer.SetStyle(table.Style{
		Box:     boxStyle,
		Color:   colorOpts,
		Format:  table.FormatOptions{},
		HTML:    table.DefaultHTMLOptions,
		Options: table.OptionsDefault,
		Title:   table.TitleOptionsDefault,
	})

	return &Table{
		Writer:      writer,
		output:      out,
		align:       make([]text.Align, 0),
		alignHeader: make([]text.Align, 0),
	}
}

// Render writes the formatted table to output.
func (t *Table) Render() {
	t.applyColumnConfigs()
	fmt.Fprintln(t.output, t.Writer.Render())
}

// SetHeaders sets the table header and default column alignments.
func (t *Table) SetHeaders(headers ...string) {
	t.align = make([]text.Align, len(headers))
	t.alignHeader = make([]text.Align, len(headers))

	row := make(table.Row, len(headers))
	for i, h := range headers {
		row[i] = h
		t.align[i] = text.AlignLeft
		t.alignHeader[i] = text.AlignCenter
	}
	t.Writer.AppendHeader(row)
}

// AddRow appends a new row to the table.
func (t *Table) AddRow(values ...string) {
	row := make(table.Row, len(values))
	for i, v := range values {
		row[i] = v
	}
	t.Writer.AppendRow(row)
}

// SetRowLines enables or disables row separators.
func (t *Table) SetRowLines(enabled bool) {
	style := t.Writer.Style()
	style.Options.SeparateRows = enabled
	t.Writer.SetStyle(*style)
}

// SetAlignment sets column alignment.
func (t *Table) SetAlignment(alignments ...text.Align) {
	copy(t.align, alignments)
}

// SetHeaderAlignment sets header alignment.
func (t *Table) SetHeaderAlignment(alignments ...text.Align) {
	copy(t.alignHeader, alignments)
}

// applyColumnConfigs applies alignment and column width settings.
func (t *Table) applyColumnConfigs() {
	configs := make([]table.ColumnConfig, len(t.align))
	for i := range configs {
		configs[i] = table.ColumnConfig{
			AlignHeader:      t.alignHeader[i],
			Align:            t.align[i],
			WidthMax:         60,
			WidthMaxEnforcer: text.WrapSoft,
		}
	}
	t.Writer.SetColumnConfigs(configs)
}
