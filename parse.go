package dockerfile

import (
	"fmt"
	"io"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

const (
	dockerfileCommentToken = "#"
)

type Parsed struct {
	ast         *parser.Node
	escapeToken rune
}

func Parse(dockerfile io.Reader) (*Parsed, error) {
	result, err := parser.Parse(dockerfile)
	if err != nil {
		return nil, err
	}

	return &Parsed{ast: result.AST, escapeToken: result.EscapeToken}, nil
}

func renderNode(w io.Writer, node *parser.Node) error {
	curr := node
	for curr != nil {
		//TODO: make sure to pick up https://github.com/moby/buildkit/pull/2375
		for _, line := range curr.PrevComment {
			if _, err := fmt.Fprintln(w, dockerfileCommentToken, line); err != nil {
				return err
			}
		}
		if curr.Original != "" {
			if _, err := fmt.Fprintln(w, curr.Original); err != nil {
				return err
			}
		}
		for i, child := range curr.Children {
			if child.Value == "FROM" && i != 0 {
				// Write newlines between `FROM` blocks for great readability.
				fmt.Fprintln(w)
			}
			renderNode(w, child)
		}
		curr = curr.Next
	}
	return nil
}

// Render writes a human-readable canonical Dockerfile to the provided Writer
func (df *Parsed) Render(w io.Writer) error {
	return renderNode(w, df.ast)
}

// String implements Stringer
func (df *Parsed) String() string {
	var sb strings.Builder
	df.Render(&sb)
	return sb.String()
}
