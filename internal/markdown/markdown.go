package markdown

import (
	"fmt"
	"strings"
)

type Markdown struct {
	builder strings.Builder
}

func (md *Markdown) write(content string) {
	md.builder.WriteString(content)
}

func (md *Markdown) writeln(content string) {
	md.write(content + "\n")
}

func (md *Markdown) Head1(title string) {
	md.writeln(fmt.Sprintf("# %s", title))
}

func (md *Markdown) Head2(title string) {
	md.writeln(fmt.Sprintf("## %s", title))
}

func (md *Markdown) Head3(title string) {
	md.writeln(fmt.Sprintf("### %s", title))
}

func (md *Markdown) Link(text, link string) {
	md.writeln(fmt.Sprintf("[%s](%s) ", text, link))
}
