package parser

import (
	"strings"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

func scanADD(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	inst := &statement.TODO{}
	st, rawArgs, statementLines, remainingLines, err := scanInstructionLines(lines, escapeCharacter)
	if err != nil {
		return nil, lines, err
	}
	inst.InstructionType = st
	inst.Lines = statementLines
	inst.Args = strings.Fields(rawArgs) // TODO accept JSON list?
	return inst, remainingLines, nil
}
