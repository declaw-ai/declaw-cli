package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Printer struct {
	JSON   bool
	Writer io.Writer
}

func New(jsonMode bool) *Printer {
	return &Printer{JSON: jsonMode, Writer: os.Stdout}
}

func (p *Printer) PrintJSON(data interface{}) error {
	enc := json.NewEncoder(p.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (p *Printer) PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}
	t := &Table{Headers: headers, Rows: rows}
	t.Render(p.Writer)
}

func (p *Printer) PrintMessage(format string, args ...interface{}) {
	fmt.Fprintf(p.Writer, format+"\n", args...)
}
