package dockerfile

import (
	"fmt"
	"io"

	"github.com/dekkagaijin/go-dockerfile/internal/parser"
	"github.com/dekkagaijin/go-dockerfile/statements"
)

type Renderer struct {
	escapeCharacter rune
	//skipComments    bool
}

var defaultRenderer = Renderer{escapeCharacter: DefaultExcapeCharacter}

// func NewRenderer(escapeCharacter rune, skipComments bool) *Renderer {
// 	return &Renderer{escapeCharacter: escapeCharacter, skipComments: skipComments}
// }

func (p Renderer) Render(df *Parsed, out io.Writer) error {
	for i, stmt := range df.Statements {
		if i > 0 {
			// Avoid adding a newline at the end of the file.
			fmt.Fprintln(out)
		}
		st := stmt.Type()
		if cmnt, ok := stmt.(*statements.Comment); ok {
			if i > 0 && df.Statements[i-1].Type() == statements.CommentType {
				// Add a blank line between distinct comment blocks
				fmt.Fprintln(out)
			}
			for j, line := range cmnt.Lines {
				if j > 0 {
					fmt.Fprintln(out)
				}
				fmt.Fprint(out, parser.CommentToken, line)
			}
		} else if inst, ok := stmt.(statements.Instruction); ok {
			if st == statements.FromType && i > 0 && df.Statements[i-1].Type() != statements.CommentType {
				// Add a blank line between FROM statement blocks
				fmt.Fprintln(out)
			}
			fmt.Fprint(out, string(inst.Type()))
			for k, v := range inst.Flags() {
				fmt.Fprint(out, " --", k, "=", v)
			}
			for _, arg := range inst.Arguments() {
				fmt.Fprint(out, " ", arg)
			}
		} else {
			return fmt.Errorf("unknown statement type: %s", stmt.Type())
		}
	}
	return nil
}

// Render parses the
func Render(df *Parsed, out io.Writer) error {
	return defaultRenderer.Render(df, out)
}
