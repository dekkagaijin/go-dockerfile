package statement

type Type string

const (
	ADD         Type = "ADD"
	ARG         Type = "ARG"
	CMD         Type = "CMD"
	CommentType Type = "#"
	COPY        Type = "COPY"
	ENTRYPOINT  Type = "ENTRYPOINT"
	ENV         Type = "ENV"
	EXPOSE      Type = "EXPOSE"
	FROM        Type = "FROM"
	HEALTHCHECK Type = "HEALTHCHECK"
	LABEL       Type = "LABEL"
	MAINTAINER  Type = "MAINTAINER"
	ONBUILD     Type = "ONBUILD"
	RUN         Type = "RUN"
	SHELL       Type = "SHELL"
	STOPSIGNAL  Type = "STOPSIGNAL"
	USER        Type = "USER"
	VOLUME      Type = "VOLUME"
	WORKDIR     Type = "WORKDIR"
)

var Known = map[Type]bool{
	ADD:         true,
	ARG:         true,
	CMD:         true,
	CommentType: true,
	COPY:        true,
	ENTRYPOINT:  true,
	ENV:         true,
	EXPOSE:      true,
	FROM:        true,
	HEALTHCHECK: true,
	LABEL:       true,
	MAINTAINER:  true,
	ONBUILD:     true,
	RUN:         true,
	SHELL:       true,
	STOPSIGNAL:  true,
	USER:        true,
	VOLUME:      true,
	WORKDIR:     true,
}

// TypeMatchers is a mapping of statement types to matching regular expressions.
// var TypeMatchers = map[Type]*regexp.Regexp{
// 	ADD:         instructionMatcher(ADD),
// 	ARG:         instructionMatcher(ARG),
// 	CMD:         instructionMatcher(CMD),
// 	CommentType:     regexp.MustCompile(reStartOfLine + reMaybeWhitespace + string(CommentType) + ".*" + reEndOfLine),
// 	COPY:        instructionMatcher(COPY),
// 	ENTRYPOINT:  instructionMatcher(ENTRYPOINT),
// 	ENV:         instructionMatcher(ENV),
// 	EXPOSE:      instructionMatcher(EXPOSE),
// 	FROM:        instructionMatcher(FROM),
// 	HEALTHCHECK: instructionMatcher(HEALTHCHECK),
// 	LABEL:       instructionMatcher(LABEL),
// 	MAINTAINER:  instructionMatcher(MAINTAINER),
// 	ONBUILD:     instructionMatcher(ONBUILD),
// 	RUN:         instructionMatcher(RUN),
// 	SHELL:       instructionMatcher(SHELL),
// 	STOPSIGNAL:  instructionMatcher(STOPSIGNAL),
// 	USER:        instructionMatcher(USER),
// 	VOLUME:      instructionMatcher(VOLUME),
// 	WORKDIR:     instructionMatcher(WORKDIR),
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

type AddInstruction TODO
type ArgInstruction TODO
type CmdInstruction TODO
type CopyInstruction TODO
type EntrypointInstruction TODO
type EnvInstruction TODO
type ExposeInstruction TODO
type FromInstruction TODO
type HealthcheckInstruction TODO
type LabelInstruction TODO
type MaintainerInstruction TODO
type OnBuildInstruction TODO
type RunInstruction TODO
type ShellInstruction TODO
type StopSignalInstruction TODO
type UserInstruction TODO
type VolumeInstruction TODO
type WorkdirInstruction TODO
