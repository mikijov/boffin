package lib

import (
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestUpdate(t *testing.T) {
	dir := filepath.Join(getTestDir(), "update2", ".boffin")

	boffin, err := LoadBoffin(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = Update(boffin, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []*FileInfo{
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "equal-with-history.ext",
					Size:     23,
					Time:     parseTime("2020-05-21T04:06:07.654626019Z"),
					Checksum: "XEvlLE++MWj0ybhJqLLM44Ai5VzyrjxymWAHm0GMN/I=",
				},
				&FileEvent{
					Path:     "equal-with-history-1.ext",
					Size:     23,
					Time:     parseTime("2020-05-21T04:06:07.654626019Z"),
					Checksum: "XEvlLE++MWj0ybhJqLLM44Ai5VzyrjxymWAHm0GMN/I=",
				},
				&FileEvent{
					Path:     "equal-with-history.ext",
					Size:     23,
					Time:     parseTime("2020-05-21T04:06:07.654626019Z"),
					Checksum: "XEvlLE++MWj0ybhJqLLM44Ai5VzyrjxymWAHm0GMN/I=",
				},
			},
		},
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "equal.ext",
					Size:     10,
					Time:     parseTime("2020-02-06T13:56:51.571756332Z"),
					Checksum: "mv0rsY4Lof04c4eVesQRoggxIMQBLzv82jX0gglIhrI=",
				},
			},
		},
		// {
		// 	History: []*FileEvent{
		// 		&FileEvent{
		// 			Path:     "multi-hash/a-equal.ext",
		// 			Size:     30,
		// 			Time:     parseTime("2020-02-25T14:03:45.270218079Z"),
		// 			Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
		// 		},
		// 	},
		// },
		// {
		// 	History: []*FileEvent{
		// 		&FileEvent{
		// 			Path:     "multi-hash/changed.ext",
		// 			Size:     87,
		// 			Time:     parseTime("2020-02-24T14:01:07.828877429Z"),
		// 			Checksum: "OfhhlgGfv8NbaPqLE6EGi2c9EJR8edzvXFnxl3QUvYo=",
		// 		},
		// 		&FileEvent{
		// 			Path:     "multi-hash/changed.ext",
		// 			Size:     87,
		// 			Time:     parseTime("2020-02-25T14:01:07.828877429Z"),
		// 			Checksum: "ofhhlgGfv8NbaPqLE6EGi2c9EJR8edzvXFnxl3QUvYo=",
		// 		},
		// 	},
		// },
		// {
		// 	History: []*FileEvent{
		// 		&FileEvent{
		// 			Path:     "multi-hash/m-equal.ext",
		// 			Size:     30,
		// 			Time:     parseTime("2020-02-25T14:03:57.917265752Z"),
		// 			Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
		// 		},
		// 	},
		// },
		// {
		// 	History: []*FileEvent{
		// 		&FileEvent{
		// 			Path:     "multi-hash/moved.ext",
		// 			Size:     30,
		// 			Time:     parseTime("2020-02-25T14:02:41.831652522Z"),
		// 			Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
		// 		},
		// 		&FileEvent{
		// 			Path:     "multi-hash/moved/moved.ext",
		// 			Size:     30,
		// 			Time:     parseTime("2020-02-25T14:02:41.831652522Z"),
		// 			Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
		// 		},
		// 	},
		// },
		// {
		// 	History: []*FileEvent{
		// 		&FileEvent{
		// 			Path:     "multi-hash/renamed-before.ext",
		// 			Size:     30,
		// 			Time:     parseTime("2020-02-25T14:03:45.270218079Z"),
		// 			Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
		// 		},
		// 		&FileEvent{
		// 			Path:     "multi-hash/renamed.ext",
		// 			Size:     30,
		// 			Time:     parseTime("2020-02-25T14:03:45.270218079Z"),
		// 			Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
		// 		},
		// 	},
		// },
		// {
		// 	History: []*FileEvent{
		// 		&FileEvent{
		// 			Path:     "multi-hash/z-equal.ext",
		// 			Size:     30,
		// 			Time:     parseTime("2020-02-25T14:04:02.354066271Z"),
		// 			Checksum: "/C5STUNIN3N2elIyckcY1xYP8pw9Dh5VyVs+wt5pePQ=",
		// 		},
		// 	},
		// },
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "sub1/cross-rename-2.ext",
					Size:     19,
					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
					Checksum: "Iag4g9z39+jJOdVGxqCNXziaAFwFZ8dnfdMrZQz1qKM=",
				},
				&FileEvent{
					Path:     "sub1/cross-rename-1.ext",
					Size:     19,
					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
					Checksum: "Iag4g9z39+jJOdVGxqCNXziaAFwFZ8dnfdMrZQz1qKM=",
				},
			},
		},
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "sub1/cross-rename-1.ext",
					Size:     19,
					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
					Checksum: "K1L3GOGZF5wiOtiJdkN6+xZiAKwG77ueF+KnMyCXAuI=",
				},
				&FileEvent{
					Path:     "sub1/cross-rename-2.ext",
					Size:     19,
					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
					Checksum: "K1L3GOGZF5wiOtiJdkN6+xZiAKwG77ueF+KnMyCXAuI=",
				},
			},
		},
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "sub1/deleted.ext",
					Size:     10,
					Time:     parseTime("2020-02-06T13:59:21.099018324Z"),
					Checksum: "71JuQzM1k9ZV2tMnnhemjf+FUfbEEs8YS170IORPpA4=",
				},
				&FileEvent{
					Path: "sub1/deleted.ext",
					Time: time.Now(),
				},
			},
		},
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "sub1/equal.ext",
					Size:     10,
					Time:     parseTime("2020-02-06T13:57:02.90203166Z"),
					Checksum: "vQTuoHT8OnxI9g7fcZnEeTC9jcbX1NuRsS4gyDQkxjE=",
				},
			},
		},
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "sub1/new.ext",
					Size:     10,
					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
					Checksum: "Z12qAGMLMXMmfBWqZw8LHTJD2Ifpp8AMJYmCa4eMYac=",
				},
			},
		},
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "sub1/renamed-before.ext",
					Size:     10,
					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
					Checksum: "4PFd3bElTqFi8wvTlY2eRK6sJo65UivdK95nd7it5h4=",
				},
				&FileEvent{
					Path:     "sub1/renamed-after.ext",
					Size:     10,
					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
					Checksum: "4PFd3bElTqFi8wvTlY2eRK6sJo65UivdK95nd7it5h4=",
				},
			},
		},
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "sub1/move-rename-before.ext",
					Size:     16,
					Time:     parseTime("2020-02-05T13:57:12.378926011Z"),
					Checksum: "Ir6w9XOc7mlfgjFEhsjZAdhiqNosCRCf9iqzt3o7ndY=",
				},
				&FileEvent{
					Path:     "sub2/move-rename.ext",
					Size:     16,
					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
					Checksum: "Ir6w9XOc7mlfgjFEhsjZAdhiqNosCRCf9iqzt3o7ndY=",
				},
			},
		},
		{
			History: []*FileEvent{
				&FileEvent{
					Path:     "sub1/moved.ext",
					Size:     10,
					Time:     parseTime("2020-02-06T13:57:12.378926011Z"),
					Checksum: "xP4lKAtsUEfiZZ+Z4wlwZ3yFIxq8w7PPdIBvNBzZhd4=",
				},
				&FileEvent{
					Path:     "sub2/moved.ext",
					Size:     10,
					Time:     parseTime("2020-02-25T04:19:14.250535938Z"),
					Checksum: "xP4lKAtsUEfiZZ+Z4wlwZ3yFIxq8w7PPdIBvNBzZhd4=",
				},
			},
		},
	}
	actual := boffin.GetFiles()
	sort.Slice(actual, func(i, j int) bool {
		return actual[i].Path() < actual[j].Path()
	})

	margin, _ := time.ParseDuration("2s")
	opt1 := cmpopts.EquateApproxTime(margin)
	opt2 := cmpopts.IgnoreUnexported(FileInfo{})
	// opt3 := cmpopts.IgnoreFields(FileEvent{}, "Time")

	if diff := cmp.Diff(expected, actual, opt1, opt2); diff != "" {
		t.Errorf("file.History:\n%s", diff)
	}
}
