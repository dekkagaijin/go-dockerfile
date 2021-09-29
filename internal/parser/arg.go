package parser

import (
	"fmt"
	"regexp"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

var argInstructionArgsMatcher = regexp.MustCompile(
	reStartOfLine +
		reOptionalWhitespace +
		"(" + reNotWhitespaceOrEquals + ")" + // key
		reOptionalWhitespace +
		"(?:=" + reOptionalWhitespace + "(" + reNotWhitespaceOrEquals + "))?" + // defaultvalue
		reDontCare +
		reEndOfLine)

// ARG is an instruction of the form:
// `ARG <name>[=<default value>]`
// See: https://docs.docker.com/engine/reference/builder/#arg
func scanARG(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	st, rawArgs, statementLines, remainingLines, err := scanInstructionLines(lines, escapeCharacter)
	if err != nil {
		return nil, lines, err
	}
	if st != statement.ARG {
		return nil, lines, fmt.Errorf("not an ARG statement: %q", statementLines[0])
	}

	reMatches := argInstructionArgsMatcher.FindStringSubmatch(rawArgs)
	if len(reMatches) != 3 {
		return nil, lines, fmt.Errorf("syntax error, could not parse ARG args: %q", rawArgs)
	}
	inst := &statement.ArgInstruction{
		Name:       reMatches[1],
		DefaultVal: reMatches[2],
	}
	return inst, remainingLines, nil
}
