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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// import "crypto/sha256"

// import "encoding/base64"

const defaultDbDir = ".aether"
const filesFilename = "files.json"
const newFilesFilename = "files.json.tmp"

// FileMeta ...
type FileMeta struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Time     time.Time `json:"time"`
	Checksum []string  `json:"checksum"`
}

// comment
const (
	Equal = iota
	LeftChanged
	RightChanged
	Conflict
)

// Compare ...
func Compare(left, right FileMeta) {
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

	BaseDir string     `json:"baseDir"`
	Files   []FileMeta `json:"files"`
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

	return nil
}

// Update ...
func (db *fileDb) Update() error {
	return fmt.Errorf("not implemented")
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

// func ensureDir(dir string) string, error {
// 	dir, err := cleanPath(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	info, err := os.Stat(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !info.IsDir() {
// 		return nil, fmt.Errorf("'%s' is not a directory", dir)
// 	}
// 	return dir, nil
// }

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

	if filepath.IsAbs(db.dbDir) {
		db.absBaseDir = db.dbDir
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

// type filterFunc func(info os.FileInfo) (accepted bool, reason string)
//
// func okFilter(info os.FileInfo) (accepted bool, reason string) {
// 	return true, ""
// }
//
// func getFiles(dir string, accept filterFunc) error {
// 	dir, err := filepath.Abs(dir)
// 	if err != nil {
// 		return err
// 	}
// 	dir = filepath.Clean(dir)
// 	info, err := os.Stat(dir)
// 	if err != nil {
// 		return err
// 	}
// 	if !info.IsDir() {
// 		return fmt.Errorf("'%s' is not a directory", dir)
// 	}
//
// 	encoder := json.NewEncoder(os.Stdout)
// 	if encoder == nil {
// 		return fmt.Errorf("failed to create json encoder")
// 	}
// 	encoder.Encode('[')
//
// 	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			fmt.Printf("%s: error getting file info: %s\n", path, err)
// 			return nil
// 		}
// 		if info.IsDir() {
// 			return nil
// 		}
//
// 		if accept != nil {
// 			if ok, reason := accept(info); !ok {
// 				fmt.Printf("%s: skipping: %s\n", path, reason)
// 				return nil
// 			}
// 		}
//
// 		root := path[:len(dir)]
// 		if dir != root {
// 			// TODO: just checking if the beginning of the path matches expectation
// 			fmt.Printf("Root mismatch '%s' != '%s'\n", dir, root)
// 		}
//
// 		file, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer file.Close()
//
// 		hash := sha256.New()
// 		if _, err := io.Copy(hash, file); err != nil {
// 			return err
// 		}
//
// 		path = path[len(dir)+1:]
//
// 		meta := &FileMeta{
// 			Name:     path,
// 			Size:     info.Size(),
// 			Time:     info.ModTime().UTC(),
// 			Checksum: []string{base64.StdEncoding.EncodeToString(hash.Sum(nil))},
// 		}
//
// 		encoder.Encode(meta)
// 		// // buf, err := json.MarshalIndent(c, "", "  ")
// 		// buf, err := json.Marshal(meta)
// 		// if err != nil {
// 		// 	return err
// 		// }
// 		//
// 		// fmt.Printf("%s\n", string(buf))
//
// 		return nil
// 	})
//
// 	if err != nil {
// 		return err
// 	}
//
// 	encoder.Encode(']')
//
// 	return nil
// }

// // Scan ...
// func Scan(dir string) error {
//
// 	db := LoadFileDB(dir)
// }
