package statement

type ArgInstruction struct {
	Name       string
	DefaultVal string
}

func (*ArgInstruction) Type() Type {
	return ARG
}

func (*ArgInstruction) Flags() map[string]string {
	return nil
}

func (i *ArgInstruction) Arguments() Arguments {
	arg := i.Name
	if i.DefaultVal != "" {
		arg += "=" + i.DefaultVal
	}
	return Arguments{
		List: []string{arg},
	}
}
