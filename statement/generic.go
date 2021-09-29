package statement

type TODO struct {
	InstructionType Type

	Args []string

	// Lines are all of the input lines of the statement, minus newline escape characters and leading/trailing whitespace.
	// This includes the instruction name in the first line, as well as any comment lines.
	Lines []string
}

func (i *TODO) Type() Type {
	return i.InstructionType
}

func (*TODO) Flags() map[string]string {
	return nil
}

func (i *TODO) Arguments() []string {
	return i.Args
}
