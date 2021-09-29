package parser

import (
	"strings"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

func scanENTRYPOINT(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	inst := &statement.GenericInstruction{InstructionType: statement.ENTRYPOINT}
	st, rawArgs, statementLines, remainingLines, err := scanInstructionLines(lines, escapeCharacter)
	if err != nil {
		return nil, lines, err
	}
	inst.InstructionType = st
	inst.Lines = statementLines
	if args, err := parseJSONStringList(rawArgs); err == nil {
		inst.Args = statement.Arguments{
			List:     args,
			Execable: true,
		}
	} else {
		inst.Args = statement.Arguments{
			List:     strings.Fields(rawArgs),
			Execable: false,
		}
	}
	return inst, remainingLines, nil
}
