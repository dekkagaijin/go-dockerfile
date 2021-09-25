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
	for _, stmt := range df.Statements {
		if cmnt, ok := stmt.(*Comment); ok {
			for _, line := range cmnt.Lines {
				fmt.Fprint(out, dockerfileCommentToken)
				fmt.Fprintln(out, line)
			}
		} else if inst, ok := stmt.(Instruction); ok {
			fmt.Fprint(out, string(inst.StatementType()))
			for k, v := range inst.Flags() {
				fmt.Fprint(out, " --", k, "=", v)
			}
			for _, arg := range inst.Arguments() {
				fmt.Fprint(out, " ", arg)
			}
			fmt.Fprintln(out)
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
