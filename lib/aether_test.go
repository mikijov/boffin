package lib

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func getTestDir() string {
	dir, _ := os.Getwd()
	dir = filepath.Join(dir, "../test")
	dir = filepath.Clean(dir)
	return dir
}

func parseTime(s string) time.Time {
	retVal, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return retVal
}

func TestFindAether(t *testing.T) {
	testRoot := getTestDir()

	dir := testRoot
	actual, err := FindAetherDir(dir)
	if actual != "" {
		t.Errorf("expecting '' but got '%s'", dir)
	}
	if err == nil {
		t.Error("expected error but got none")
	} else if err.Error() != "could not find .aether dir" {
		t.Errorf("expecting 'could not find .aether dir' but got '%s'", err.Error())
	}

	dir = filepath.Join(testRoot, "find-aether")
	expected := filepath.Join(testRoot, "find-aether", ".aether")
	actual, err = FindAetherDir(dir)
	if actual != expected {
		t.Errorf("expecting '%s' but got '%s'", expected, actual)
	}
	if err != nil {
		t.Errorf("did not expect error but got '%s'", err.Error())
	}

	dir = filepath.Join(testRoot, "find-aether", "sub1")
	expected = filepath.Join(testRoot, "find-aether", ".aether")
	actual, err = FindAetherDir(dir)
	if actual != expected {
		t.Errorf("expecting '%s' but got '%s'", expected, actual)
	}
	if err != nil {
		t.Errorf("did not expect error but got '%s'", err.Error())
	}

	dir = filepath.Join(testRoot, "find-aether", "sub1", "sub2")
	expected = filepath.Join(testRoot, "find-aether", ".aether")
	actual, err = FindAetherDir(dir)
	if actual != expected {
		t.Errorf("expecting '%s' but got '%s'", expected, actual)
	}
	if err != nil {
		t.Errorf("did not expect error but got '%s'", err.Error())
	}
}

func TestLoadAether(t *testing.T) {
	dir := filepath.Join(getTestDir(), "load-aether", ".aether")

	aether, err := LoadAether(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := aether.GetFiles()
	if len(files) != 1 {
		t.Errorf("GetFiles: 1 != %d", len(files))
	} else {
		file := files[0]
		{
			expected := "dir/file.ext"
			if file.Path != expected {
				t.Errorf("file.Path: '%s' != '%s'", expected, file.Path)
			}
		}
		{
			expected := int64(12345)
			if file.Size != expected {
				t.Errorf("file.Size: '%d' != '%d'", expected, file.Size)
			}
		}
		{

			expected := parseTime("2006-01-02T15:04:05Z")
			if file.Time != expected {
				t.Errorf("file.Time: '%v' != '%v'", expected, file.Time)
			}
		}
		{
			expected := "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz"
			if file.Checksum != expected {
				t.Errorf("file.Checksum: '%s' != '%s'", expected, file.Checksum)
			}
		}
		{
			expected := []*FileEvent{
				&FileEvent{
					Type: "deleted",
					Time: parseTime("2006-01-01T15:04:05Z"),
				},
				&FileEvent{
					Type:     "changed",
					Time:     parseTime("2006-01-02T15:04:05Z"),
					Checksum: "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz",
				},
			}
			actual := file.History
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("file.History:\n%s", diff)
			}
		}
	}

	expected := dir
	if aether.GetDbDir() != expected {
		t.Errorf("GetDbDir: '%s' != '%s'", expected, aether.GetDbDir())
	}

	expected = filepath.Dir(dir)
	if aether.GetBaseDir() != expected {
		t.Errorf("GetBaseDir: '%s' != '%s'", expected, aether.GetBaseDir())
	}

	expected = filepath.Join(filepath.Dir(dir), "import")
	if aether.GetImportDir() != expected {
		t.Errorf("GetImportDir: '%s' != '%s'", expected, aether.GetImportDir())
	}
}
