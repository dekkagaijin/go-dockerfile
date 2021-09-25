package dockerfile

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	DefaultExcapeCharacter = '\\'
	WindowsEscapeCharacter = '`'

	dockerfileCommentToken = "#"
)

type Parser struct {
	escapeCharacter rune
}

var defaultParser = Parser{escapeCharacter: DefaultExcapeCharacter}

func NewParser(escapeCharacter rune) (*Parser, error) {
	if escapeCharacter != DefaultExcapeCharacter && escapeCharacter != WindowsEscapeCharacter {
		return nil, fmt.Errorf("escapeCharacter must be one of [%q, %q]", DefaultExcapeCharacter, WindowsEscapeCharacter)
	}
	return &Parser{escapeCharacter: escapeCharacter}, nil
}

func (p Parser) Parse(file io.Reader) (*AST, error) {
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	sp := sequentialParser{escapeCharacter: p.escapeCharacter}
	return sp.Parse(lines)
}

type sequentialParser struct {
	escapeCharacter rune
	ast             *AST
}

func (p *sequentialParser) Parse(lines []string) (*AST, error) {
	totalLines := len(lines)
	if totalLines == 0 {
		return nil, errors.New("dockerfile was empty")
	}
	p.ast = &AST{}

	remainingLines := lines

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
			return nil, fmt.Errorf("error parsing statement on line %d: %w", lineNum, err)
		}
	}
	return p.ast, nil
}

const (
	reStartOfLine        = "^"
	reDontCare           = ".*"
	reOptionalWhitespace = "[[:space:]]*"
	reWhitespace         = "[[:space:]]+"
	reNonWhitespace      = "[[:^space:]]+"
	reEndOfLine          = "$"
)

var (
	blankLineMatcher       = regexp.MustCompile(reStartOfLine + reOptionalWhitespace + reEndOfLine)
	instructionLineMatcher = regexp.MustCompile(
		reStartOfLine +
			reOptionalWhitespace +
			"(" + reNonWhitespace + ")" +
			reWhitespace +
			reNonWhitespace +
			reDontCare +
			reEndOfLine)
	commentLineMatcher = regexp.MustCompile(
		reStartOfLine +
			reOptionalWhitespace +
			dockerfileCommentToken +
			"(" + reDontCare + ")" +
			reEndOfLine)
)

func (p *sequentialParser) parseGenericInstruction(st StatementType, lines []string) (remainingLines []string, err error) {
	inst := &TODOInstruction{Type: st}
	remainingLines = lines
	terminated := false
	for len(remainingLines) > 0 && !terminated {
		currentLine := strings.TrimSpace(remainingLines[0])
		remainingLines = remainingLines[1:]
		if currentLine == "" {
			continue
		}
		if commentLineMatcher.MatchString(currentLine) {
			inst.Raw = append(inst.Raw, currentLine)
			continue
		}
		terminated = true
		if strings.HasSuffix(currentLine, string(p.escapeCharacter)) {
			terminated = false
			currentLine = strings.TrimSuffix(currentLine, string(p.escapeCharacter))
		}
		inst.Raw = append(inst.Raw, currentLine)
		if len(inst.Raw) == 1 {
			// Remove the command from the first line of args
			currentLine = currentLine[len(string(st)):]
		}
		inst.Args = append(inst.Args, strings.Fields(currentLine)...)
	}
	p.ast.Statements = append(p.ast.Statements, inst)
	return remainingLines, nil
}

func (p *sequentialParser) parseStatement(lines []string) (remainingLines []string, err error) {
	if commentLineMatcher.MatchString(lines[0]) {
		return p.parseComment(lines)
	}
	return p.parseInstruction(lines)
}

func (p *sequentialParser) parseInstruction(lines []string) (remainingLines []string, err error) {
	currentLine := strings.TrimSpace(lines[0])
	reMatches := instructionLineMatcher.FindStringSubmatch(currentLine)
	if len(reMatches) != 2 {
		return lines, fmt.Errorf("syntax error: %q", currentLine)
	}
	instruction := StatementType(strings.ToUpper(reMatches[1]))
	if _, known := KnownStatementTypes[instruction]; !known {
		return lines, fmt.Errorf("unknown instruction: %q", instruction)
	}
	switch instruction {
	default:
		return p.parseGenericInstruction(instruction, lines)
	}
}

func (p *sequentialParser) parseComment(lines []string) (remainingLines []string, err error) {
	remainingLines = lines
	cmnt := &Comment{}
	for len(remainingLines) > 0 && commentLineMatcher.MatchString(remainingLines[0]) {
		curr := strings.TrimSpace(remainingLines[0])
		curr = strings.TrimPrefix(curr, dockerfileCommentToken) // remove leading "#"
		cmnt.Lines = append(cmnt.Lines, curr)
		remainingLines = remainingLines[1:]
	}
	p.ast.Statements = append(p.ast.Statements, cmnt)
	return remainingLines, nil
}

// Parse parses the given Dockerfile into an AST.
func Parse(file io.Reader) (*AST, error) {
	return defaultParser.Parse(file)
}
