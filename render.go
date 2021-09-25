package dockerfile

import (
	"fmt"
	"io"
)

type Renderer struct {
	escapeCharacter rune
	//skipComments    bool
}

var defaultRenderer = Renderer{escapeCharacter: DefaultExcapeCharacter}

// func NewRenderer(escapeCharacter rune, skipComments bool) *Renderer {
// 	return &Renderer{escapeCharacter: escapeCharacter, skipComments: skipComments}
// }

func (p Renderer) Render(df *AST, out io.Writer) error {
	for i, stmt := range df.Statements {
		if i > 0 {
			// Only add newlines between statements, not at the end of the file.
			fmt.Fprintln(out)
		}
		st := stmt.StatementType()
		if cmnt, ok := stmt.(*Comment); ok {
			for j, line := range cmnt.Lines {
				if j > 0 {
					fmt.Fprintln(out)
				}
				fmt.Fprint(out, dockerfileCommentToken, line)
			}
		} else if inst, ok := stmt.(Instruction); ok {
			if st == FromStatement && i > 0 && df.Statements[i-1].StatementType() == CommentStatement {
				// Add a blank line between FROM statement blocks
				fmt.Fprintln(out)
			}
			fmt.Fprint(out, string(inst.StatementType()))
			for k, v := range inst.Flags() {
				fmt.Fprint(out, " --", k, "=", v)
			}
			for _, arg := range inst.Arguments() {
				fmt.Fprint(out, " ", arg)
			}
		} else {
			return fmt.Errorf("unknown statement type: %s", stmt.StatementType())
		}
	}
	return nil
}

// Render parses the
func Render(df *AST, out io.Writer) error {
	return defaultRenderer.Render(df, out)
}
