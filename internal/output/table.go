package output

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type Table struct {
	Headers []string
	Rows    [][]string
}

func (t *Table) Render(w io.Writer) {
	if len(t.Headers) == 0 {
		return
	}

	widths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		widths[i] = len(h)
	}
	for _, row := range t.Rows {
		for i := 0; i < len(widths) && i < len(row); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	for i, h := range t.Headers {
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%-*s", widths[i], strings.ToUpper(h))
	}
	fmt.Fprintln(w)

	for _, row := range t.Rows {
		for i := 0; i < len(t.Headers); i++ {
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			val := ""
			if i < len(row) {
				val = row[i]
			}
			fmt.Fprintf(w, "%-*s", widths[i], val)
		}
		fmt.Fprintln(w)
	}
}

func RelativeTime(t *time.Time) string {
	if t == nil {
		return "-"
	}
	d := time.Since(*t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
