/*
Copyright (C) 2020 Milutin Jovanvović

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
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// .8888b oo dP          oo          .8888b
// 88   "    88                      88   "
// 88aaa  dP 88 .d8888b. dP 88d888b. 88aaa  .d8888b.
// 88     88 88 88ooood8 88 88'  `88 88     88'  `88
// 88     88 88 88.  ... 88 88    88 88     88.  .88
// dP     dP dP `88888P' dP dP    dP dP     `88888P'

// FileEvent ...
type FileEvent struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size,omitempty"`
	Time     time.Time `json:"time"`
	Checksum string    `json:"checksum,omitempty"`
}

// FileInfo ...
type FileInfo struct {
	History []*FileEvent `json:"history,omitempty"`
}

// Checksum ...
func (fi *FileInfo) Checksum() string {
	return fi.History[len(fi.History)-1].Checksum
}

// Path ...
func (fi *FileInfo) Path() string {
	for i := range fi.History {
		event := fi.History[len(fi.History)-1-i]
		if event.Checksum != "" {
			return event.Path
		}
	}
	return ""
}

// Size ...
func (fi *FileInfo) Size() int64 {
	for i := range fi.History {
		event := fi.History[len(fi.History)-1-i]
		if event.Checksum != "" {
			return event.Size
		}
	}
	return 0
}

// Time ...
func (fi *FileInfo) Time() time.Time {
	for i := range fi.History {
		event := fi.History[len(fi.History)-1-i]
		if event.Checksum != "" {
			return event.Time
		}
	}
	return time.Time{}
}

// IsDeleted ...
func (fi *FileInfo) IsDeleted() bool {
	if len(fi.History) == 0 {
		return true
	}
	return fi.History[len(fi.History)-1].Checksum == ""
}

func (fi *FileInfo) MarkDeleted() {
	if !fi.IsDeleted() {
		fi.History = append(fi.History, &FileEvent{
			Path: fi.Path(),
			Time: time.Now().UTC(),
		})
	}
}

//       dP dP
//       88 88
// .d888b88 88d888b.
// 88'  `88 88'  `88
// 88.  .88 88.  .88
// `88888P8 88Y8888'

// Boffin ...
type Boffin interface {
	GetFiles() []*FileInfo
	AddFile(file *FileInfo)

	GetDbDir() string
	GetBaseDir() string
	GetImportDir() string
	GetRelImportDir() string

	Save() error
}

type db struct {
	dbDir        string
	absBaseDir   string
	absImportDir string

	// this is simply kept for saving purposes
	baseDir   string
	importDir string
	files     []*FileInfo
}

// GetDbDir ...
func (db *db) GetDbDir() string {
	return db.dbDir
}

// GetBaseDir ...
func (db *db) GetBaseDir() string {
	return db.absBaseDir
}

// GetImportDir ...
func (db *db) GetImportDir() string {
	return db.absImportDir
}

// GetRelImportDir ...
func (db *db) GetRelImportDir() string {
	return db.importDir
}

// GetFiles ...
func (db *db) GetFiles() []*FileInfo {
	return append([]*FileInfo{}, db.files...)
}

// AddFile ...
func (db *db) AddFile(file *FileInfo) {
	db.files = append(db.files, file)
}

func cleanPath(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.Clean(dir), nil
}

// dP                                dP      d8' .d88888b
// 88                                88     d8'  88.    "'
// 88        .d8888b. .d8888b. .d888b88    d8'   `Y88888b. .d8888b. dP   .dP .d8888b.
// 88        88'  `88 88'  `88 88'  `88   d8'          `8b 88'  `88 88   d8' 88ooood8
// 88        88.  .88 88.  .88 88.  .88  d8'     d8'   .8P 88.  .88 88 .88'  88.  ...
// 88888888P `88888P' `88888P8 `88888P8 88        Y88888P  `88888P8 8888P'   `88888P'

const defaultDbDir = ".boffin"
const filesFilename = "files.json"
const newFilesFilename = "files.json.tmp"

type jsonStruct struct {
	V1 *v1Struct `json:"v1,omitempty"`
}

type v1Struct struct {
	BaseDir   string      `json:"base-dir"`
	ImportDir string      `json:"import-dir"`
	Files     []*FileInfo `json:"files"`
}

// InitDbDir ...
func InitDbDir(dbDir, baseDir string) (Boffin, error) {
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

	db := &db{
		dbDir:      dbDir,
		absBaseDir: baseDir,
	}

	if relDir, err := filepath.Rel(dbDir, baseDir); err == nil {
		// if we can deduce relative path, use it instead of absolute one
		db.baseDir = relDir
	} else {
		db.baseDir = baseDir
	}

	if err = db.Save(); err != nil {
		return nil, err
	}

	return db, nil
}

// Save ...
func (db *db) Save() error {
	sort.Slice(db.files, func(i, j int) bool {
		return db.files[i].Path() < db.files[j].Path()
	})

	rawJSON := &jsonStruct{
		V1: &v1Struct{
			BaseDir: db.baseDir,
			Files:   db.files,
		},
	}

	newFilename := filepath.Join(db.dbDir, newFilesFilename)
	defer os.Remove(newFilename) // cleanup; will work only if the old file could not be replaced

	{ // write new file
		file, err := os.OpenFile(newFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		if encoder == nil {
			return fmt.Errorf("failed to create json encoder")
		}
		encoder.SetIndent("", "  ")

		encoder.Encode(rawJSON)
	}

	{ // now replace old file with the new one
		filename := filepath.Join(db.dbDir, filesFilename)

		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to overwrite '%s'", filename)
		}
		if err := os.Rename(newFilename, filename); err != nil {
			return fmt.Errorf("critical error; failed to rename '%s' to '%s'", newFilename, filename)
		}

		if fi, err := os.Stat(filename); err == nil {
			os.Chmod(filename, fi.Mode()&0444)
		} else {
			log.Printf("warning: failed to make repo file read only")
		}
	}

	return nil
}

// LoadBoffin ...
func LoadBoffin(dbDir string) (Boffin, error) {
	boffinPath := filepath.Join(dbDir, filesFilename)

	boffinFile, err := os.Open(boffinPath)
	if err != nil {
		return nil, err
	}
	defer boffinFile.Close()

	decoder := json.NewDecoder(boffinFile)
	decoder.DisallowUnknownFields()

	rawJSON := &jsonStruct{}
	if err := decoder.Decode(&rawJSON); err != nil {
		return nil, err
	}

	// ensure there is nothing after the first json object
	dummy := &jsonStruct{}
	if err = decoder.Decode(&dummy); err != io.EOF {
		return nil, fmt.Errorf("unexpected contents at the end of config file")
	}

	if rawJSON.V1 == nil {
		return nil, fmt.Errorf("config file is empty")
	}

	db := &db{
		dbDir:     dbDir,
		baseDir:   rawJSON.V1.BaseDir,
		importDir: rawJSON.V1.ImportDir,
		files:     rawJSON.V1.Files,
	}

	if filepath.IsAbs(db.baseDir) {
		db.absBaseDir, err = cleanPath(db.baseDir)
	} else {
		db.absBaseDir, err = cleanPath(filepath.Join(dbDir, db.baseDir))
	}
	if filepath.IsAbs(db.importDir) {
		db.absImportDir, err = cleanPath(db.importDir)
	} else {
		db.absImportDir, err = cleanPath(filepath.Join(db.absBaseDir, db.importDir))
	}
	if err != nil {
		return nil, err
	}

	return db, nil
}

// ConstuctDbPath ...
func ConstuctDbPath(baseDir string) string {
	return filepath.Join(baseDir, defaultDbDir)
}

// FindBoffinDir ...
func FindBoffinDir(dir string) (string, error) {
	// if dir is empty, start in current directory
	if dir == "" {
		var err error
		if dir, err = os.Getwd(); err != nil {
			return "", err
		}
	}

	dir, err := cleanPath(dir)
	if err != nil {
		return "", err
	}

	// look into current or any parent directory for a root which has defaultDbDir
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

func CalculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}
