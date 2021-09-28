package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dekkagaijin/go-dockerfile/statements"
)

const (
	DefaultExcapeCharacter = '\\'
	WindowsEscapeCharacter = '`'
)

var canHaveContinuation = map[statements.Type]bool{
	statements.AddType:         true,
	statements.ArgType:         true,
	statements.CmdType:         true,
	statements.CopyType:        true,
	statements.EntrypointType:  true,
	statements.EnvType:         true,
	statements.ExposeType:      true,
	statements.FromType:        true,
	statements.HealthcheckType: true,
	statements.LabelType:       true,
	statements.MaintainerType:  true,
	statements.OnbuildType:     true,
	statements.RunType:         true,
	statements.ShellType:       true,
	statements.StopSignalType:  true,
	statements.UserType:        true,
	statements.VolumeType:      true,
	statements.WorkdirType:     true,
}

type Sequential struct {
	escapeCharacter rune
	statements      []statements.Statement
	directives      map[string]string
}

func NewSequentialParser() *Sequential {
	return &Sequential{escapeCharacter: DefaultExcapeCharacter}
}

func (p *Sequential) Parse(lines []string) ([]statements.Statement, rune, error) {
	totalLines := len(lines)
	if totalLines == 0 {
		return nil, 0, errors.New("dockerfile was empty")
	}
	p.statements = nil
	p.directives = map[string]string{}

	remainingLines, err := p.parseParserDirectives(lines)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read parser directives: %w", err)
	}

	for len(remainingLines) > 0 {
		if blankLineMatcher.MatchString(remainingLines[0]) {
			// skip blank lines
			remainingLines = remainingLines[1:]
			continue
		}
		var err error
		remainingLines, err = p.parseStatement(remainingLines)
		if err != nil {
			lineNum := totalLines - len(remainingLines) + 1
			return nil, 0, fmt.Errorf("failed parsing statement on line %d: %w", lineNum, err)
		}
	}
	return p.statements, p.escapeCharacter, nil
}

const (
	CommentToken = "#"

	reStartOfLine           = "^"
	reDontCare              = ".*"
	reOptionalWhitespace    = "[[:space:]]*"
	reWhitespace            = "[[:space:]]+"
	reNotWhitespace         = "[[:^space:]]+"
	reNotWhitespaceOrEquals = "[^=[:space:]]+"
	reEndOfLine             = "$"
)

var (
	blankLineMatcher       = regexp.MustCompile(reStartOfLine + reOptionalWhitespace + reEndOfLine)
	instructionLineMatcher = regexp.MustCompile(
		reStartOfLine +
			reOptionalWhitespace +
			"(" + reNotWhitespace + ")" +
			reWhitespace +
			reNotWhitespace +
			reDontCare +
			reEndOfLine)
	commentLineMatcher = regexp.MustCompile(
		reStartOfLine +
			reOptionalWhitespace +
			CommentToken +
			"(" + reDontCare + ")" +
			reEndOfLine)
	parserDirectiveMatcher = regexp.MustCompile(
		reStartOfLine +
			reOptionalWhitespace +
			CommentToken +
			reOptionalWhitespace +
			"(" + reNotWhitespaceOrEquals + ")" + // key
			reOptionalWhitespace +
			"=" +
			reOptionalWhitespace +
			"(" + reNotWhitespaceOrEquals + ")" + // value
			reOptionalWhitespace +
			reEndOfLine)
)

// Dockerfile parser directives are a special type of comment of the format
// `# key=value` which occur at the beginning of the file.
// As soon as the parser encounters a blank line, an instruction,
// or a comment that does not match this form, it will treat all remaining comments as ordinary.
// See: https://docs.docker.com/engine/reference/builder/#parser-directives
func (p *Sequential) parseParserDirectives(lines []string) (remainingLines []string, err error) {
	remainingLines = lines
	for len(remainingLines) > 0 {
		matches := parserDirectiveMatcher.FindStringSubmatch(remainingLines[0])
		if len(matches) != 3 {
			break
		}
		k, v := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
		k = strings.ToLower(k)
		if _, alreadySet := p.directives[k]; alreadySet {
			return remainingLines, fmt.Errorf("directive %q set multiple times", k)
		}
		p.directives[k] = v
		remainingLines = remainingLines[1:]
	}
	return remainingLines, nil
}

func hasContinuation(line string, escapeCharacter rune) bool {
	// TODO: support long stretches of terminal escapes? This is consistent with buildkit's impl
	// https://github.com/moby/buildkit/blob/1031116f12ec6f80c11782c93a48891f848168b9/frontend/dockerfile/parser/parser.go#L164
	return strings.HasSuffix(line, string(escapeCharacter)) && !strings.HasSuffix(line, string(escapeCharacter)+string(escapeCharacter))
}

func (p *Sequential) parseGenericInstruction(st statements.Type, lines []string) (remainingLines []string, err error) {
	inst := &statements.GenericInstruction{InstructionType: st}
	remainingLines = lines
	terminated := false
	for len(remainingLines) > 0 && !terminated {
		currentLine := strings.TrimSpace(remainingLines[0])
		remainingLines = remainingLines[1:]
		if currentLine == "" {
			continue
		}
		if commentLineMatcher.MatchString(currentLine) {
			inst.Lines = append(inst.Lines, currentLine)
			continue
		}
		terminated = true

		if canHaveContinuation[st] && hasContinuation(currentLine, p.escapeCharacter) {
			terminated = false
			currentLine = strings.TrimSuffix(currentLine, string(p.escapeCharacter))
			currentLine = strings.TrimSpace(currentLine)
		}
		inst.Lines = append(inst.Lines, currentLine)
		if len(inst.Lines) == 1 {
			// Remove the command from the first line of args
			currentLine = currentLine[len(string(st)):]
		}
		inst.Args = append(inst.Args, strings.Fields(currentLine)...)
	}
	p.statements = append(p.statements, inst)
	return remainingLines, nil
}

func (p *Sequential) parseStatement(lines []string) (remainingLines []string, err error) {
	if commentLineMatcher.MatchString(lines[0]) {
		return p.parseComment(lines)
	}
	return p.parseInstruction(lines)
}

func (p *Sequential) parseInstruction(lines []string) (remainingLines []string, err error) {
	currentLine := strings.TrimSpace(lines[0])
	reMatches := instructionLineMatcher.FindStringSubmatch(currentLine)
	if len(reMatches) != 2 {
		return lines, fmt.Errorf("syntax error: %q", currentLine)
	}
	instruction := statements.Type(strings.ToUpper(reMatches[1]))
	if !statements.KnownTypes[instruction] {
		return lines, fmt.Errorf("unknown instruction: %q", instruction)
	}
	switch instruction {
	default:
		return p.parseGenericInstruction(instruction, lines)
	}
}

func (p *Sequential) parseComment(lines []string) (remainingLines []string, err error) {
	remainingLines = lines
	cmnt := &statements.Comment{}
	for len(remainingLines) > 0 && commentLineMatcher.MatchString(remainingLines[0]) {
		curr := strings.TrimSpace(remainingLines[0])
		curr = strings.TrimPrefix(curr, CommentToken) // remove leading "#"
		cmnt.Lines = append(cmnt.Lines, curr)
		remainingLines = remainingLines[1:]
	}
	p.statements = append(p.statements, cmnt)
	return remainingLines, nil
}
