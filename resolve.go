package dockerfile

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dekkagaijin/go-dockerfile/internal/parser"
	"github.com/dekkagaijin/go-dockerfile/statement"
)

type scopedVars struct {
	ARG, ENV map[string]string
}

type buildStage []statement.Statement

type resolver struct {
	escapeCharacter rune
	input, global   scopedVars
}

func (r *resolver) Resolve(df *Parsed) (*Parsed, error) {
	r.escapeCharacter = df.EscapeCharacter
	r.global.ARG = map[string]string{}
	// r.global.ENV = map[string]string{} // no such thing as ENV outside of a build stage

	statements := df.Statements

	resolved := Parsed{
		EscapeCharacter: df.EscapeCharacter,
	}

	// Consume the preamble before the first build stage...
	var argTombstones *statement.Comment // this is an ugly hack to get the renderer to group the comment lines together
	for ; len(statements) > 0 && statements[0].Type() != statement.FROM; statements = statements[1:] {
		stmt := statements[0]
		if arg, isARGStatement := stmt.(*statement.ArgInstruction); isARGStatement {
			var err error
			cmnt, err := r.resolveArgInstruction(arg, r.global)
			if err != nil {
				return nil, err
			}
			if argTombstones == nil {
				argTombstones = cmnt
			} else {
				argTombstones.Lines = append(argTombstones.Lines, cmnt.Lines...)
			}
			continue
		}
		if argTombstones != nil {
			resolved.Statements = append(resolved.Statements, argTombstones)
			argTombstones = nil
		}
		resolved.Statements = append(resolved.Statements, stmt)
	}
	if argTombstones != nil {
		resolved.Statements = append(resolved.Statements, argTombstones)
		argTombstones = nil
	}
	if len(statements) == 0 {
		return nil, errors.New("dockerfile did not contain a `FROM` statement")
	}

	// Group the statements into build stages.
	stages := []buildStage{}
	for _, stmt := range statements {
		if stmt.Type() == statement.FROM {
			stages = append(stages, buildStage{})
		}
		stages[len(stages)-1] = append(stages[len(stages)-1], stmt)
	}
	for _, stage := range stages {
		resolvedStmts, err := r.resolveBuildStage(stage)
		if err != nil {
			return nil, err
		}
		resolved.Statements = append(resolved.Statements, resolvedStmts...)
	}

	return &resolved, nil
}

func (r *resolver) resolveBuildStage(stage buildStage) ([]statement.Statement, error) {
	local := scopedVars{
		ARG: map[string]string{},
		ENV: map[string]string{},
	}
	var resolved []statement.Statement
	for _, stmt := range stage {
		stmt, err := r.resolveStatement(stmt, local)
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			resolved = append(resolved, stmt)
		}
	}
	return resolved, nil
}

func (r *resolver) resolveStatement(stmt statement.Statement, local scopedVars) (statement.Statement, error) {
	switch inst := stmt.(type) {
	case *statement.FromInstruction:
		local.ARG = map[string]string{}
		local.ENV = map[string]string{}
		return r.resolveFromInstruction(inst)
	case *statement.ArgInstruction:
		return r.resolveArgInstruction(inst, local)
	case *statement.EnvInstruction:
		return r.resolveEnvInstruction(inst, local)
	case *statement.AddInstruction:
		return r.resolveAddInstruction(inst, local), nil
	case *statement.GenericInstruction:
		return r.resolveGenericInstruction(inst, local), nil
	}
	return stmt, nil
}

func (r *resolver) resolveFromInstruction(raw *statement.FromInstruction) (*statement.FromInstruction, error) {
	return &statement.FromInstruction{
		Platform: substituteVars(raw.Platform, r.global.ARG),
		Image:    substituteVars(raw.Image, r.global.ARG),
		Alias:    raw.Alias,
	}, nil
}

