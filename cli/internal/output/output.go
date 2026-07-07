package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	Bold   = color.New(color.Bold)
	Green  = color.New(color.FgGreen, color.Bold)
	Red    = color.New(color.FgRed, color.Bold)
	Yellow = color.New(color.FgYellow, color.Bold)
	Cyan   = color.New(color.FgCyan, color.Bold)
	Blue   = color.New(color.FgBlue, color.Bold)
	Dim    = color.New(color.FgHiBlack)
)

func Success(format string, args ...interface{}) {
	Green.Printf("✓ ")
	fmt.Fprintf(color.Output, format+"\n", args...)
}

func Error(format string, args ...interface{}) {
	Red.Printf("✗ ")
	fmt.Fprintf(color.Output, format+"\n", args...)
}

func Warn(format string, args ...interface{}) {
	Yellow.Printf("! ")
	fmt.Fprintf(color.Output, format+"\n", args...)
}

func Info(format string, args ...interface{}) {
	Cyan.Printf("▶ ")
	fmt.Fprintf(color.Output, format+"\n", args...)
}

func Section(title string) {
	fmt.Println()
	Bold.Println(title)
	Dim.Println(strings.Repeat("─", len(title)))
}

func JSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func Fatal(err error) {
	Error(err.Error())
	os.Exit(1)
}

func Table(header []string, rows [][]string) {
	if len(rows) == 0 {
		Dim.Println("  (none)")
		return
	}
	colWidths := make([]int, len(header))
	for i, h := range header {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	total := 0
	for _, w := range colWidths {
		total += w + 3
	}
	Bold.Println(strings.Repeat("─", total))
	line := ""
	for i, h := range header {
		line += fmt.Sprintf(" %-*s  ", colWidths[i], h)
	}
	Bold.Println(line)
	Bold.Println(strings.Repeat("─", total))
	for _, row := range rows {
		line := ""
		for i, cell := range row {
			line += fmt.Sprintf(" %-*s  ", colWidths[i], cell)
		}
		fmt.Println(line)
	}
	Dim.Println(strings.Repeat("─", total))
}
