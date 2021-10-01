package dockerfile

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func TestRenderE2E(t *testing.T) {
	testCases := map[string]struct {
		originalPath string
		expectedPath string
	}{}

	if err := filepath.WalkDir("testdata/render", func(path string, d fs.DirEntry, err error) error {
		if !(d.IsDir() || strings.HasSuffix(path, ".rendered")) {
			tc := testCases[path]
			tc.originalPath = path
			tc.expectedPath = path + ".rendered"
			testCases[path] = tc
		}
		return nil
	}); err != nil {
		t.Fatalf("failed to load testdata: %v", err)
	}

	if len(testCases) == 0 {
		t.Fatal("failed to load testdata")
	}

	for desc, tc := range testCases {
		tc := tc
		t.Run(desc, func(t *testing.T) {
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
			sb := strings.Builder{}
			if err := Render(parsed, &sb); err != nil {
				t.Fatalf("Render() error'd: %v", err)
			}
			rendered := sb.String()
			if diff := cmp.Diff(expected, rendered); diff != "" {
				t.Error("mismatch (-want +got):\n", diff)
			}
		})
	}
}
