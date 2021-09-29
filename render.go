package dockerfile

import (
	"fmt"
	"io"
	"strings"

	"github.com/dekkagaijin/go-dockerfile/internal/parser"
	"github.com/dekkagaijin/go-dockerfile/statement"
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
	if df.EscapeCharacter != DefaultExcapeCharacter {
		fmt.Fprintln(out, CommentToken, parser.EscapeParserDirectiveKey+"="+string(df.EscapeCharacter))
		fmt.Fprintln(out)
	}
	for i, stmt := range df.Statements {
		if i > 0 {
			// Avoid adding a newline at the end of the file.
			fmt.Fprintln(out)
		}
		st := stmt.Type()
		if cmnt, ok := stmt.(*statement.Comment); ok {
			if i > 0 && df.Statements[i-1].Type() == statement.CommentType {
				// Add a blank line between distinct comment blocks
				fmt.Fprintln(out)
			}
			for j, line := range cmnt.Lines {
				if j > 0 {
					fmt.Fprintln(out)
				}
				fmt.Fprint(out, parser.CommentToken, line)
			}
		} else if inst, ok := stmt.(statement.Instruction); ok {
			if st == statement.FROM && i > 0 && df.Statements[i-1].Type() != statement.CommentType {
				// Add a blank line between FROM statement blocks
				fmt.Fprintln(out)
			}
			fmt.Fprint(out, string(inst.Type()))
			for k, v := range inst.Flags() {
				fmt.Fprint(out, " --", k, "=", v)
			}
			arguments := inst.Arguments()
			if arguments.Execable {
				fmt.Fprint(out, ` [ "`, strings.Join(arguments.List, `", "`), `" ]`)
			} else {
				for _, arg := range arguments.List {
					fmt.Fprint(out, " ", arg)
				}
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
