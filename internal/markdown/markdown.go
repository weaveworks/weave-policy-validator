package markdown

import (
	"fmt"
	"strings"
)

type Markdown struct {
	builder strings.Builder
}

func New() *Markdown {
	return &Markdown{}
}

// String returns markdown text as string
func (md *Markdown) String() string {
	return md.builder.String()
}

func (md *Markdown) write(content string) {
	md.builder.WriteString(content)
}

func (md *Markdown) writeln(content string) {
	md.write(content + "\n")
}

// Head1 adds h1
func (md *Markdown) Head1(text string, args ...interface{}) {
	md.writeln(fmt.Sprintf("# %s", fmt.Sprintf(text, args...)))
}

// Head2 adds h2
func (md *Markdown) Head2(text string, args ...interface{}) {
	md.writeln(fmt.Sprintf("## %s", fmt.Sprintf(text, args...)))
}

// Head3 adds h3
func (md *Markdown) Head3(text string, args ...interface{}) {
	md.writeln(fmt.Sprintf("## %s", fmt.Sprintf(text, args...)))
}

// Paragraph adds paragraph
func (md *Markdown) Paragraph(text string, args ...interface{}) {
	md.writeln(fmt.Sprintf(text, args...))
}

// Link adds link
func (md *Markdown) Link(text, link string) {
	md.writeln(fmt.Sprintf("[%s](%s) ", text, link))
}

// Table adds table
func (md *Markdown) Table(columns []string, rows [][]string) {
	sperator := []string{}
	for range columns {
		sperator = append(sperator, "---")
	}
	md.writeln(fmt.Sprintf("| %s |", strings.Join(columns, "|")))
	md.writeln(fmt.Sprintf("| %s |", strings.Join(sperator, "|")))
	for _, row := range rows {
		md.writeln(fmt.Sprintf("| %s |", strings.Join(row, "|")))
	}
	md.write("\n")
}
