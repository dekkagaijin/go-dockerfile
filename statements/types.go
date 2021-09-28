package statements

type Type string

const (
	AddType         Type = "ADD"
	ArgType         Type = "ARG"
	CmdType         Type = "CMD"
	CommentType     Type = "#"
	CopyType        Type = "COPY"
	EntrypointType  Type = "ENTRYPOINT"
	EnvType         Type = "ENV"
	ExposeType      Type = "EXPOSE"
	FromType        Type = "FROM"
	HealthcheckType Type = "HEALTHCHECK"
	LabelType       Type = "LABEL"
	MaintainerType  Type = "MAINTAINER"
	OnbuildType     Type = "ONBUILD"
	RunType         Type = "RUN"
	ShellType       Type = "SHELL"
	StopSignalType  Type = "STOPSIGNAL"
	UserType        Type = "USER"
	VolumeType      Type = "VOLUME"
	WorkdirType     Type = "WORKDIR"
)

var KnownTypes = map[Type]bool{
	AddType:         true,
	ArgType:         true,
	CmdType:         true,
	CommentType:     true,
	CopyType:        true,
	EntrypointType:  true,
	EnvType:         true,
	ExposeType:      true,
	FromType:        true,
	HealthcheckType: true,
	LabelType:       true,
	MaintainerType:  true,
	OnbuildType:     true,
	RunType:         true,
	ShellType:       true,
	StopSignalType:  true,
	UserType:        true,
	VolumeType:      true,
	WorkdirType:     true,
}

// TypeMatchers is a mapping of statement types to matching regular expressions.
// var TypeMatchers = map[Type]*regexp.Regexp{
// 	AddType:         instructionMatcher(AddType),
// 	ArgType:         instructionMatcher(ArgType),
// 	CmdType:         instructionMatcher(CmdType),
// 	CommentType:     regexp.MustCompile(reStartOfLine + reMaybeWhitespace + string(CommentType) + ".*" + reEndOfLine),
// 	CopyType:        instructionMatcher(CopyType),
// 	EntrypointType:  instructionMatcher(EntrypointType),
// 	EnvType:         instructionMatcher(EnvType),
// 	ExposeType:      instructionMatcher(ExposeType),
// 	FromType:        instructionMatcher(FromType),
// 	HealthcheckType: instructionMatcher(HealthcheckType),
// 	LabelType:       instructionMatcher(LabelType),
// 	MaintainerType:  instructionMatcher(MaintainerType),
// 	OnbuildType:     instructionMatcher(OnbuildType),
// 	RunType:         instructionMatcher(RunType),
// 	ShellType:       instructionMatcher(ShellType),
// 	StopSignalType:  instructionMatcher(StopSignalType),
// 	UserType:        instructionMatcher(UserType),
// 	VolumeType:      instructionMatcher(VolumeType),
// 	WorkdirType:     instructionMatcher(WorkdirType),
// }

type Statement interface {
	Type() Type
}

type Instruction interface {
	Statement
	// Flags are the flags passed to
	Flags() map[string]string
	Arguments() []string
}
