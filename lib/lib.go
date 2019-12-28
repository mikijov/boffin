/*
Copyright (C) 2019 Milutin Jovanvović

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package lib

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const defaultDbDir = ".aether"
const filesFilename = "files.json"
const newFilesFilename = "files.json.tmp"
const deleted = "[deleted]"

// fileMeta ...
type fileMeta struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Time     time.Time `json:"time"`
	Checksum []string  `json:"checksum"`

	checked bool
	changed bool
}

// comment
const (
	Equal = iota
	LeftChanged
	RightChanged
	Conflict
)

func (fm *fileMeta) isChecked() bool {
	return fm.checked
}

func (fm *fileMeta) isChanged() bool {
	return fm.changed
}

func (fm *fileMeta) markDeleted() {
	fm.Checksum = append(fm.Checksum, deleted)
}

func (fm *fileMeta) update(path string, info os.FileInfo) error {
	fm.checked = true

	if info.Size() == fm.Size && info.ModTime().UTC() == fm.Time {
		// size and time matches, assume no change
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	hashStr := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	// fmt.Printf("Hash: %s\n", hashStr)

	fm.Size = info.Size()
	fm.Time = info.ModTime().UTC()

	if fm.Checksum == nil {
		fm.Checksum = []string{hashStr}
		fm.changed = true
	} else if len(fm.Checksum) == 0 || fm.Checksum[len(fm.Checksum)-1] != hashStr {
		fm.Checksum = append(fm.Checksum, hashStr)
		fm.changed = true
	}

	return nil
}

// FileDB ...
type FileDB interface {
	GetDbDir() string
	GetBaseDir() string
	Save() error
	Update() error
	IsChanged() bool
}

type fileDb struct {
	dbDir      string
	absBaseDir string
	changed    bool

	BaseDir string      `json:"baseDir"`
	Files   []*fileMeta `json:"files"`
}

// GetDbDir ...
func (db *fileDb) GetDbDir() string {
	return db.dbDir
}

// GetBaseDir ...
func (db *fileDb) GetBaseDir() string {
	return db.absBaseDir
}

// Save ...
func (db *fileDb) Save() error {
	if !db.changed {
		return nil
	}

	newFilename := filepath.Join(db.dbDir, newFilesFilename)
	defer os.Remove(newFilename) // cleanup; will work only if the old file could not be replaced

	{ // write new file
		file, err := os.OpenFile(newFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0444)
		if err != nil {
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if encoder == nil {
			return fmt.Errorf("failed to create json encoder")
		}
		encoder.SetIndent("", "  ")

		encoder.Encode(db)

		file.Close()
	}

	{ // now replace old file with the new one
		filename := filepath.Join(db.dbDir, filesFilename)

		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to overwrite '%s'", filename)
		}
		if err := os.Rename(newFilename, filename); err != nil {
			return fmt.Errorf("critical error; failed to rename '%s' to '%s'", newFilename, filename)
		}
	}

	db.changed = false

	return nil
}

// Update ...
func (db *fileDb) Update() error {
	files := make(map[string]*fileMeta)
	for _, file := range db.Files {
		files[file.Name] = file
	}

	dir := db.GetBaseDir()
	// fmt.Printf("dir: %s\n", dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// fmt.Printf("path: %s\n", path)
		if err != nil {
			fmt.Printf("%s: error getting file info: %s\n", path, err)
			return nil
		}
		if info.IsDir() {
			if info.Name() == defaultDbDir {
				return filepath.SkipDir
			}
			return nil
		}

		root := path[:len(dir)]
		if dir != root {
			// TODO: just checking if the beginning of the path matches expectation
			fmt.Printf("Root mismatch '%s' != '%s'\n", dir, root)
		}

		relPath := path[len(dir)+1:]

		meta, ok := files[relPath]
		if !ok {
			meta = &fileMeta{
				Name: relPath,
			}
			files[relPath] = meta
		}

		err = meta.update(path, info)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	fileSlice := make([]*fileMeta, 0, len(files))
	changed := false
	for _, file := range files {
		if !file.isChecked() {
			file.markDeleted()
		} else {
			changed = changed || file.isChanged()
		}
		fileSlice = append(fileSlice, file)
	}

	if changed {
		sort.Slice(fileSlice, func(i, j int) bool {
			return fileSlice[i].Name < fileSlice[j].Name
		})
		db.Files = fileSlice

		db.changed = changed
	}

	return nil
}

// IsChanged ...
func (db *fileDb) IsChanged() bool {
	return db.changed
}

func cleanPath(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.Clean(dir), nil
}

// ConstuctDbDirPath ...
func ConstuctDbDirPath(dir string) string {
	return filepath.Join(dir, defaultDbDir)
}

// InitDbDir ...
func InitDbDir(dbDir, baseDir string) (FileDB, error) {
	baseDir, err := cleanPath(baseDir)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("'%s' does not exist", baseDir)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("'%s' is not a directory", baseDir)
	}

	dbDir, err = cleanPath(dbDir)
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(dbDir)
	if err == nil {
		return nil, fmt.Errorf("'%s' already exists", dbDir)
	}
	err = os.Mkdir(dbDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	db := &fileDb{
		dbDir:      dbDir,
		absBaseDir: baseDir,
		changed:    true,
	}

	if relDir, err := filepath.Rel(dbDir, baseDir); err == nil {
		// if we can deduce relative path, use it instead of absolute one
		db.BaseDir = relDir
	} else {
		db.BaseDir = baseDir
	}

	if err = db.Save(); err != nil {
		return nil, err
	}

	return db, nil
}

// FindDbDir ...
func FindDbDir(dir string) (string, error) {
	dir, err := cleanPath(dir)
	if err != nil {
		return "", err
	}

	for true {
		dbDir := filepath.Join(dir, defaultDbDir)
		info, err := os.Stat(dbDir)
		if err == nil && info.IsDir() {
			return dbDir, nil
		}
		if dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}

	return "", fmt.Errorf("could not find %s dir", defaultDbDir)
}

// LoadFileDB ...
func LoadFileDB(dbDir string) (FileDB, error) {
	dbDir, err := cleanPath(dbDir)
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(dbDir, filesFilename)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	db := &fileDb{
		dbDir: dbDir,
	}

	if err := decoder.Decode(&db); err != nil {
		return nil, err
	}

	// ensure there is nothing after the first json object
	dummy := &fileDb{}
	if err = decoder.Decode(&dummy); err != io.EOF {
		return nil, fmt.Errorf("unexpected contents at the end of config file")
	}

	if filepath.IsAbs(db.BaseDir) {
		db.absBaseDir = db.BaseDir
	} else {
		db.absBaseDir, err = cleanPath(filepath.Join(dbDir, db.BaseDir))
		if err != nil {
			return nil, err
		}
	}

	info, err := os.Stat(db.absBaseDir)
	if err != nil {
		return nil, fmt.Errorf("base directory '%s' does not exist", db.absBaseDir)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("base directory '%s' is not a directory", db.absBaseDir)
	}

	return db, nil
}
