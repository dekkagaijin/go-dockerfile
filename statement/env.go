package statement

type EnvInstruction struct {
	Env      map[string]string
	KeyOrder []string
}

func (*EnvInstruction) Type() Type {
	return ENV
}

func (*EnvInstruction) Flags() map[string]string {
	return nil
}

func (i *EnvInstruction) Arguments() Arguments {
	args := make([]string, 0, len(i.Env))
	for _, k := range i.KeyOrder {
		args = append(args, k+"="+i.Env[k])
	}
	return Arguments{
		List: args,
	}
}
