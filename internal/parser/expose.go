package parser

import "github.com/dekkagaijin/go-dockerfile/statements"

func scanEXPOSE(lines []string, escapeCharacter rune) (stmt statements.Statement, remainingLines []string, err error) {
	return scanGenericInstruction(lines, escapeCharacter)
}
