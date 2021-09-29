package parser

import "github.com/dekkagaijin/go-dockerfile/statement"

func scanSTOPSIGNAL(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	return scanTODO(lines, escapeCharacter)
}
