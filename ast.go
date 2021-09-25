package dockerfile

import (
	"strings"
)

type StatementType string

const (
	AddStatement         StatementType = "ADD"
	ArgStatement         StatementType = "ARG"
	CmdStatement         StatementType = "CMD"
	CommentStatement     StatementType = "#"
	CopyStatement        StatementType = "COPY"
	EntrypointStatement  StatementType = "ENTRYPOINT"
	EnvStatement         StatementType = "ENV"
	ExposeStatement      StatementType = "EXPOSE"
	FromStatement        StatementType = "FROM"
	HealthcheckStatement StatementType = "HEALTHCHECK"
	LabelStatement       StatementType = "LABEL"
	MaintainerStatement  StatementType = "MAINTAINER"
	OnbuildStatement     StatementType = "ONBUILD"
	RunStatement         StatementType = "RUN"
	ShellStatement       StatementType = "SHELL"
	StopSignalStatement  StatementType = "STOPSIGNAL"
	UserStatement        StatementType = "USER"
	VolumeStatement      StatementType = "VOLUME"
	WorkdirStatement     StatementType = "WORKDIR"
)

var KnownStatementTypes = map[StatementType]interface{}{
	AddStatement:         nil,
	ArgStatement:         nil,
	CmdStatement:         nil,
	CommentStatement:     nil,
	CopyStatement:        nil,
	EntrypointStatement:  nil,
	EnvStatement:         nil,
	ExposeStatement:      nil,
	FromStatement:        nil,
	HealthcheckStatement: nil,
	LabelStatement:       nil,
	MaintainerStatement:  nil,
	OnbuildStatement:     nil,
	RunStatement:         nil,
	ShellStatement:       nil,
	StopSignalStatement:  nil,
	UserStatement:        nil,
	VolumeStatement:      nil,
	WorkdirStatement:     nil,
}

// StatementMatchers is a mapping of statement types to matching regular expressions.
// var StatementMatchers = map[StatementType]*regexp.Regexp{
// 	AddStatement:         instructionMatcher(AddStatementType),
// 	ArgStatement:         instructionMatcher(ArgStatementType),
// 	CmdStatement:         instructionMatcher(CmdStatementType),
// 	CommentStatement:     regexp.MustCompile(reStartOfLine + reMaybeWhitespace + string(CommentStatementType) + ".*" + reEndOfLine),
// 	CopyStatement:        instructionMatcher(CopyStatementType),
// 	EntrypointStatement:  instructionMatcher(EntrypointStatementType),
// 	EnvStatement:         instructionMatcher(EnvStatementType),
// 	ExposeStatement:      instructionMatcher(ExposeStatementType),
// 	FromStatement:        instructionMatcher(FromStatementType),
// 	HealthcheckStatement: instructionMatcher(HealthcheckStatementType),
// 	LabelStatement:       instructionMatcher(LabelStatementType),
// 	MaintainerStatement:  instructionMatcher(MaintainerStatementType),
// 	OnbuildStatement:     instructionMatcher(OnbuildStatementType),
// 	RunStatement:         instructionMatcher(RunStatementType),
// 	ShellStatement:       instructionMatcher(ShellStatementType),
// 	StopSignalStatement:  instructionMatcher(StopSignalStatementType),
// 	UserStatement:        instructionMatcher(UserStatementType),
// 	VolumeStatement:      instructionMatcher(VolumeStatementType),
// 	WorkdirStatement:     instructionMatcher(WorkdirStatementType),
// }

type Statement interface {
	StatementType() StatementType
}

type Instruction interface {
	Statement
	// Flags are the flags passed to
	Flags() map[string]string
	Arguments() []string
}

type AST struct {
	Statements []Statement
}

type TODOInstruction struct {
	Type StatementType

	Args []string

	// Raw is all of the raw input lines of the statement, minus newline escape characters.
	// This includes the instruction name in the first line, as well as any comment lines.
	Raw []string
}

func (i *TODOInstruction) StatementType() StatementType {
	return i.Type
}

func (*TODOInstruction) Flags() map[string]string {
	return nil
}

func (i *TODOInstruction) Arguments() []string {
	return i.Args
}

type Comment struct {
	// Lines are the lines of the comment (including leading whitespace), minus the "#" token.
	Lines []string
}

func (s *Comment) StatementType() StatementType {
	return CommentStatement
}

func (s *Comment) String() string {
	return dockerfileCommentToken + strings.Join(s.Lines, "\n"+dockerfileCommentToken)
}
