package lib

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestDiff(t *testing.T) {
	var local Boffin = &db{
		files: []*FileInfo{
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "equal",
						Size:     10,
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
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "equal2-hash-1",
					},
					&FileEvent{
						Path:     "equal2",
						Size:     10,
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
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "equal3-hash-1",
					},
					&FileEvent{
						Path:     "equal3",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "equal3-hash-2",
					},
					&FileEvent{
						Path:     "equal3",
						Size:     10,
						Time:     parseTime("2020-01-03T12:34:56Z"),
						Checksum: "equal3-hash-3",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "renamed-local",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "renamed-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "added-local",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "added-local-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "added-local2",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "added-local2-hash-1",
					},
					&FileEvent{
						Path:     "added-local2",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "added-local2-hash-2",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "hanging-delete-local",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "hanging-delete-local-hash-1",
					},
					&FileEvent{
						Path: "hanging-delete-local",
						Time: parseTime("2020-01-02T12:34:56Z"),
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "local-changed-l-1-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "local-changed-hash-1-1",
					},
					&FileEvent{
						Path:     "local-changed-l-1-2",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "local-changed-hash-1-2",
					},
					&FileEvent{
						Path:     "local-changed-l-1-3",
						Size:     10,
						Time:     parseTime("2020-01-03T12:34:56Z"),
						Checksum: "local-changed-hash-1-3",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "local-changed-l-2-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "local-changed-hash-2-1",
					},
					&FileEvent{
						Path:     "local-changed-l-2-2",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "local-changed-hash-2-2",
					},
					&FileEvent{
						Path:     "local-changed-l-2-3",
						Size:     10,
						Time:     parseTime("2020-01-03T12:34:56Z"),
						Checksum: "local-changed-hash-2-3",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "remote-changed-l-1-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "remote-changed-hash-1-1",
					},
					&FileEvent{
						Path:     "remote-changed-l-1-2",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "remote-changed-hash-1-2",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "remote-changed-l-2-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "remote-changed-hash-2-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "local-changed-conflict-l-1-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "local-changed-conflict-hash-1",
					},
					&FileEvent{
						Path:     "local-changed-conflict-l-1-1",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "local-changed-conflict-hash-2",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "remote-changed-conflict-l-1-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "remote-changed-conflict-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "remote-changed-conflict-l-1-2",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "remote-changed-conflict-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "both-changed-conflict-1-l",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-1",
					},
					&FileEvent{
						Path:     "both-changed-conflict-1-l",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-1-l",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "both-changed-conflict-2-1-l",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-2",
					},
					&FileEvent{
						Path:     "both-changed-conflict-2-1-l",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-2-1-l",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "both-changed-conflict-2-2-l",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-2",
					},
					&FileEvent{
						Path:     "both-changed-conflict-2-2-l",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-2-2-l",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "same-name-conflict",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "same-name-conflict-hash-l",
					},
				},
			},
		},
	}
	var remote Boffin = &db{
		files: []*FileInfo{
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "equal",
						Size:     10,
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
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "equal2-hash-1",
					},
					&FileEvent{
						Path:     "equal2",
						Size:     10,
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
						Time:     parseTime("2020-01-03T12:34:56Z"),
						Checksum: "equal3-hash-3",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "renamed-remote",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "renamed-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "added-remote",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "added-remote-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "added-remote2",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "added-remote2-hash-1",
					},
					&FileEvent{
						Path:     "added-remote2",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "added-remote2-hash-2",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "hanging-delete-remote",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "hanging-delete-remote-hash-1",
					},
					&FileEvent{
						Path: "hanging-delete-remote",
						Time: parseTime("2020-01-02T12:34:56Z"),
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "local-changed-r-1-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "local-changed-hash-1-1",
					},
					&FileEvent{
						Path:     "local-changed-r-1-2",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "local-changed-hash-1-2",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "local-changed-r-2-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "local-changed-hash-2-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "remote-changed-r-1-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "remote-changed-hash-1-1",
					},
					&FileEvent{
						Path:     "remote-changed-r-1-2",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "remote-changed-hash-1-2",
					},
					&FileEvent{
						Path:     "remote-changed-r-1-3",
						Size:     10,
						Time:     parseTime("2020-01-03T12:34:56Z"),
						Checksum: "remote-changed-hash-1-3",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "remote-changed-r-2-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "remote-changed-hash-2-1",
					},
					&FileEvent{
						Path:     "remote-changed-r-2-2",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "remote-changed-hash-2-2",
					},
					&FileEvent{
						Path:     "remote-changed-r-2-3",
						Size:     10,
						Time:     parseTime("2020-01-03T12:34:56Z"),
						Checksum: "remote-changed-hash-2-3",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "local-changed-conflict-r-1-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "local-changed-conflict-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "local-changed-conflict-r-1-2",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "local-changed-conflict-hash-1",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "remote-changed-conflict-r-1-1",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "remote-changed-conflict-hash-1",
					},
					&FileEvent{
						Path:     "remote-changed-conflict-r-1-1",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "remote-changed-conflict-hash-2",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "both-changed-conflict-1-r",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-1",
					},
					&FileEvent{
						Path:     "both-changed-conflict-1-r",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-1-r",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "both-changed-conflict-2-1-r",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-2",
					},
					&FileEvent{
						Path:     "both-changed-conflict-2-1-r",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-2-1-r",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "both-changed-conflict-2-2-r",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-2",
					},
					&FileEvent{
						Path:     "both-changed-conflict-2-2-r",
						Size:     10,
						Time:     parseTime("2020-01-02T12:34:56Z"),
						Checksum: "both-changed-conflict-hash-2-2-r",
					},
				},
			},
			{
				History: []*FileEvent{
					&FileEvent{
						Path:     "same-name-conflict",
						Size:     10,
						Time:     parseTime("2020-01-01T12:34:56Z"),
						Checksum: "same-name-conflict-hash-r",
					},
				},
			},
		},
	}

	expected := []*result{
		{Result: "conflict", Local: []string{"both-changed-conflict-1-l"}, Remote: []string{"both-changed-conflict-1-r"}},
		{Result: "conflict", Local: []string{"both-changed-conflict-2-1-l", "both-changed-conflict-2-2-l"}, Remote: []string{"both-changed-conflict-2-1-r", "both-changed-conflict-2-2-r"}},
		{Result: "conflict", Local: []string{"local-changed-conflict-l-1-1"}, Remote: []string{"local-changed-conflict-r-1-1", "local-changed-conflict-r-1-2"}},
		{Result: "conflict", Local: []string{"remote-changed-conflict-l-1-1", "remote-changed-conflict-l-1-2"}, Remote: []string{"remote-changed-conflict-r-1-1"}},
		{Result: "conflict", Local: []string{"same-name-conflict"}, Remote: []string{"same-name-conflict"}},
		{Result: "local-changed", Local: []string{"local-changed-l-1-3"}, Remote: []string{"local-changed-r-1-2"}},
		{Result: "local-changed", Local: []string{"local-changed-l-2-3"}, Remote: []string{"local-changed-r-2-1"}},
		{Result: "local-old", Local: []string{"hanging-delete-local"}},
		{Result: "local-only", Local: []string{"added-local"}},
		{Result: "local-only", Local: []string{"added-local2"}},
		{Result: "moved", Local: []string{"renamed-local"}, Remote: []string{"renamed-remote"}},
		{Result: "remote-changed", Local: []string{"remote-changed-l-1-2"}, Remote: []string{"remote-changed-r-1-3"}},
		{Result: "remote-changed", Local: []string{"remote-changed-l-2-1"}, Remote: []string{"remote-changed-r-2-3"}},
		{Result: "remote-old", Remote: []string{"hanging-delete-remote"}},
		{Result: "remote-only", Remote: []string{"added-remote"}},
		{Result: "remote-only", Remote: []string{"added-remote2"}},
		{Result: "unchanged", Local: []string{"equal"}, Remote: []string{"equal"}},
		{Result: "unchanged", Local: []string{"equal2"}, Remote: []string{"equal2"}},
		{Result: "unchanged", Local: []string{"equal3"}, Remote: []string{"equal3"}},
	}

	var actual testAction
	err := Diff(local, remote, &actual)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	actual.Sort()

	margin, _ := time.ParseDuration("2s")
	opt1 := cmpopts.EquateApproxTime(margin)
	opt2 := cmpopts.IgnoreUnexported(FileInfo{})
	// opt3 := cmpopts.IgnoreFields(FileEvent{}, "Time")

	if diff := cmp.Diff(expected, actual.Result, opt1, opt2); diff != "" {
		t.Errorf("Diff:\n%s", diff)
	}
}
