package statement

type GenericInstruction struct {
	InstructionType Type

	Args []string

	// Lines are all of the input lines of the statement, minus newline escape characters and leading/trailing whitespace.
	// This includes the instruction name in the first line, as well as any comment lines.
	Lines []string
}

func (i *GenericInstruction) Type() Type {
	return i.InstructionType
}

func (*GenericInstruction) Flags() map[string]string {
	return nil
}

func (i *GenericInstruction) Arguments() []string {
	return i.Args
}
