package parser

import "github.com/dekkagaijin/go-dockerfile/statement"

func scanENV(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	return scanGenericInstruction(lines, escapeCharacter)
}
