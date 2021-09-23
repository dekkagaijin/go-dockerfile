package dockerfile

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func mustOpen(t *testing.T, path string) io.ReadCloser {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("could not open file: %v", err)
	}
	return f
}

func TestRoundtrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		originalPath string
		expectedPath string
	}{{
		desc: "basic",

		originalPath: "testdata/Dockerfile.basic",
		expectedPath: "testdata/rendered/Dockerfile.basic",
	}, {
		desc: "multistage",

		originalPath: "testdata/Dockerfile.multistage",
		expectedPath: "testdata/rendered/Dockerfile.multistage",
		//TODO: pick up https://github.com/moby/buildkit/pull/2375
		// }, {
		// 	desc: "apache2",

		// 	originalPath: "testdata/Dockerfile.withapache2",
		// 	expectedPath: "testdata/rendered/Dockerfile.withapache2",
	}}
	for _, tc := range testCases {
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
			expected := string(expectedBytes)
			rendered := parsed.String()
			if diff := cmp.Diff(expected, rendered); diff != "" {
				t.Error("mismatch (-want +got)\n", diff)
			}
		})
	}

}
