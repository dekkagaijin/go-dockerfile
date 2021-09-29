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

	EscapeParserDirectiveKey = "escape"
)

type sequentialstatementScannerFor func(lines []string, escapeCharacter rune) (stmt statements.Statement, remainingLines []string, err error)

// var canHaveContinuation = map[statements.Type]bool{
// 	statements.AddType:         true,
// 	statements.ArgType:         true,
// 	statements.CmdType:         true,
// 	statements.CommentType:     false,
// 	statements.CopyType:        true,
// 	statements.EntrypointType:  true,
// 	statements.EnvType:         true,
// 	statements.ExposeType:      true,
// 	statements.FromType:        true,
// 	statements.HealthcheckType: true,
// 	statements.LabelType:       true,
// 	statements.MaintainerType:  true,
// 	statements.OnbuildType:     true,
// 	statements.RunType:         true,
// 	statements.ShellType:       true,
// 	statements.StopSignalType:  true,
// 	statements.UserType:        true,
// 	statements.VolumeType:      true,
// 	statements.WorkdirType:     true,
// }

var statementScannerFor = map[statements.Type]sequentialstatementScannerFor{
	statements.AddType: scanGenericInstruction,
	statements.ArgType: scanGenericInstruction,
	statements.CmdType: scanGenericInstruction,
	//statements.CommentType:     nil,
	statements.CopyType:        scanGenericInstruction,
	statements.EntrypointType:  scanGenericInstruction,
	statements.EnvType:         scanGenericInstruction,
	statements.ExposeType:      scanGenericInstruction,
	statements.FromType:        scanGenericInstruction,
	statements.HealthcheckType: scanGenericInstruction,
	statements.LabelType:       scanGenericInstruction,
	statements.MaintainerType:  scanGenericInstruction,
	statements.OnbuildType:     scanGenericInstruction,
	statements.RunType:         scanGenericInstruction,
	statements.ShellType:       scanGenericInstruction,
	statements.StopSignalType:  scanGenericInstruction,
	statements.UserType:        scanGenericInstruction,
	statements.VolumeType:      scanGenericInstruction,
	statements.WorkdirType:     scanGenericInstruction,
}

type Sequential struct {
	escapeCharacter rune
	statements      []statements.Statement
	directives      map[string]string
}

