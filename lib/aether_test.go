package lib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindAether(t *testing.T) {
	testRoot, _ := os.Getwd()
	testRoot = filepath.Join(testRoot, "../test")
	testRoot = filepath.Clean(testRoot)

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
