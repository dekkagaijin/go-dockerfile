package statement

type AddInstruction struct {
	Args Arguments

	// Lines are all of the input lines of the statement, minus newline escape characters and leading/trailing whitespace.
	// This includes the instruction name in the first line, as well as any comment lines.
	Lines []string
}

func (i *AddInstruction) Type() Type {
	return ADD
}

func (*AddInstruction) Flags() map[string]string {
	return nil
}

func (i *AddInstruction) Arguments() Arguments {
	return i.Args
}
