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

type sequentialParser struct {
	escapeCharacter rune
	ast             *AST
	directives      map[string]string
}

func (p *sequentialParser) Parse(lines []string) (*AST, error) {
	totalLines := len(lines)
	if totalLines == 0 {
		return nil, errors.New("dockerfile was empty")
	}
	p.ast = &AST{}
	p.directives = map[string]string{}

	remainingLines, err := p.parseParserDirectives(lines)
	if err != nil {
		return nil, fmt.Errorf("failed to read parser directives: %w", err)
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
			return nil, fmt.Errorf("failed parsing statement on line %d: %w", lineNum, err)
		}
	}
	return p.ast, nil
}

const (
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
			dockerfileCommentToken +
			"(" + reDontCare + ")" +
			reEndOfLine)
	parserDirectiveMatcher = regexp.MustCompile(
		reStartOfLine +
			reOptionalWhitespace +
			dockerfileCommentToken +
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
func (p *sequentialParser) parseParserDirectives(lines []string) (remainingLines []string, err error) {
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
		//TODO: it's not safe in general to get rid of whitespace between args
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
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	sp := sequentialParser{escapeCharacter: DefaultExcapeCharacter}
	return sp.Parse(lines)
}
