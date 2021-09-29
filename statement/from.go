package statement

type FromInstruction struct {
	Platform string
	Image    string
	Alias    string
}

func (*FromInstruction) Type() Type {
	return FROM
}

func (i *FromInstruction) Flags() map[string]string {
	if i.Platform != "" {
		return map[string]string{
			"platform": i.Platform,
		}
	}
	return nil
}

func (i *FromInstruction) Arguments() Arguments {
	args := []string{i.Image}
	if i.Alias != "" {
		args = append(args, "AS", i.Alias)
	}
	return Arguments{
		List: args,
	}
}
