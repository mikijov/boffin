package lib

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func TestUpdate(t *testing.T) {
	dir := filepath.Join(getTestDir(), "update", ".boffin")

	boffin, err := LoadBoffin(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = boffin.Update()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []*FileInfo{
		{
			Path:     "file1.ext",
			Size:     10,
			Time:     parseTime("2020-02-06T13:56:51.571756332Z"),
			Checksum: "mv0rsY4Lof04c4eVesQRoggxIMQBLzv82jX0gglIhrI=",
			History: []*FileEvent{
				&FileEvent{
					Type:     "changed",
					Time:     parseTime("2020-02-06T13:56:51.571756332Z"),
					Checksum: "mv0rsY4Lof04c4eVesQRoggxIMQBLzv82jX0gglIhrI=",
				},
			},
		},
		{
			Path:     "sub1/file2.ext",
			Size:     10,
			Time:     parseTime("2020-02-06T13:57:02.90203166Z"),
			Checksum: "vQTuoHT8OnxI9g7fcZnEeTC9jcbX1NuRsS4gyDQkxjE=",
			History: []*FileEvent{
				&FileEvent{
					Type:     "changed",
					Time:     parseTime("2020-02-06T13:57:02.90203166Z"),
					Checksum: "vQTuoHT8OnxI9g7fcZnEeTC9jcbX1NuRsS4gyDQkxjE=",
				},
			},
		},
		{
			Path:     "sub1/file3.ext",
			Size:     10,
			Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
			Checksum: "4PFd3bElTqFi8wvTlY2eRK6sJo65UivdK95nd7it5h4=",
			History: []*FileEvent{
				&FileEvent{
					Type:     "changed",
					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
					Checksum: "4PFd3bElTqFi8wvTlY2eRK6sJo65UivdK95nd7it5h4=",
				},
			},
		},
		{
			Path:     "sub1/file4.ext",
			Size:     0,
			Time:     time.Now(),
			Checksum: "",
			History: []*FileEvent{
				&FileEvent{
					Type:     "changed",
					Time:     parseTime("2020-02-06T13:59:21.099018324Z"),
					Checksum: "71JuQzM1k9ZV2tMnnhemjf+FUfbEEs8YS170IORPpA4=",
				},
				&FileEvent{
					Type: "deleted",
					Time: time.Now(),
				},
			},
		},
		{
			Path:     "sub1/file5.ext",
			Size:     10,
			Time:     parseTime("2020-02-07T21:01:11.11974727Z"),
			Checksum: "Z12qAGMLMXMmfBWqZw8LHTJD2Ifpp8AMJYmCa4eMYac=",
			History: []*FileEvent{
				&FileEvent{
					Type:     "changed",
					Time:     parseTime("2020-02-07T21:01:11.11974727Z"),
					Checksum: "Z12qAGMLMXMmfBWqZw8LHTJD2Ifpp8AMJYmCa4eMYac=",
				},
			},
		},
	}
	actual := boffin.GetFiles()

	margin, _ := time.ParseDuration("2s")
	opt1 := cmpopts.EquateApproxTime(margin)
	opt2 := cmpopts.IgnoreUnexported(FileInfo{})

	if diff := cmp.Diff(expected, actual, opt1, opt2); diff != "" {
		t.Errorf("file.History:\n%s", diff)
	}
}
