package lib

import (
	"os"
	"path/filepath"
	"sort"
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

type result struct {
	Result string
	Local  []string
	Remote []string
}

type testAction struct {
	Result []*result
}

func (t *testAction) Unchanged(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "unchanged",
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) MetaDataChanged(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "metadata",
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) Moved(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "moved",
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) LocalOnly(localFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "local-only",
		Local:  []string{localFile.Path()},
		Remote: nil,
	})
}

func (t *testAction) LocalOld(localFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "local-old",
		Local:  []string{localFile.Path()},
		Remote: nil,
	})
}

func (t *testAction) RemoteOnly(remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "remote-only",
		Local:  nil,
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) RemoteOld(remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "remote-old",
		Local:  nil,
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) LocalDeleted(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "local-deleted",
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) RemoteDeleted(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "remote-deleted",
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) LocalChanged(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "local-changed",
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) RemoteChanged(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "remote-changed",
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) ConflictPath(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: "conflict",
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) ConflictHash(localFiles, remoteFiles []*FileInfo) {
	local := []string{}
	for _, file := range localFiles {
		local = append(local, file.Path())
	}
	sort.Strings(local)

	remote := []string{}
	for _, file := range remoteFiles {
		remote = append(remote, file.Path())
	}
	sort.Strings(remote)

	t.Result = append(t.Result, &result{
		Result: "conflict",
		Local:  local,
		Remote: remote,
	})
}

func (t *testAction) Sort() {
	// sort results according to local path names
	sort.Slice(t.Result, func(i, j int) bool {
		left := t.Result[i]
		right := t.Result[j]

		if left.Result != right.Result {
			return left.Result < right.Result
		}

		for index := 0; index < len(left.Local); index++ {
			if index >= len(right.Local) {
				return false
			}
			if left.Local[index] != right.Local[index] {
				return left.Local[index] < right.Local[index]
			}
		}
		if len(left.Local) < len(right.Local) {
			return true
		}

		for index := 0; index < len(left.Remote); index++ {
			if index >= len(right.Remote) {
				return false
			}
			if left.Remote[index] != right.Remote[index] {
				return left.Remote[index] < right.Remote[index]
			}
		}
		if len(left.Remote) < len(right.Remote) {
			return true
		}

		return false
	})
}

func TestFindBoffin(t *testing.T) {
	testRoot := getTestDir()

	dir := testRoot
	actual, err := FindBoffinDir(dir)
	if actual != "" {
		t.Errorf("expecting '' but got '%s'", dir)
	}
	if err == nil {
		t.Error("expected error but got none")
	} else if err.Error() != "could not find .boffin dir" {
		t.Errorf("expecting 'could not find .boffin dir' but got '%s'", err.Error())
	}

	dir = filepath.Join(testRoot, "find-boffin")
	expected := filepath.Join(testRoot, "find-boffin", ".boffin")
	actual, err = FindBoffinDir(dir)
	if actual != expected {
		t.Errorf("expecting '%s' but got '%s'", expected, actual)
	}
	if err != nil {
		t.Errorf("did not expect error but got '%s'", err.Error())
	}

	dir = filepath.Join(testRoot, "find-boffin", "sub0")
	expected = filepath.Join(testRoot, "find-boffin", ".boffin")
	actual, err = FindBoffinDir(dir)
	if actual != expected {
		t.Errorf("expecting '%s' but got '%s'", expected, actual)
	}
	if err != nil {
		t.Errorf("did not expect error but got '%s'", err.Error())
	}

	dir = filepath.Join(testRoot, "find-boffin", "sub0", "sub2")
	expected = filepath.Join(testRoot, "find-boffin", ".boffin")
	actual, err = FindBoffinDir(dir)
	if actual != expected {
		t.Errorf("expecting '%s' but got '%s'", expected, actual)
	}
	if err != nil {
		t.Errorf("did not expect error but got '%s'", err.Error())
	}
}

func TestLoadBoffin(t *testing.T) {
	dir := filepath.Join(getTestDir(), "load-boffin", ".boffin")

	boffin, err := LoadBoffin(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files := boffin.GetFiles()
	if len(files) != 1 {
		t.Errorf("GetFiles: 1 != %d", len(files))
	} else {
		file := files[0]
		{
			expected := "dir/file.ext"
			if file.Path() != expected {
				t.Errorf("file.Path: '%s' != '%s'", expected, file.Path())
			}
		}
		{
			expected := int64(12345)
			if file.Size() != expected {
				t.Errorf("file.Size: '%d' != '%d'", expected, file.Size())
			}
		}
		{

			expected := parseTime("2006-01-02T15:04:05Z")
			if file.Time() != expected {
				t.Errorf("file.Time: '%v' != '%v'", expected, file.Time())
			}
		}
		{
			expected := "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz"
			if file.Checksum() != expected {
				t.Errorf("file.Checksum: '%s' != '%s'", expected, file.Checksum())
			}
		}
		{
			expected := []*FileEvent{
				&FileEvent{
					Path: "dir/file.ext",
					Time: parseTime("2006-01-01T15:04:05Z"),
				},
				&FileEvent{
					Path:     "dir/file.ext",
					Size:     12345,
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
	if boffin.GetDbDir() != expected {
		t.Errorf("GetDbDir: '%s' != '%s'", expected, boffin.GetDbDir())
	}

	expected = filepath.Dir(dir)
	if boffin.GetBaseDir() != expected {
		t.Errorf("GetBaseDir: '%s' != '%s'", expected, boffin.GetBaseDir())
	}

	expected = filepath.Join(filepath.Dir(dir), "import")
	if boffin.GetImportDir() != expected {
		t.Errorf("GetImportDir: '%s' != '%s'", expected, boffin.GetImportDir())
	}
}
