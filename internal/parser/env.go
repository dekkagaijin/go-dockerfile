package parser

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/dekkagaijin/go-dockerfile/statement"
)

var envArgsMatcher = regexp.MustCompile(
	reStartOfLine +
		reOptionalWhitespace +
		"(" + reNotWhitespace + "(?:" + reWhitespace + reNotWhitespace + ")*)" + // args minus leading/trailing whitespace
		reOptionalWhitespace +
		reEndOfLine)

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

	reMatches := envArgsMatcher.FindStringSubmatch(rawArgs)
	if len(reMatches) != 2 {
		return nil, lines, fmt.Errorf("syntax error, could not parse ENV args: %q", rawArgs)
	}
	argsStr := reMatches[1]
	env, keyOrder, err := parseEnvArgs(argsStr, escapeCharacter)
	if err != nil {
		return nil, lines, fmt.Errorf("syntax error in ENV args: %w", err)
	}
	return &statement.EnvInstruction{
		Env:      env,
		KeyOrder: keyOrder,
	}, remainingLines, nil
}

func removeLeadingWhitespace(runes []rune) []rune {
	for len(runes) > 0 && unicode.IsSpace(runes[0]) {
		runes = runes[1:]
	}
	return runes
}

// This is so gross. I'm sorry.
func parseEnvArgs(argsStr string, escapeCharacter rune) (map[string]string, []string, error) {
	argsStr = strings.TrimSpace(argsStr)
	unparsed := []rune(argsStr)
	keyOrder := []string{}

	partialKey := []rune{}
	for len(unparsed) > 0 && !(unparsed[0] == '=' || unicode.IsSpace(unparsed[0])) {
		partialKey = append(partialKey, unparsed[0])
		unparsed = unparsed[1:]
	}
	if len(unparsed) == 0 {
		return nil, nil, fmt.Errorf("args were not of the form `key value` or `key=value ...`: %q", argsStr)
	}
	if unparsed[0] != '=' {
		// This is the legacy form: `key value`
		unparsed = removeLeadingWhitespace(unparsed)
		val := string(unparsed)
		if i := strings.IndexFunc(val, unicode.IsSpace); i != -1 {
			// val contains whitespace, needs to be quoted or escaped
			val = `"` + val + `"`
		}
		key := string(partialKey)
		keyOrder = append(keyOrder, key)
		return map[string]string{
			key: val, // add quotes to unquoted value
		}, keyOrder, nil
	}

	unparsed = unparsed[1:] // remove leading '='

	key := string(partialKey)
	partialKey = partialKey[:0]

	parsed := map[string]string{}

	partialVal := []rune{}
	var currentQuote rune = 0
	for len(unparsed) > 0 {
		char := unparsed[0]
		unparsed = unparsed[1:]

		if key == "" {
			// Need to build the key string...

			if unicode.IsSpace(char) {
				if len(partialKey) > 0 {
					return nil, nil, fmt.Errorf("whitespace is not permitted between key strings and '=': %q", argsStr)
				}
				continue
			}

			if char == '=' {
				if len(partialKey) == 0 {
					return nil, nil, fmt.Errorf("env keys cannot be blank: %q", argsStr)
				}
				key = string(partialKey)
				partialKey = partialKey[:0]
				continue
			}

			partialKey = append(partialKey, char)
			continue
		}

		// Key has been built. Need to build the value string...

		if unicode.IsSpace(char) {
			if currentQuote != 0 {
				// Spaces are allowed in quoted strings.
				partialVal = append(partialVal, char)
			} else if len(partialVal) > 0 {
				// Spaces terminate unquoted strings.
				parsed[key] = string(partialVal)
				keyOrder = append(keyOrder, key)

				// Reset all of the state vars.
				key = ""
				partialVal = partialVal[:0]
				currentQuote = 0
			}
			continue
		}

		if char == '\'' || char == '"' {
			partialVal = append(partialVal, char)
			if char == currentQuote {
				// We've parsed a full quoted key-value pair.
				parsed[key] = string(partialVal)
				keyOrder = append(keyOrder, key)

				// Reset all of the state vars.
				key = ""
				partialVal = partialVal[:0]
				currentQuote = 0
				continue
			}
			currentQuote = char
			continue
		}

		partialVal = append(partialVal, char)
	}

	return parsed, keyOrder, nil
}
