package parser

import (
	"fmt"
	"regexp"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

const reArgKeyValuePair = "(?:" +
	"(" + reNotWhitespaceOrEquals + ")" + // key
	"=" +
	("(" +
		(`(?:'` + `(?:\'|[^'])*` + `')`) + // single-quoted val
		"|" + // or
		(`(?:"` + `(?:\"|[^"])*` + `")`) + // double quoted val
		"|" + // or
		(`(?:` + `(?:\\.|[[:^space:]])+` + `)`) + // unquoted (possibly escaped) val
		")") +
	")"

var argInstructionArgsMatcher = regexp.MustCompile(
	reStartOfLine +
		"(?:" +
		reArgKeyValuePair +
		"|" + // or
		"(" + reNotWhitespaceOrEquals + ")" + // just an arg name
		")" +
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
	if len(reMatches) == 0 {
		return nil, lines, fmt.Errorf("syntax error, ARG args must be of the form `<name>[=<default value>]`: %q", rawArgs)
	}

	inst := &statement.ArgInstruction{
		Name:       reMatches[1],
		DefaultVal: reMatches[2],
	}
	if reMatches[3] != "" {
		inst.Name = reMatches[3]
	}
	return inst, remainingLines, nil
}
