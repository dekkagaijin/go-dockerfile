package parser

import "github.com/dekkagaijin/go-dockerfile/statement"

func scanLABEL(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	return scanGenericInstruction(lines, escapeCharacter)
}