func (p *Sequential) Parse(lines []string) ([]statements.Statement, rune, error) {
	totalLines := len(lines)
	if totalLines == 0 {
		return nil, 0, errors.New("dockerfile was empty")
	}
	p.escapeCharacter = DefaultExcapeCharacter
	p.statements = nil
	p.directives = map[string]string{}

	directives, remainingLines, err := scanParserDirectives(lines)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read parser directives: %w", err)
	}
	p.directives = directives
	if val, exists := p.directives[EscapeParserDirectiveKey]; exists {
		vRunes := []rune(val)
		if len(vRunes) > 1 || !(vRunes[0] == DefaultExcapeCharacter || vRunes[0] == WindowsEscapeCharacter) {
			return nil, 0, fmt.Errorf("escape token must be one of [%q, %q], got %q", DefaultExcapeCharacter, WindowsEscapeCharacter, val)
		}
		p.escapeCharacter = vRunes[0]
	}

	for len(remainingLines) > 0 {
		if blankLineMatcher.MatchString(remainingLines[0]) {
			// skip blank lines
			remainingLines = remainingLines[1:]
			continue
		}
		var stmt statements.Statement
		var err error
		stmt, remainingLines, err = scanStatement(remainingLines, p.escapeCharacter)
		if err != nil {
			lineNum := totalLines - len(remainingLines) + 1
			return nil, 0, fmt.Errorf("failed parsing statement on line %d: %w", lineNum, err)
		}
		p.statements = append(p.statements, stmt)
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
func scanParserDirectives(lines []string) (directives map[string]string, remainingLines []string, err error) {
	directives = map[string]string{}
	remainingLines = lines
	for len(remainingLines) > 0 {
		matches := parserDirectiveMatcher.FindStringSubmatch(remainingLines[0])
		if len(matches) != 3 {
			break
		}
		k, v := strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
		k = strings.ToLower(k)
		if _, alreadySet := directives[k]; alreadySet {
			return nil, lines, fmt.Errorf("directive %q set multiple times", k)
		}
		directives[k] = v
		remainingLines = remainingLines[1:]
	}
	return directives, remainingLines, nil
}

func hasContinuation(line string, escapeCharacter rune) bool {
	// TODO: support long stretches of terminal escapes? This is consistent with buildkit's impl
	// https://github.com/moby/buildkit/blob/1031116f12ec6f80c11782c93a48891f848168b9/frontend/dockerfile/parser/parser.go#L164
	return strings.HasSuffix(line, string(escapeCharacter)) && !strings.HasSuffix(line, string(escapeCharacter)+string(escapeCharacter))
}

func scanStatement(lines []string, escapeCharacter rune) (stmt statements.Statement, remainingLines []string, err error) {

	if commentLineMatcher.MatchString(lines[0]) {
		return scanComment(lines)
	}
	return scanInstruction(lines, escapeCharacter)
}

func scanComment(lines []string) (stmt statements.Statement, remainingLines []string, err error) {
	remainingLines = lines
	cmnt := &statements.Comment{}
	for len(remainingLines) > 0 && commentLineMatcher.MatchString(remainingLines[0]) {
		curr := strings.TrimSpace(remainingLines[0])
		curr = strings.TrimPrefix(curr, CommentToken) // remove leading "#"
		cmnt.Lines = append(cmnt.Lines, curr)
		remainingLines = remainingLines[1:]
	}
	return cmnt, remainingLines, nil
}

func scanInstruction(lines []string, escapeCharacter rune) (stmt statements.Statement, remainingLines []string, err error) {
	currentLine := strings.TrimSpace(lines[0])
	reMatches := instructionLineMatcher.FindStringSubmatch(currentLine)
	if len(reMatches) != 2 {
		return nil, lines, fmt.Errorf("syntax error: %q", currentLine)
	}
	instruction := statements.Type(strings.ToUpper(reMatches[1]))
	if !statements.KnownTypes[instruction] {
		return nil, lines, fmt.Errorf("unknown instruction: %q", instruction)
	}

	if scanInstruction, exists := statementScannerFor[instruction]; exists {
		return scanInstruction(lines, escapeCharacter)
	}
	return scanGenericInstruction(lines, escapeCharacter)
}

func trimContinuation(line string, escapeCharacter rune) string {
	line = strings.TrimSuffix(line, string(escapeCharacter))
	return strings.TrimSpace(line)
}

func scanGenericInstruction(lines []string, escapeCharacter rune) (stmt statements.Statement, remainingLines []string, err error) {
	inst := &statements.GenericInstruction{}

	currentLine := strings.TrimSpace(lines[0])
	remainingLines = lines[1:]

	terminated := true
	if hasContinuation(currentLine, escapeCharacter) {
		terminated = false
		currentLine = trimContinuation(currentLine, escapeCharacter)
	}

	args := strings.Fields(currentLine)
	inst.InstructionType = statements.Type(strings.ToUpper(args[0]))
	inst.Lines = append(inst.Lines, currentLine)
	inst.Args = append(inst.Args, args[1:]...)

	for !terminated && len(remainingLines) > 0 {
		currentLine = strings.TrimSpace(remainingLines[0])
		remainingLines = remainingLines[1:]
		if currentLine == "" {
			continue
		}
		if commentLineMatcher.MatchString(currentLine) {
			inst.Lines = append(inst.Lines, currentLine)
			continue
		}

		if hasContinuation(currentLine, escapeCharacter) {
			currentLine = trimContinuation(currentLine, escapeCharacter)
		} else {
			terminated = true
		}

		inst.Lines = append(inst.Lines, currentLine)
		inst.Args = append(inst.Args, strings.Fields(currentLine)...)
	}
	if !terminated {
		return nil, lines, errors.New("multi-line statement is not terminated")
	}
	return inst, remainingLines, nil
}