func (r *resolver) resolveArgInstruction(arg *statement.ArgInstruction, scope scopedVars) (*statement.Comment, error) {
	cmnt := &statement.Comment{}
	originalStatement := "ARG " + arg.Name
	if arg.DefaultVal != "" {
		originalStatement += "=" + arg.DefaultVal
	}
	if val, global := r.global.ARG[arg.Name]; global {
		scope.ARG[arg.Name] = val
		cmnt.Lines = []string{fmt.Sprintf(" `%s` was resolved to `%s=%s` from prior declaration.", originalStatement, arg.Name, val)}
		return cmnt, nil
	}
	if val, provided := r.input.ARG[arg.Name]; provided {
		scope.ARG[arg.Name] = val
		cmnt.Lines = []string{fmt.Sprintf(" `%s` was resolved to `%s=%s` from build argument.", originalStatement, arg.Name, val)}
		return cmnt, nil
	}
	if arg.DefaultVal != "" {
		scope.ARG[arg.Name] = arg.DefaultVal
		cmnt.Lines = []string{fmt.Sprintf(" `%s` was resolved to `%s=%s` from default value.", originalStatement, arg.Name, arg.DefaultVal)}
		return cmnt, nil
	}
	return nil, fmt.Errorf("could not resolve `%s`, did not have a default or provided value", originalStatement)
}

func (r *resolver) resolveEnvInstruction(raw *statement.EnvInstruction, scope scopedVars) (*statement.EnvInstruction, error) {
	resolved := &statement.EnvInstruction{
		Env:      make(map[string]string, len(raw.Env)),
		KeyOrder: make([]string, 0, len(raw.KeyOrder)),
	}
	for _, key := range raw.KeyOrder {
		val := substituteVars(raw.Env[key], scope.ENV, scope.ARG)
		val = parser.EnsureModernEnvVal(val, r.escapeCharacter)
		resolved.KeyOrder = append(resolved.KeyOrder, key)
		resolved.Env[key] = val
		scope.ENV[key] = val
	}
	return resolved, nil
}

func (r *resolver) resolveGenericInstruction(raw *statement.GenericInstruction, scope scopedVars) *statement.GenericInstruction {
	resolved := &statement.GenericInstruction{
		InstructionType: raw.InstructionType,
		Args: statement.Arguments{
			Execable: raw.Args.Execable,
		},
	}
	for _, arg := range raw.Args.List {
		resolved.Args.List = append(resolved.Args.List, substituteVars(arg, scope.ENV, scope.ARG))
	}
	for _, line := range raw.Lines {
		resolved.Lines = append(resolved.Lines, substituteVars(line, scope.ENV, scope.ARG))
	}
	return resolved
}

func (r *resolver) resolveAddInstruction(raw *statement.AddInstruction, scope scopedVars) *statement.AddInstruction {
	resolved := &statement.AddInstruction{
		Args: statement.Arguments{
			Execable: raw.Args.Execable,
		},
	}
	for _, arg := range raw.Args.List {
		resolved.Args.List = append(resolved.Args.List, substituteVars(arg, scope.ENV, scope.ARG))
	}
	for _, line := range raw.Lines {
		resolved.Lines = append(resolved.Lines, substituteVars(line, scope.ENV, scope.ARG))
	}
	return resolved
}

// `substituteVars` expands environment variables present in the given string.
// Multiple environments may be passed, which will be searched in order of decreasing preference.
//
// Dockerfiles support several forms of variable replacement:
//
// a) `$<name>`
// b) `${<name>}`
// c) `${<name>:-<value>}`
// d) `${<name>:+<value>}`
//
// `<value>` can be any string, including other env vars.
//
// See: https://docs.docker.com/engine/reference/builder/#environment-replacement
func substituteVars(s string, scopes ...map[string]string) string {
	return os.Expand(s, func(ph string) string {
		if split := strings.SplitN(ph, ":+", 2); len(split) > 1 {
			key := split[0]
			valIfSet := split[1]
			valIfSet = substituteVars(valIfSet, scopes...)
			if _, isSet := lookupEnvVar(key, scopes...); isSet {
				return valIfSet
			}
			return ""
		}

		key := ph
		defaultVal := ""

		if split := strings.SplitN(ph, ":-", 2); len(split) > 1 {
			key = split[0]
			defaultVal = split[1]
			defaultVal = substituteVars(defaultVal, scopes...)
		}
		if val, isSet := lookupEnvVar(key, scopes...); isSet {
			return val
		}
		return defaultVal
	})
}

func lookupEnvVar(key string, scopes ...map[string]string) (string, bool) {
	for _, scope := range scopes {
		if val, isSet := scope[key]; isSet {
			return val, true
		}
	}
	return "", false
}

// Resolve resolves the given values in the Dockerfile.
func Resolve(df *Parsed, buildArg, env map[string]string) (*Parsed, error) {
	r := resolver{
		input: scopedVars{
			ARG: buildArg,
			ENV: env,
		},
	}
	return r.Resolve(df)
}
