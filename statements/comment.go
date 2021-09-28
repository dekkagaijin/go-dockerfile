package statements

import (
	"strings"
)

const commentToken = "#"

type Comment struct {
	// Lines are the lines of the comment (including leading whitespace), minus the "#" token.
	Lines []string
}

func (s *Comment) Type() Type {
	return CommentType
}

func (s *Comment) String() string {
	return commentToken + strings.Join(s.Lines, "\n"+commentToken)
}
