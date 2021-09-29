package parser

import (
	"strings"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

func scanComment(lines []string) (stmt statement.Statement, remainingLines []string, err error) {
	remainingLines = lines
	cmnt := &statement.Comment{}
	for len(remainingLines) > 0 && commentLineMatcher.MatchString(remainingLines[0]) {
		curr := strings.TrimSpace(remainingLines[0])
		curr = strings.TrimPrefix(curr, CommentToken) // remove leading "#"
		cmnt.Lines = append(cmnt.Lines, curr)
		remainingLines = remainingLines[1:]
	}
	return cmnt, remainingLines, nil
}
