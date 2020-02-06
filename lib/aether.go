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
	// "sort"
	"time"
)

const defaultDbDir = ".aether"
const filesFilename = "files.json"
const newFilesFilename = "files.json.tmp"

const changed = "changed"
const deleted = "deleted"

// FileEvent ...
type FileEvent struct {
	Type     string    `json:"event"`
	Time     time.Time `json:"time"`
	Checksum string    `json:"checksum,omitempty"`
}

// FileInfo ...
type FileInfo struct {
	Path     string       `json:"path"`
	Size     int64        `json:"size"`
	Time     time.Time    `json:"time"`
	Checksum string       `json:"checksum,omitempty"`
	History  []*FileEvent `json:"history,omitempty"`

	checked bool
}

// Aether ...
type Aether interface {
	GetFiles() []*FileInfo

	GetDbDir() string
	GetBaseDir() string
	GetImportDir() string

	Save() error
	Update() error
}

// func stringInSlice(a string, list []string) bool {
// 	for _, b := range list {
// 		if b == a {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (fm *FileInfo) compare(other *FileInfo) cmpResult {
// 	if other == nil {
// 		return cmpNewer
// 	}
//
// 	lhash := fm.GetHash()
// 	rhash := fm.GetHash()
//
// 	if lhash == rhash {
// 		return cmpEqual
// 	}
//
// 	if stringInSlice(rhash, fm.Checksum) && !stringInSlice(lhash, other.Checksum) {
// 		return cmpNewer
// 	}
// 	if stringInSlice(lhash, other.Checksum) && !stringInSlice(rhash, fm.Checksum) {
// 		return cmpOlder
// 	}
//
// 	return cmpConflict
// }

func (fi *FileInfo) isDeleted() bool {
	return fi.Checksum == ""
}

func (fi *FileInfo) markDeleted() {
	if !fi.isDeleted() {
		fi.Checksum = ""
		fi.Size = 0
		fi.Time = time.Now().UTC()

		fi.History = append(fi.History, &FileEvent{
			Type: deleted,
			Time: fi.Time,
		})
	}
}

func (fi *FileInfo) update(path string, info os.FileInfo) error {
	fi.checked = true
	// fmt.Printf("checking %s\n", fm.Name)

	if !fi.isDeleted() { // check size/time only if not marked as deleted
		if info.Size() == fi.Size && info.ModTime().UTC() == fi.Time {
			// size and time matches, assume no change
			return nil
		}
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

	fi.Size = info.Size()
	fi.Time = info.ModTime().UTC()
	fi.Checksum = hashStr

	fi.History = append(fi.History, &FileEvent{
		Type:     changed,
		Time:     fi.Time,
		Checksum: hashStr,
	})

	return nil
}

//       dP dP
//       88 88
// .d888b88 88d888b.
// 88'  `88 88'  `88
// 88.  .88 88.  .88
// `88888P8 88Y8888'

type db struct {
	dbDir      string
	absBaseDir string

	// this is simply kept for saving purposes
	baseDir string
	files   []*FileInfo
}

// GetDbDir ...
func (db *db) GetDbDir() string {
	return db.dbDir
}

// GetBaseDir ...
func (db *db) GetBaseDir() string {
	return db.absBaseDir
}

func (db *db) GetImportDir() string {
	return filepath.Join(db.GetBaseDir(), "import")
}

func (db *db) GetFiles() []*FileInfo {
	return db.files
}

func sliceToMap(files []*FileInfo) map[string]*FileInfo {
	fileMap := make(map[string]*FileInfo)

	for _, file := range files {
		fileMap[file.Path] = file
	}

	return fileMap
}

// Update ...
func (db *db) Update() error {
	dir := db.GetBaseDir()

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("base directory '%s' does not exist", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("base directory '%s' is not a directory", dir)
	}

	// TODO: consider if this should deep copy if object should stay unchanged on error
	fileMap := sliceToMap(db.files)

	// fmt.Printf("walking %s\n", dir)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("%s: error getting file info: %s\n", path, err)
			return nil
		}
		if info.IsDir() {
			if info.Name() == defaultDbDir { // skip DB directory
				return filepath.SkipDir
			}
			return nil
		}

		root := path[:len(dir)]
		if dir != root {
			// this should never happen
			return fmt.Errorf("unexpected error; root mismatch '%s' != '%s'", dir, root)
		}

		relPath := path[len(dir)+1:]

		file, ok := fileMap[relPath]
		if !ok {
			file = &FileInfo{
				Path: relPath,
			}
			fileMap[relPath] = file
			db.files = append(db.files, file)
		}

		err = file.update(path, info)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, file := range db.files {
		if !file.checked {
			file.markDeleted()
		}
	}

	return nil
}

