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

type result struct {
	Result int
	Local  []string
	Remote []string
}

type testAction struct {
	Result []*result
}

func (t *testAction) Equal(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: DiffEqual,
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) LocalAdded(localFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: DiffLocalAdded,
		Local:  []string{localFile.Path()},
		Remote: nil,
	})
}

func (t *testAction) RemoteAdded(remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: DiffRemoteAdded,
		Local:  nil,
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) LocalDeleted(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: DiffLocalDeleted,
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) RemoteDeleted(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: DiffRemoteDeleted,
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) LocalChanged(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: DiffLocalChanged,
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) RemoteChanged(localFile, remoteFile *FileInfo) {
	t.Result = append(t.Result, &result{
		Result: DiffLocalChanged,
		Local:  []string{localFile.Path()},
		Remote: []string{remoteFile.Path()},
	})
}

func (t *testAction) Conflict(localFiles, remoteFiles []*FileInfo) {
	local := []string{}
	for _, file := range localFiles {
		local = append(local, file.Path())
	}

	remote := []string{}
	for _, file := range remoteFiles {
		remote = append(remote, file.Path())
	}

	t.Result = append(t.Result, &result{
		Result: DiffConflict,
		Local:  local,
		Remote: remote,
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
	boffin.Sort()

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
					Type: "deleted",
					Time: parseTime("2006-01-01T15:04:05Z"),
				},
				&FileEvent{
					Path:     "dir/file.ext",
					Size:     12345,
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

// func TestUpdate(t *testing.T) {
// 	dir := filepath.Join(getTestDir(), "update2", ".boffin")
//
// 	boffin, err := LoadBoffin(dir)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
// 	err = boffin.Update(nil)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
// 	boffin.Sort()
//
// 	expected := []*FileInfo{
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "equal.ext",
// 					Size:     10,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-06T13:56:51.571756332Z"),
// 					Checksum: "mv0rsY4Lof04c4eVesQRoggxIMQBLzv82jX0gglIhrI=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "multi-hash/a-equal.ext",
// 					Size:     30,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T14:03:45.270218079Z"),
// 					Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "multi-hash/changed.ext",
// 					Size:     87,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-24T14:01:07.828877429Z"),
// 					Checksum: "OfhhlgGfv8NbaPqLE6EGi2c9EJR8edzvXFnxl3QUvYo=",
// 				},
// 				&FileEvent{
// 					Path:     "multi-hash/changed.ext",
// 					Size:     87,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T14:01:07.828877429Z"),
// 					Checksum: "ofhhlgGfv8NbaPqLE6EGi2c9EJR8edzvXFnxl3QUvYo=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "multi-hash/m-equal.ext",
// 					Size:     30,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T14:03:57.917265752Z"),
// 					Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
// 				},
// 			},
// 			checked: true,
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "multi-hash/moved.ext",
// 					Size:     30,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T14:02:41.831652522Z"),
// 					Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
// 				},
// 				&FileEvent{
// 					Path:     "multi-hash/moved/moved.ext",
// 					Size:     30,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T14:02:41.831652522Z"),
// 					Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "multi-hash/renamed-before.ext",
// 					Size:     30,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T14:03:45.270218079Z"),
// 					Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
// 				},
// 				&FileEvent{
// 					Path:     "multi-hash/renamed.ext",
// 					Size:     30,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T14:03:45.270218079Z"),
// 					Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "multi-hash/z-equal.ext",
// 					Size:     30,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T14:04:02.354066271Z"),
// 					Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "sub1/cross-rename-2.ext",
// 					Size:     19,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
// 					Checksum: "Iag4g9z39+jJOdVGxqCNXziaAFwFZ8dnfdMrZQz1qKM=",
// 				},
// 				&FileEvent{
// 					Path:     "sub1/cross-rename-1.ext",
// 					Size:     19,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
// 					Checksum: "Iag4g9z39+jJOdVGxqCNXziaAFwFZ8dnfdMrZQz1qKM=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "sub1/cross-rename-1.ext",
// 					Size:     19,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
// 					Checksum: "K1L3GOGZF5wiOtiJdkN6+xZiAKwG77ueF+KnMyCXAuI=",
// 				},
// 				&FileEvent{
// 					Path:     "sub1/cross-rename-2.ext",
// 					Size:     19,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
// 					Checksum: "K1L3GOGZF5wiOtiJdkN6+xZiAKwG77ueF+KnMyCXAuI=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "sub1/deleted.ext",
// 					Size:     10,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-06T13:59:21.099018324Z"),
// 					Checksum: "71JuQzM1k9ZV2tMnnhemjf+FUfbEEs8YS170IORPpA4=",
// 				},
// 				&FileEvent{
// 					Path: "sub1/deleted.ext",
// 					Type: "deleted",
// 					Time: time.Now(),
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "sub1/equal.ext",
// 					Size:     10,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-06T13:57:02.90203166Z"),
// 					Checksum: "vQTuoHT8OnxI9g7fcZnEeTC9jcbX1NuRsS4gyDQkxjE=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "sub1/new.ext",
// 					Size:     10,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
// 					Checksum: "Z12qAGMLMXMmfBWqZw8LHTJD2Ifpp8AMJYmCa4eMYac=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "sub1/renamed-before.ext",
// 					Size:     10,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
// 					Checksum: "4PFd3bElTqFi8wvTlY2eRK6sJo65UivdK95nd7it5h4=",
// 				},
// 				&FileEvent{
// 					Path:     "sub1/renamed-after.ext",
// 					Size:     10,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
// 					Checksum: "4PFd3bElTqFi8wvTlY2eRK6sJo65UivdK95nd7it5h4=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "sub1/move-rename-before.ext",
// 					Size:     16,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-05T13:57:12.378926011Z"),
// 					Checksum: "Ir6w9XOc7mlfgjFEhsjZAdhiqNosCRCf9iqzt3o7ndY=",
// 				},
// 				&FileEvent{
// 					Path:     "sub2/move-rename.ext",
// 					Size:     16,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
// 					Checksum: "Ir6w9XOc7mlfgjFEhsjZAdhiqNosCRCf9iqzt3o7ndY=",
// 				},
// 			},
// 		},
// 		{
// 			History: []*FileEvent{
// 				&FileEvent{
// 					Path:     "sub1/moved.ext",
// 					Size:     10,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
// 					Checksum: "xP4lKAtsUEfiZZ+Z4wlwZ3yFIxq8w7PPdIBvNBzZhd4=",
// 				},
// 				&FileEvent{
// 					Path:     "sub2/moved.ext",
// 					Size:     10,
// 					Type:     "changed",
// 					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
// 					Checksum: "xP4lKAtsUEfiZZ+Z4wlwZ3yFIxq8w7PPdIBvNBzZhd4=",
// 				},
// 			},
// 		},
// 	}
// 	actual := boffin.GetFiles()
//
// 	margin, _ := time.ParseDuration("2s")
// 	opt1 := cmpopts.EquateApproxTime(margin)
// 	opt2 := cmpopts.IgnoreUnexported(FileInfo{})
// 	// opt3 := cmpopts.IgnoreFields(FileEvent{}, "Time")
//
// 	if diff := cmp.Diff(expected, actual, opt1, opt2); diff != "" {
// 		t.Errorf("file.History:\n%s", diff)
// 	}
// }

// func TestDiff(t *testing.T) {
// 	var local Boffin = &db{
// 		files: []*FileInfo{
// 			{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Path:     "equal",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-01T12:34:56Z"),
// 						Checksum: "equal-hash-1",
// 					},
// 					&FileEvent{
// 						Path:     "equal",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-02T12:34:56Z"),
// 						Checksum: "equal-hash-2",
// 					},
// 				},
// 			},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "local-added",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "local-added-hash-1",
// 			// 			},
// 			// 		},
// 			// 	},
// 			{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Path:     "local-changed",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-01T12:34:56Z"),
// 						Checksum: "local-changed-hash-1",
// 					},
// 					&FileEvent{
// 						Path:     "local-changed",
// 						Size:     20,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-02T12:34:56Z"),
// 						Checksum: "local-changed-hash-2",
// 					},
// 				},
// 			},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "local-deleted",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "local-deleted-hash-1",
// 			// 			},
// 			// 			&FileEvent{
// 			// 				Path: "local-deleted",
// 			// 				Type: "deleted",
// 			// 				Time: parseTime("2020-01-02T12:34:56Z"),
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "remote-changed",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "remote-changed-hash-1",
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "remote-deleted",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "remote-deleted-hash-1",
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "conflict",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "conflict-hash-1",
// 			// 			},
// 			// 			&FileEvent{
// 			// 				Path:     "conflict",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-02T12:34:56Z"),
// 			// 				Checksum: "conflict-hash-L",
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	// {
// 			// 	// 	History: []*FileEvent{
// 			// 	// 		&FileEvent{
// 			// 	// 			Path:     "duplicate-deleted-edge-exists",
// 			// 	// 			Size:     10,
// 			// 	// 			Type:     "changed",
// 			// 	// 			Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 	// 			Checksum: "duplicate-deleted-edge",
// 			// 	// 		},
// 			// 	// 	},
// 			// 	// },
// 		},
// 	}
// 	var remote Boffin = &db{
// 		files: []*FileInfo{
// 			{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Path:     "equal",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-01T12:34:56Z"),
// 						Checksum: "equal-hash-1",
// 					},
// 					&FileEvent{
// 						Path:     "equal",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-02T12:34:56Z"),
// 						Checksum: "equal-hash-2",
// 					},
// 				},
// 			},
// 			{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Path:     "local-changed",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-01T12:34:56Z"),
// 						Checksum: "local-changed-hash-1",
// 					},
// 				},
// 			},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "local-deleted",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "local-deleted-hash-1",
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "remote-added",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "remote-added-hash-1",
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "remote-changed",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "remote-changed-hash-1",
// 			// 			},
// 			// 			&FileEvent{
// 			// 				Path:     "remote-changed",
// 			// 				Size:     11,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-02T12:34:56Z"),
// 			// 				Checksum: "remote-changed-hash-2",
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "remote-deleted",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "remote-deleted-hash-1",
// 			// 			},
// 			// 			&FileEvent{
// 			// 				Path: "remote-deleted",
// 			// 				Type: "deleted",
// 			// 				Time: parseTime("2020-01-02T12:34:56Z"),
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	{
// 			// 		History: []*FileEvent{
// 			// 			&FileEvent{
// 			// 				Path:     "conflict",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 				Checksum: "conflict-hash-1",
// 			// 			},
// 			// 			&FileEvent{
// 			// 				Path:     "conflict",
// 			// 				Size:     10,
// 			// 				Type:     "changed",
// 			// 				Time:     parseTime("2020-01-03T12:34:56Z"),
// 			// 				Checksum: "conflict-hash-R",
// 			// 			},
// 			// 		},
// 			// 	},
// 			// 	// {
// 			// 	// 	History: []*FileEvent{
// 			// 	// 		&FileEvent{
// 			// 	// 			Path:     "duplicate-deleted-edge-deleted",
// 			// 	// 			Size:     10,
// 			// 	// 			Type:     "changed",
// 			// 	// 			Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 	// 			Checksum: "duplicate-deleted-edge",
// 			// 	// 		},
// 			// 	// 		&FileEvent{
// 			// 	// 			Path: "duplicate-deleted-edge-deleted",
// 			// 	// 			Type: "delated",
// 			// 	// 			Time: parseTime("2020-01-03T12:34:56Z"),
// 			// 	// 		},
// 			// 	// 	},
// 			// 	// },
// 			// 	// {
// 			// 	// 	History: []*FileEvent{
// 			// 	// 		&FileEvent{
// 			// 	// 			Path:     "duplicate-deleted-edge-exists",
// 			// 	// 			Size:     10,
// 			// 	// 			Type:     "changed",
// 			// 	// 			Time:     parseTime("2020-01-01T12:34:56Z"),
// 			// 	// 			Checksum: "duplicate-deleted-edge",
// 			// 	// 		},
// 			// 	// 	},
// 			// 	// },
// 		},
// 	}
//
// 	expected := []DiffResult{
// 		// {
// 		// 	Result: DiffConflict,
// 		// 	Local: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "conflict",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "conflict-hash-1",
// 		// 			},
// 		// 			&FileEvent{
// 		// 				Path:     "conflict",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-02T12:34:56Z"),
// 		// 				Checksum: "conflict-hash-L",
// 		// 			},
// 		// 		},
// 		// 	},
// 		// 	Remote: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "conflict",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "conflict-hash-1",
// 		// 			},
// 		// 			&FileEvent{
// 		// 				Path:     "conflict",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-03T12:34:56Z"),
// 		// 				Checksum: "conflict-hash-R",
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 		{
// 			Result: DiffEqual,
// 			Local: &FileInfo{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Path:     "equal",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-01T12:34:56Z"),
// 						Checksum: "equal-hash-1",
// 					},
// 					&FileEvent{
// 						Path:     "equal",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-02T12:34:56Z"),
// 						Checksum: "equal-hash-2",
// 					},
// 				},
// 			},
// 			Remote: &FileInfo{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Path:     "equal",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-01T12:34:56Z"),
// 						Checksum: "equal-hash-1",
// 					},
// 					&FileEvent{
// 						Path:     "equal",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-02T12:34:56Z"),
// 						Checksum: "equal-hash-2",
// 					},
// 				},
// 			},
// 		},
// 		// {
// 		// 	Result: DiffLocalAdded,
// 		// 	Local: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "local-added",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "local-added-hash-1",
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 		{
// 			Result: DiffLocalChanged,
// 			Local: &FileInfo{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Path:     "local-changed",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-01T12:34:56Z"),
// 						Checksum: "local-changed-hash-1",
// 					},
// 					&FileEvent{
// 						Path:     "local-changed",
// 						Size:     20,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-02T12:34:56Z"),
// 						Checksum: "local-changed-hash-2",
// 					},
// 				},
// 			},
// 			Remote: &FileInfo{
// 				History: []*FileEvent{
// 					&FileEvent{
// 						Path:     "local-changed",
// 						Size:     10,
// 						Type:     "changed",
// 						Time:     parseTime("2020-01-01T12:34:56Z"),
// 						Checksum: "local-changed-hash-1",
// 					},
// 				},
// 			},
// 		},
// 		// {
// 		// 	Result: DiffLocalDeleted,
// 		// 	Local: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "local-deleted",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "local-deleted-hash-1",
// 		// 			},
// 		// 			&FileEvent{
// 		// 				Path: "local-deleted",
// 		// 				Type: "deleted",
// 		// 				Time: parseTime("2020-01-02T12:34:56Z"),
// 		// 			},
// 		// 		},
// 		// 	},
// 		// 	Remote: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "local-deleted",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "local-deleted-hash-1",
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 		// {
// 		// 	Result: DiffRemoteAdded,
// 		// 	Remote: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "remote-added",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "remote-added-hash-1",
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 		// {
// 		// 	Result: DiffRemoteChanged,
// 		// 	Local: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "remote-changed",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "remote-changed-hash-1",
// 		// 			},
// 		// 		},
// 		// 	},
// 		// 	Remote: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "remote-changed",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "remote-changed-hash-1",
// 		// 			},
// 		// 			&FileEvent{
// 		// 				Path:     "remote-changed",
// 		// 				Size:     11,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-02T12:34:56Z"),
// 		// 				Checksum: "remote-changed-hash-2",
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 		// {
// 		// 	Result: DiffRemoteDeleted,
// 		// 	Local: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "remote-deleted",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "remote-deleted-hash-1",
// 		// 			},
// 		// 		},
// 		// 	},
// 		// 	Remote: &FileInfo{
// 		// 		History: []*FileEvent{
// 		// 			&FileEvent{
// 		// 				Path:     "remote-deleted",
// 		// 				Size:     10,
// 		// 				Type:     "changed",
// 		// 				Time:     parseTime("2020-01-01T12:34:56Z"),
// 		// 				Checksum: "remote-deleted-hash-1",
// 		// 			},
// 		// 			&FileEvent{
// 		// 				Path: "remote-deleted",
// 		// 				Type: "deleted",
// 		// 				Time: parseTime("2020-01-02T12:34:56Z"),
// 		// 			},
// 		// 		},
// 		// 	},
// 		// },
// 	}
//
// 	actual := local.Diff(remote)
//
// 	margin, _ := time.ParseDuration("2s")
// 	opt1 := cmpopts.EquateApproxTime(margin)
// 	opt2 := cmpopts.IgnoreUnexported(FileInfo{})
// 	// opt3 := cmpopts.IgnoreFields(FileEvent{}, "Time")
//
// 	if diff := cmp.Diff(expected, actual, opt1, opt2); diff != "" {
// 		t.Errorf("Diff2:\n%s", diff)
// 	}
// }

func TestDiff3(t *testing.T) {
	var local Boffin = &db{
		files: []*FileInfo{
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "equal",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "equal-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "equal2",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "equal2-hash-1",
					},
					&FileEvent{
						Path:     "equal2",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "equal2-hash-2",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "equal3",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "equal3-hash-1",
					},
					&FileEvent{
						Path:     "equal3",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "equal3-hash-2",
					},
					&FileEvent{
						Path:     "equal3",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-03T12:34:56Z"),
						Checksum: "equal3-hash-3",
					},
				},
			},
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "local-added",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "local-added-hash-1",
			// 			},
			// 		},
			// 	},
			// {
			// 	History: []*FileEvent{
			// 		&FileEvent{
			// 			Path:     "local-changed",
			// 			Size:     10,
			// 			Type:     "changed",
			// 			Time:     parseTime("2020-01-01T12:34:56Z"),
			// 			Checksum: "local-changed-hash-1",
			// 		},
			// 		&FileEvent{
			// 			Path:     "local-changed",
			// 			Size:     20,
			// 			Type:     "changed",
			// 			Time:     parseTime("2020-01-02T12:34:56Z"),
			// 			Checksum: "local-changed-hash-2",
			// 		},
			// 	},
			// },
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "local-deleted",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "local-deleted-hash-1",
			// 			},
			// 			&FileEvent{
			// 				Path: "local-deleted",
			// 				Type: "deleted",
			// 				Time: parseTime("2020-01-02T12:34:56Z"),
			// 			},
			// 		},
			// 	},
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "remote-changed",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "remote-changed-hash-1",
			// 			},
			// 		},
			// 	},
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "remote-deleted",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "remote-deleted-hash-1",
			// 			},
			// 		},
			// 	},
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "conflict",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "conflict-hash-1",
			// 			},
			// 			&FileEvent{
			// 				Path:     "conflict",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-02T12:34:56Z"),
			// 				Checksum: "conflict-hash-L",
			// 			},
			// 		},
			// 	},
			// 	// {
			// 	// 	History: []*FileEvent{
			// 	// 		&FileEvent{
			// 	// 			Path:     "duplicate-deleted-edge-exists",
			// 	// 			Size:     10,
			// 	// 			Type:     "changed",
			// 	// 			Time:     parseTime("2020-01-01T12:34:56Z"),
			// 	// 			Checksum: "duplicate-deleted-edge",
			// 	// 		},
			// 	// 	},
			// 	// },
		},
	}
	var remote Boffin = &db{
		files: []*FileInfo{
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "equal",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "equal-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "equal2",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "equal2-hash-1",
					},
					&FileEvent{
						Path:     "equal2",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "equal2-hash-2",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "equal3",
						Size:     10,
						Type:     "changed",
						Time:     parseTime("2020-01-03T12:34:56Z"),
						Checksum: "equal3-hash-3",
					},
				},
			},
			// {
			// 	History: []*FileEvent{
			// 		&FileEvent{
			// 			Path:     "local-changed",
			// 			Size:     10,
			// 			Type:     "changed",
			// 			Time:     parseTime("2020-01-01T12:34:56Z"),
			// 			Checksum: "local-changed-hash-1",
			// 		},
			// 	},
			// },
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "local-deleted",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "local-deleted-hash-1",
			// 			},
			// 		},
			// 	},
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "remote-added",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "remote-added-hash-1",
			// 			},
			// 		},
			// 	},
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "remote-changed",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "remote-changed-hash-1",
			// 			},
			// 			&FileEvent{
			// 				Path:     "remote-changed",
			// 				Size:     11,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-02T12:34:56Z"),
			// 				Checksum: "remote-changed-hash-2",
			// 			},
			// 		},
			// 	},
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "remote-deleted",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "remote-deleted-hash-1",
			// 			},
			// 			&FileEvent{
			// 				Path: "remote-deleted",
			// 				Type: "deleted",
			// 				Time: parseTime("2020-01-02T12:34:56Z"),
			// 			},
			// 		},
			// 	},
			// 	{
			// 		History: []*FileEvent{
			// 			&FileEvent{
			// 				Path:     "conflict",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-01T12:34:56Z"),
			// 				Checksum: "conflict-hash-1",
			// 			},
			// 			&FileEvent{
			// 				Path:     "conflict",
			// 				Size:     10,
			// 				Type:     "changed",
			// 				Time:     parseTime("2020-01-03T12:34:56Z"),
			// 				Checksum: "conflict-hash-R",
			// 			},
			// 		},
			// 	},
			// 	// {
			// 	// 	History: []*FileEvent{
			// 	// 		&FileEvent{
			// 	// 			Path:     "duplicate-deleted-edge-deleted",
			// 	// 			Size:     10,
			// 	// 			Type:     "changed",
			// 	// 			Time:     parseTime("2020-01-01T12:34:56Z"),
			// 	// 			Checksum: "duplicate-deleted-edge",
			// 	// 		},
			// 	// 		&FileEvent{
			// 	// 			Path: "duplicate-deleted-edge-deleted",
			// 	// 			Type: "delated",
			// 	// 			Time: parseTime("2020-01-03T12:34:56Z"),
			// 	// 		},
			// 	// 	},
			// 	// },
			// 	// {
			// 	// 	History: []*FileEvent{
			// 	// 		&FileEvent{
			// 	// 			Path:     "duplicate-deleted-edge-exists",
			// 	// 			Size:     10,
			// 	// 			Type:     "changed",
			// 	// 			Time:     parseTime("2020-01-01T12:34:56Z"),
			// 	// 			Checksum: "duplicate-deleted-edge",
			// 	// 		},
			// 	// 	},
			// 	// },
		},
	}

	expected := []*result{
		{Result: DiffEqual, Local: []string{"equal"}, Remote: []string{"equal"}},
		{Result: DiffEqual, Local: []string{"equal2"}, Remote: []string{"equal2"}},
		{Result: DiffEqual, Local: []string{"equal3"}, Remote: []string{"equal3"}},
	}

	var actual testAction
	err := local.Diff3(remote, &actual)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	margin, _ := time.ParseDuration("2s")
	opt1 := cmpopts.EquateApproxTime(margin)
	opt2 := cmpopts.IgnoreUnexported(FileInfo{})
	// opt3 := cmpopts.IgnoreFields(FileEvent{}, "Time")

	if diff := cmp.Diff(expected, actual.Result, opt1, opt2); diff != "" {
		t.Errorf("Diff2:\n%s", diff)
	}
}
