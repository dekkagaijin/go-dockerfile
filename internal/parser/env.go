package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

var envArgMatcher = regexp.MustCompile(
	reStartOfLine +
		reKeyValuePair + // one k-v pair at beginning of line
		"(?:" + reWhitespace + "|" + reEndOfLine + ")") // whitespace or EOL

/*
ENV instructions declare environment variables in the container's context.
There are 2 forms:

	a) `ENV <key> <value>`
	b) `ENV <key>=<value> ...`

The first "legacy" form defines exactly one env key-value pair separated
by whitespace:

	`ENV ONE TWO= THREE=world` results in `env["ONE"] = "TWO= THREE=world"`

The second form allows for multiple env vars to be defined at once.
Values in this form may be single-quoted, double-quoted, or escaped to
preserve whitespace. For example, these statements are equivalent:

	```
	ENV MY_NAME="John Doe"
	ENV MY_DOG=Rex\ The\ Dog
	ENV MY_CAT='fluffy kitty'
	```

and

	`ENV MY_NAME="John Doe" MY_DOG=Rex\ The\ Dog MY_CAT='fluffy kitty'`

See: https://docs.docker.com/engine/reference/builder/#env
*/
func scanENV(lines []string, escapeCharacter rune) (stmt statement.Statement, remainingLines []string, err error) {
	st, rawArgs, statementLines, remainingLines, err := scanInstructionLines(lines, escapeCharacter)
	if err != nil {
		return nil, lines, err
	}
	if st != statement.ENV {
		return nil, lines, fmt.Errorf("not an ENV statement: %q", statementLines[0])
	}

	rawArgs = strings.TrimSpace(rawArgs)
	if rawArgs == "" {
		return nil, lines, errors.New("ENV requires arguments")
	}

	if !envArgMatcher.MatchString(rawArgs) {
		// This is the legacy form
		key := strings.Fields(rawArgs)[0]
		rawVal := rawArgs[len(key):]
		return &statement.EnvInstruction{
			Env:      map[string]string{key: EnsureModernEnvVal(rawVal, escapeCharacter)},
			KeyOrder: []string{key},
		}, remainingLines, nil
	}

	inst := &statement.EnvInstruction{
		Env: map[string]string{},
	}
	unparsed := rawArgs
	for unparsed != "" {
		argMatch := envArgMatcher.FindStringSubmatch(unparsed)
		if len(argMatch) == 0 {
			return nil, lines, fmt.Errorf("could not parse next `key=value`: %q", unparsed)
		}
		key := argMatch[1]
		inst.Env[key] = argMatch[2]
		inst.KeyOrder = append(inst.KeyOrder, key)
		unparsed = unparsed[len(argMatch[0]):]
	}
	return inst, remainingLines, nil
}

func EnsureModernEnvVal(rawVal string, escapeCharacter rune) string {
	val := strings.TrimSpace(rawVal)
	if i := strings.IndexFunc(val, unicode.IsSpace); i != -1 {
		if !((strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`)) || (strings.HasPrefix(val, `'`) && strings.HasSuffix(val, `'`))) {
			// val contains whitespace, needs to be quoted or escaped
			val = `"` + val + `"`
		}
	}
	return val
}
