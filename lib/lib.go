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

// FileMeta ...
type FileMeta struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Time     time.Time `json:"time"`
	Checksum []string  `json:"checksum"`

	checked bool
	changed bool
}

// FileDb ...
type FileDb interface {
	GetDbDir() string
	GetBaseDir() string

	GetFile(path string) *FileMeta

	Save() error
	Update() error
	Import(remote FileDb) error
}

//  88888888b oo dP          8888ba.88ba             dP
//  88           88          88  `8b  `8b            88
// a88aaaa    dP 88 .d8888b. 88   88   88 .d8888b. d8888P .d8888b.
//  88        88 88 88ooood8 88   88   88 88ooood8   88   88'  `88
//  88        88 88 88.  ... 88   88   88 88.  ...   88   88.  .88
//  dP        dP dP `88888P' dP   dP   dP `88888P'   dP   `88888P8

func (fm *FileMeta) isChecked() bool {
	return fm.checked
}

func (fm *FileMeta) isChanged() bool {
	return fm.changed
}

func (fm *FileMeta) isDeleted() bool {
	l := len(fm.Checksum)
	if l == 0 {
		return false
	}
	return fm.Checksum[l-1] == deleted
}

func (fm *FileMeta) markDeleted() {
	// fmt.Printf("deleting %s\n", fm.Name)
	if !fm.isDeleted() {
		fm.Checksum = append(fm.Checksum, deleted)
		fm.Size = 0
		fm.Time = time.Now().UTC()
		fm.changed = true
	}
}

func (fm *FileMeta) markChecked() {
	fm.checked = true
}

func (fm *FileMeta) markChanged() {
	fm.changed = true
}

func (fm *FileMeta) update(path string, info os.FileInfo) error {
	fm.checked = true
	// fmt.Printf("checking %s\n", fm.Name)

	if !fm.isDeleted() {
		if info.Size() == fm.Size && info.ModTime().UTC() == fm.Time {
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

// .8888b oo dP          888888ba  dP
// 88   "    88          88    `8b 88
// 88aaa  dP 88 .d8888b. 88     88 88d888b.
// 88     88 88 88ooood8 88     88 88'  `88
// 88     88 88 88.  ... 88    .8P 88.  .88
// dP     dP dP `88888P' 8888888P  88Y8888'

type fileDb struct {
	dbDir      string
	absBaseDir string

	baseDir string // this is simply kept for saving purposes
	files   map[string]*FileMeta
}

type jsonStruct struct {
	V1 *v1Struct `json:"v1"`
}

type v1Struct struct {
	BaseDir string      `json:"baseDir"`
	Files   []*FileMeta `json:"files"`
}

// GetDbDir ...
func (db *fileDb) GetDbDir() string {
	return db.dbDir
}

// GetBaseDir ...
func (db *fileDb) GetBaseDir() string {
	return db.absBaseDir
}

func (db *fileDb) GetFile(path string) *FileMeta {
	if file, ok := db.files[path]; ok {
		return file
	}
	return nil
}

// Save ...
func (db *fileDb) Save() error {
	// prepare simplified json structure
	fileSlice := make([]*FileMeta, 0, len(db.files))
	for _, file := range db.files {
		fileSlice = append(fileSlice, file)
	}
	sort.Slice(fileSlice, func(i, j int) bool {
		return fileSlice[i].Name < fileSlice[j].Name
	})
	rawDb := &jsonStruct{
		V1: &v1Struct{
			BaseDir: db.baseDir,
			Files:   fileSlice,
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

		encoder.Encode(rawDb)
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

// Update ...
func (db *fileDb) Update() error {
	dir := db.GetBaseDir()

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("base directory '%s' does not exist", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("base directory '%s' is not a directory", dir)
	}

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
			// TODO: just checking if the beginning of the path matches expectation
			fmt.Printf("Root mismatch '%s' != '%s'\n", dir, root)
		}

		relPath := path[len(dir)+1:]

		file, ok := db.files[relPath]
		if !ok {
			// fmt.Printf("new file %s\n", relPath)
			file = &FileMeta{
				Name: relPath,
			}
			db.files[relPath] = file
			// } else {
			// 	fmt.Printf("existing file %s\n", relPath)
		}

		err = (*file).update(path, info)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, file := range db.files {
		if !file.isChecked() {
			file.markDeleted()
		}
	}

	return nil
}

// Import ...
func (db *fileDb) Import(remote FileDb) error {
	return fmt.Errorf("not implemented")
	// files := make(map[string]*FileMeta)
	// for _, file := range db.Files {
	// 	files[file.Name] = file
	// }
	//
	// dir := db.GetBaseDir()
	//
	// err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
	// 	if err != nil {
	// 		fmt.Printf("%s: error getting file info: %s\n", path, err)
	// 		return nil
	// 	}
	// 	if info.IsDir() {
	// 		if info.Name() == defaultDbDir {
	// 			return filepath.SkipDir
	// 		}
	// 		return nil
	// 	}
	//
	// 	root := path[:len(dir)]
	// 	if dir != root {
	// 		// TODO: just checking if the beginning of the path matches expectation
	// 		fmt.Printf("Root mismatch '%s' != '%s'\n", dir, root)
	// 	}
	//
	// 	relPath := path[len(dir)+1:]
	//
	// 	meta, ok := files[relPath]
	// 	if !ok {
	// 		meta = &FileMeta{
	// 			Name: relPath,
	// 		}
	// 		files[relPath] = meta
	// 	}
	//
	// 	err = meta.update(path, info)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	return nil
	// })
	//
	// if err != nil {
	// 	return err
	// }
	//
	// fileSlice := make([]*FileMeta, 0, len(files))
	// changed := false
	// for _, file := range files {
	// 	if !file.isChecked() {
	// 		if !file.isDeleted() {
	// 			file.markDeleted()
	// 			changed = true
	// 		}
	// 	} else {
	// 		changed = changed || file.isChanged()
	// 	}
	// 	fileSlice = append(fileSlice, file)
	// }
	//
	// if changed {
	// 	sort.Slice(fileSlice, func(i, j int) bool {
	// 		return fileSlice[i].Name < fileSlice[j].Name
	// 	})
	// 	db.Files = fileSlice
	//
	// 	db.changed = changed
	// }
	//
	// return nil
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
func InitDbDir(dbDir, baseDir string) (FileDb, error) {
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

// LoadFileDb ...
func LoadFileDb(dbDir string) (FileDb, error) {
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

	rawDb := &jsonStruct{}

	if err := decoder.Decode(&rawDb); err != nil {
		return nil, err
	}

	// ensure there is nothing after the first json object
	dummy := &jsonStruct{}
	if err = decoder.Decode(&dummy); err != io.EOF {
		return nil, fmt.Errorf("unexpected contents at the end of config file")
	}

	if rawDb.V1 == nil {
		return nil, fmt.Errorf("config file is empty")
	}

	db := &fileDb{
		dbDir:   dbDir,
		baseDir: rawDb.V1.BaseDir,
		files:   make(map[string]*FileMeta),
	}

	for _, file := range rawDb.V1.Files {
		db.files[file.Name] = file
	}

	if filepath.IsAbs(db.baseDir) {
		db.absBaseDir = db.baseDir
	} else {
		db.absBaseDir, err = cleanPath(filepath.Join(dbDir, db.baseDir))
		if err != nil {
			return nil, err
		}
	}

	// fmt.Printf("loaded %#v\n", db)

	return db, nil
}
