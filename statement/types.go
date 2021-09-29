package statement

type Type string

const (
	ADDType         Type = "ADD"
	ARGType         Type = "ARG"
	CMDType         Type = "CMD"
	CommentType     Type = "#"
	COPYType        Type = "COPY"
	ENTRYPOINTType  Type = "ENTRYPOINT"
	ENVType         Type = "ENV"
	EXPOSEType      Type = "EXPOSE"
	FROMType        Type = "FROM"
	HEALTHCHECKType Type = "HEALTHCHECK"
	LABELType       Type = "LABEL"
	MAINTAINERType  Type = "MAINTAINER"
	ONBUILDType     Type = "ONBUILD"
	RUNType         Type = "RUN"
	SHELLType       Type = "SHELL"
	STOPSIGNALType  Type = "STOPSIGNAL"
	USERType        Type = "USER"
	VOLUMEType      Type = "VOLUME"
	WORKDIRType     Type = "WORKDIR"
)

var Known = map[Type]bool{
	ADDType:         true,
	ARGType:         true,
	CMDType:         true,
	CommentType:     true,
	COPYType:        true,
	ENTRYPOINTType:  true,
	ENVType:         true,
	EXPOSEType:      true,
	FROMType:        true,
	HEALTHCHECKType: true,
	LABELType:       true,
	MAINTAINERType:  true,
	ONBUILDType:     true,
	RUNType:         true,
	SHELLType:       true,
	STOPSIGNALType:  true,
	USERType:        true,
	VOLUMEType:      true,
	WORKDIRType:     true,
}

// TypeMatchers is a mapping of statement types to matching regular expressions.
// var TypeMatchers = map[Type]*regexp.Regexp{
// 	ADDType:         instructionMatcher(ADDType),
// 	ARGType:         instructionMatcher(ARGType),
// 	CMDType:         instructionMatcher(CMDType),
// 	CommentType:     regexp.MustCompile(reStartOfLine + reMaybeWhitespace + string(CommentType) + ".*" + reEndOfLine),
// 	COPYType:        instructionMatcher(COPYType),
// 	ENTRYPOINTType:  instructionMatcher(ENTRYPOINTType),
// 	ENVType:         instructionMatcher(ENVType),
// 	EXPOSEType:      instructionMatcher(EXPOSEType),
// 	FROMType:        instructionMatcher(FROMType),
// 	HEALTHCHECKType: instructionMatcher(HEALTHCHECKType),
// 	LABELType:       instructionMatcher(LABELType),
// 	MAINTAINERType:  instructionMatcher(MAINTAINERType),
// 	ONBUILDType:     instructionMatcher(ONBUILDType),
// 	RUNType:         instructionMatcher(RUNType),
// 	SHELLType:       instructionMatcher(SHELLType),
// 	STOPSIGNALType:  instructionMatcher(STOPSIGNALType),
// 	USERType:        instructionMatcher(USERType),
// 	VOLUMEType:      instructionMatcher(VOLUMEType),
// 	WORKDIRType:     instructionMatcher(WORKDIRType),
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
