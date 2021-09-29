package statement

type Add struct {
	InstructionType Type

	Args Arguments

	// Lines are all of the input lines of the statement, minus newline escape characters and leading/trailing whitespace.
	// This includes the instruction name in the first line, as well as any comment lines.
	Lines []string
}

func (i *Add) Type() Type {
	return i.InstructionType
}

func (*Add) Flags() map[string]string {
	return nil
}

func (i *Add) Arguments() Arguments {
	return i.Args
}
