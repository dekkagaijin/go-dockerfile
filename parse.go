package dockerfile

import (
	"bufio"
	"io"

	"github.com/dekkagaijin/go-dockerfile/internal/parser"
	"github.com/dekkagaijin/go-dockerfile/statement"
)

const (
	DefaultExcapeCharacter = parser.DefaultExcapeCharacter
	WindowsEscapeCharacter = parser.WindowsEscapeCharacter
	CommentToken           = parser.CommentToken
)

type Parsed struct {
	Statements      []statement.Statement
	EscapeCharacter rune
}

// Parse parses the given Dockerfile.
func Parse(file io.Reader) (*Parsed, error) {
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	sp := parser.Sequential{}
	statements, escapeChar, err := sp.Parse(lines)
	if err != nil {
		return nil, err
	}
	return &Parsed{
		Statements:      statements,
		EscapeCharacter: escapeChar,
	}, nil
}
