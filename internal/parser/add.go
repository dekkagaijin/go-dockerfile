package parser

import (
	"fmt"
	"strings"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

func scanADD(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	inst := &statement.AddInstruction{}
	st, rawArgs, statementLines, remainingLines, err := scanInstructionLines(lines, escapeCharacter)
	if err != nil {
		return nil, lines, err
	}
	if st != statement.ADD {
		return nil, lines, fmt.Errorf("not an ADD statement: %q", statementLines[0])
	}
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
