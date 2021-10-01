package dockerfile

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveE2E(t *testing.T) {
	testCases := []struct {
		desc          string
		originalPath  string
		buildArg, env map[string]string

		expectedPath string
	}{
		{
			desc:         "gauntlet",
			originalPath: "testdata/resolve/gauntlet/Dockerfile",
			buildArg: map[string]string{
				"ARG1": "val1",
				"ARG2": "val2",
				"ARG3": "val3",
				"ARG4": "val4",
				"FOO":  "foo value",

				"RUNTIME_IMAGE": "runtime-image@sha256:0bf474896363505e5ea5e5d6ace8ebfb13a760a409b1fb467d428fc716f9f284",
			},
			env:          map[string]string{},
			expectedPath: "testdata/resolve/gauntlet/Dockerfile.resolved",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			originalFile := mustOpen(t, tc.originalPath)
			defer originalFile.Close()
			expectedFile := mustOpen(t, tc.expectedPath)
			defer expectedFile.Close()

			expectedBytes, err := ioutil.ReadAll(expectedFile)
			if err != nil {
				t.Fatalf("failed to read expected testdata file: %v", err)
			}

			parsed, err := Parse(originalFile)
			if err != nil {
				t.Fatalf("Parse() error'd: %v", err)
			}

			resolved, err := Resolve(parsed, tc.buildArg, tc.env)
			if err != nil {
				t.Fatalf("Resolve() error'd: %v", err)
			}

			sb := strings.Builder{}
			if err := Render(resolved, &sb); err != nil {
				t.Fatalf("Render() error'd: %v", err)
			}

			expected := string(expectedBytes)
			rendered := sb.String()
			if diff := cmp.Diff(expected, rendered); diff != "" {
				t.Error("mismatch (-want +got):\n", diff)
			}
		})
	}
}
