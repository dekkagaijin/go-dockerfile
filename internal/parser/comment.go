package parser

import (
	"strings"

	"github.com/dekkagaijin/go-dockerfile/statements"
)

func scanComment(lines []string) (stmt statements.Statement, remainingLines []string, err error) {
	remainingLines = lines
	cmnt := &statements.Comment{}
	for len(remainingLines) > 0 && commentLineMatcher.MatchString(remainingLines[0]) {
		curr := strings.TrimSpace(remainingLines[0])
		curr = strings.TrimPrefix(curr, CommentToken) // remove leading "#"
		cmnt.Lines = append(cmnt.Lines, curr)
		remainingLines = remainingLines[1:]
	}
	return cmnt, remainingLines, nil
}