// // Import ...
// func (db *fileDb) Import(remote FileDb) error {
//
// 	// - for each file in local DB:
// 	for _, file := range db.files {
// 		file.markChecked()
//
// 		rfile := remote.GetFile(file.Name)
// 		if rfile != nil {
// 			rfile.markChecked()
// 		}
//
// 		result := file.compare(rfile)
//
// 		if result == cmpNewer {
// 			fmt.Printf("newer: %s\n", file.Name)
// 		} else if result == cmpOlder {
// 			fmt.Printf("older: %s\n", file.Name)
// 		} else if result == cmpEqual {
// 			fmt.Printf("equal: %s\n", file.Name)
// 		} else if result == cmpEqual {
// 			fmt.Printf("conflict: %s\n", file.Name)
// 		} else {
// 			fmt.Printf("???: %s\n", file.Name)
// 		}
// 	}
//
// 	for _, file := range db.files {
// 		if !file.isChecked() {
// 			fmt.Printf("not checked!: %s\n", file.Name)
// 		}
// 	}
//
// 	for _, file := range remote.(*fileDb).files {
// 		if !file.isChecked() {
// 			fmt.Printf("new file: %s\n", file.Name)
// 		}
// 	}
//
// 	return nil
// }

// // Import2 ...
// func (db *fileDb) Import2(remote FileDb) error {
// 	remote2, ok := remote.(*fileDb)
// 	if !ok {
// 		return fmt.Errorf("remote parameter has to be fileDb")
// 	}
//
// 	hashes := make(map[string]bool)
// 	for _, file := range db.files {
// 		for _, hash := range file.Checksum {
// 			if hash != deleted {
// 				hashes[hash] = true
// 			}
// 		}
// 	}
//
// 	for _, file := range remote2.files {
// 		if !file.isDeleted() {
// 			if _, seen := hashes[file.GetHash()]; !seen {
// 				src := filepath.Join(remote.GetBaseDir(), file.Name)
// 				dest := filepath.Join(db.GetImportDir(), file.Name)
// 				if err := copyFile(src, dest); err != nil {
// 					return err
// 				}
//
// 				hashes[file.GetHash()] = true
//
// 				// TODO: is a reference OK or should we make a copy
// 				db.files[file.Name] = file
// 			}
// 		}
// 	}
//
// 	return nil
// }

func cleanPath(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.Clean(dir), nil
}

// // Copy the src file to dst. Any existing file will be overwritten and will not
// // copy file attributes.
// func copyFile(src, dst string) error {
// 	fmt.Printf("%s => %s\n", src, dst)
//
// 	stat, err := os.Stat(src)
// 	if err != nil {
// 		return err
// 	}
//
// 	in, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer in.Close()
//
// 	out, err := os.OpenFile(dst, os.O_CREAT|os.O_EXCL|os.O_WRONLY, 0600)
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()
//
// 	_, err = io.Copy(out, in)
// 	if err != nil {
// 		return err
// 	}
// 	err = out.Chmod(stat.Mode())
// 	if err != nil {
// 		return err
// 	}
// 	err = out.Close()
// 	if err != nil {
// 		return err
// 	}
// 	err = os.Chtimes(dst, stat.ModTime(), stat.ModTime())
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// dP                                dP      d8' .d88888b
// 88                                88     d8'  88.    "'
// 88        .d8888b. .d8888b. .d888b88    d8'   `Y88888b. .d8888b. dP   .dP .d8888b.
// 88        88'  `88 88'  `88 88'  `88   d8'          `8b 88'  `88 88   d8' 88ooood8
// 88        88.  .88 88.  .88 88.  .88  d8'     d8'   .8P 88.  .88 88 .88'  88.  ...
// 88888888P `88888P' `88888P8 `88888P8 88        Y88888P  `88888P8 8888P'   `88888P'

type jsonStruct struct {
	V1 *v1Struct `json:"v1,omitempty"`
}

type v1Struct struct {
	BaseDir string      `json:"base-dir"`
	Files   []*FileInfo `json:"files"`
}

// InitDbDir ...
func InitDbDir(dbDir, baseDir string) (Aether, error) {
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
	// // prepare simplified json structure
	// fileSlice := make([]*FileInfo, 0, len(db.files))
	// for _, file := range db.files {
	// 	fileSlice = append(fileSlice, file)
	// }
	// sort.Slice(fileSlice, func(i, j int) bool {
	// 	return fileSlice[i].Name < fileSlice[j].Name
	// })

	rawJSON := &jsonStruct{
		V1: &v1Struct{
			BaseDir: db.baseDir,
			Files:   db.files,
		},
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
	}

	return nil
}

// LoadAether ...
func LoadAether(dbDir string) (Aether, error) {
	aetherPath := filepath.Join(dbDir, filesFilename)

	aetherFile, err := os.Open(aetherPath)
	if err != nil {
		return nil, err
	}
	defer aetherFile.Close()

	decoder := json.NewDecoder(aetherFile)

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
		dbDir:   dbDir,
		baseDir: rawJSON.V1.BaseDir,
		files:   rawJSON.V1.Files,
	}

	if filepath.IsAbs(db.baseDir) {
		db.absBaseDir, err = cleanPath(db.baseDir)
	} else {
		db.absBaseDir, err = cleanPath(filepath.Join(dbDir, db.baseDir))
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

// FindAetherDir ...
func FindAetherDir(dir string) (string, error) {
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
