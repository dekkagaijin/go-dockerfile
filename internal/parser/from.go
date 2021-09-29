package parser

import (
	"fmt"
	"regexp"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

var fromInstructionMatcher = regexp.MustCompile(
	"(?i)" + // case insensitive
		reStartOfLine +
		reOptionalWhitespace +
		"FROM" +
		reWhitespace +
		"(?:--platform=(" + reNotWhitespace + ")" + reWhitespace + ")?" + // platform, group 1
		"(" + reNotWhitespace + ")" + // image, group 2
		"(?:AS" + reWhitespace + "(" + reNotWhitespace + "))?" + // alias, group 3
		reDontCare +
		reEndOfLine)

func scanFROM(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	st, rawArgs, statementLines, remainingLines, err := scanInstructionLines(lines, escapeCharacter)
	if err != nil {
		return nil, lines, err
	}
	if st != statement.FROM {
		return nil, lines, fmt.Errorf("not a FROM statement: %q", statementLines[0])
	}

	stmtLine := string(statement.FROM) + " " + rawArgs
	reMatches := fromInstructionMatcher.FindStringSubmatch(stmtLine)
	if len(reMatches) != 4 {
		return nil, lines, fmt.Errorf("syntax error, could not parse FROM statement: %q", stmtLine)
	}
	inst := &statement.FromInstruction{
		Platform: reMatches[1],
		Image:    reMatches[2],
		Alias:    reMatches[3],
	}
	return inst, remainingLines, nil
}
