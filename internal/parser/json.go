package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

type JSONParseErr error

func parseJSONStringList(line string) ([]string, error) {
	var strList []string
	if err := json.NewDecoder(strings.NewReader(line)).Decode(&strList); err != nil {
		return nil, JSONParseErr(fmt.Errorf("could not be parsed as JSON string list: %q", line))
	}
	return strList, nil
}
