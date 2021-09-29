package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

const (
	DefaultExcapeCharacter = '\\'
	WindowsEscapeCharacter = '`'

	EscapeParserDirectiveKey = "escape"
)

type statementScanFn func(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error)

// var canHaveContinuation = map[statement.Type]bool{
// 	statement.ADDType:         true,
// 	statement.ARGType:         true,
// 	statement.CMDType:         true,
// 	statement.CommentType:     false,
// 	statement.COPYType:        true,
// 	statement.ENTRYPOINTType:  true,
// 	statement.ENVType:         true,
// 	statement.EXPOSEType:      true,
// 	statement.FROMType:        true,
// 	statement.HEALTHCHECKType: true,
// 	statement.LABELType:       true,
// 	statement.MAINTAINERType:  true,
// 	statement.ONBUILDType:     true,
// 	statement.RUNType:         true,
// 	statement.SHELLType:       true,
// 	statement.STOPSIGNALType:  true,
// 	statement.USERType:        true,
// 	statement.VOLUMEType:      true,
// 	statement.WORKDIRType:     true,
// }

var statementScannerFor = map[statement.Type]statementScanFn{
	statement.CommentType: func(lines []string, _ rune) (stmt statement.Statement, remainingLines []string, err error) {
		// comments do not escape characters
		return scanComment(lines)
	},
	statement.ADDType:         scanADD,
	statement.ARGType:         scanARG,
	statement.CMDType:         scanCMD,
	statement.COPYType:        scanCOPY,
	statement.ENTRYPOINTType:  scanENTRYPOINT,
	statement.ENVType:         scanENV,
	statement.EXPOSEType:      scanEXPOSE,
	statement.FROMType:        scanFROM,
	statement.HEALTHCHECKType: scanHEALTHCHECK,
	statement.LABELType:       scanLABEL,
	statement.MAINTAINERType:  scanMAINTAINER,
	statement.ONBUILDType:     scanONBUILD,
	statement.RUNType:         scanRUN,
	statement.SHELLType:       scanSHELL,
	statement.STOPSIGNALType:  scanSTOPSIGNAL,
	statement.USERType:        scanUSER,
	statement.VOLUMEType:      scanVOLUME,
	statement.WORKDIRType:     scanWORKDIR,
}

type Sequential struct {
	escapeCharacter rune
	statements      []statement.Statement
	directives      map[string]string
}

func (p *Sequential) Parse(lines []string) ([]statement.Statement, rune, error) {
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
		var stmt statement.Statement
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

func scanStatement(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	if commentLineMatcher.MatchString(lines[0]) {
		return scanComment(lines)
	}
	return scanInstruction(lines, escapeCharacter)
}

func scanInstruction(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	currentLine := strings.TrimSpace(lines[0])
	reMatches := instructionLineMatcher.FindStringSubmatch(currentLine)
	if len(reMatches) != 2 {
		return nil, lines, fmt.Errorf("syntax error: %q", currentLine)
	}
	instruction := statement.Type(strings.ToUpper(reMatches[1]))
	if !statement.Known[instruction] {
		return nil, lines, fmt.Errorf("unknown instruction: %q", instruction)
	}

	if scanInstruction, exists := statementScannerFor[instruction]; exists {
		return scanInstruction(lines, escapeCharacter)
	}
	return scanGenericInstruction(lines, escapeCharacter)
}

func scanGenericInstruction(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	inst := &statement.GenericInstruction{}
	st, rawArgs, statementLines, remainingLines, err := scanInstructionLines(lines, escapeCharacter)
	if err != nil {
		return nil, lines, err
	}
	inst.InstructionType = st
	inst.Lines = statementLines
	inst.Args = strings.Fields(rawArgs) // TODO accept JSON list?
	return inst, remainingLines, nil
}

func scanInstructionLines(lines []string, escapeCharacter rune) (st statement.Type, rawArgs string, statementLines, remainingLines []string, err error) {
	currentLine := strings.TrimSpace(lines[0])
	remainingLines = lines[1:]

	reMatches := instructionLineMatcher.FindStringSubmatch(currentLine)
	if len(reMatches) != 2 {
		return statement.Type(""), "", nil, lines, fmt.Errorf("syntax error: %q", currentLine)
	}

	st = statement.Type(strings.ToUpper(reMatches[1]))
	statementLines = append(statementLines, currentLine)

	currentLine = strings.TrimSpace(currentLine[len(string(st)):]) // trim command from first line

	terminated := true
	if hasContinuation(currentLine, escapeCharacter) {
		terminated = false
		currentLine = strings.TrimSuffix(currentLine, string(escapeCharacter))
	}

	rawArgs = currentLine

	for !terminated && len(remainingLines) > 0 {
		currentLine = strings.TrimSpace(remainingLines[0])
		remainingLines = remainingLines[1:]
		if currentLine == "" {
			// Ignore blank lines in multi-line statement.
			continue
		}
		if commentLineMatcher.MatchString(currentLine) {
			// Interstitial comment lines are allowed in multi-line comments.
			statementLines = append(statementLines, currentLine)
			continue
		}

		if hasContinuation(currentLine, escapeCharacter) {
			currentLine = strings.TrimSuffix(currentLine, string(escapeCharacter))
		} else {
			terminated = true
		}

		rawArgs += currentLine
		statementLines = append(statementLines, currentLine)
	}
	if !terminated {
		return statement.Type(""), "", nil, lines, errors.New("multi-line statement does not terminate")
	}
	return st, rawArgs, statementLines, remainingLines, nil
}
